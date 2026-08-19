package service

import (
	"context"
	"fmt"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/repository"
	"github.com/google/uuid"
)

// Should have a dependency on the rule repository
type RuleService struct {
	repository        repository.RuleTableRepository
	idempotentService IdempotentService
}

func NewRuleService(
	repository repository.RuleTableRepository,
	idempotentService IdempotentService,
) *RuleService {
	return &RuleService{
		repository:        repository,
		idempotentService: idempotentService,
	}
}

func (r *RuleService) Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleTable, error) {
	fmt.Println("Fetching the rule based on the ruleid")
	return r.repository.Get(ctx, ruleID)
}

// TODO : Resolution should be taken care in next part ( making it as a can be null values)
// Another http request can change the rule's resolution version ( means create a new rule version )
// Might look overwhelming ( do not have any numerical proof , but thinking this as a optimization only)
// Will derive the resolution , if in same http call it is suitable then i will change this function

func (r *RuleService) Create(ctx context.Context, req *dto.CreateRuleRequest, idempotentKey uuid.UUID) error {

	idempotentResult, err := r.idempotentService.Lock(
		ctx,
		idempotentKey,
		req.Source,
		req.TenantID,
	)

	if !idempotentResult.Acquired {
		err := fmt.Errorf(
			"idempotency lock was not acquired for key %s (source=%s, status=%s)",
			idempotentKey,
			idempotentResult.Source,
			idempotentResult.Status,
		)

		fmt.Println(err)

		return err
	}

	if err != nil {
		fmt.Printf(
			"failed to acquire idempotency lock for key %s: %v\n",
			idempotentKey,
			err,
		)

		return fmt.Errorf(
			"acquire idempotency lock for key %s: %w",
			idempotentKey,
			err,
		)
	}
	//
	// 1. Generating the Rule ID.
	//

	ruleID := uuid.New()

	//
	// 2. Convert request -> RuleTable.
	//
	// RuleVersion is NOT set here.
	// Repository will generate it atomically.
	//

	rule := &model.RuleTable{
		RuleID:             ruleID,
		TenantID:           req.TenantID,
		RuleKey:            req.RuleKey,
		KnowledgeTime:      req.KnowledgeTime,
		EffectiveStartTime: req.EffectiveStartTime,
		EffectiveEndTime:   req.EffectiveEndTime,
		ResolutionVersion:  req.ResolutionVersion,
		InterpretorVersion: req.InterpretorVersion,
		Source:             req.Source,
	}

	//
	// 3. Converting FactsKeys -> []RuleDetail.
	//

	ruleDetails := make([]model.RuleDetail, 0, len(req.FactsKeys))

	for _, fact := range req.FactsKeys {

		ruleDetails = append(
			ruleDetails,
			model.RuleDetail{
				RuleID:            ruleID,
				FactKey:           fact.FactKey,
				FactRequiredValue: fact.FactRequiredValue,
			},
		)
	}

	//
	// 4. Repository handles the complete transaction.
	//
	// It will:
	//   - increment rule version
	//   - insert rules
	//   - insert rule details
	//   - update idempotency record
	//   - commit
	//
	err = r.repository.Create(
		ctx,
		rule,
		ruleDetails,
		idempotentKey,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create rule %s: %w",
			ruleID,
			err,
		)
	}

	return nil
}
