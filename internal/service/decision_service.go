package service

import (
	"context"
	"fmt"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
)

// RuleFinder is the part of RuleService required to evaluate a decision.
type RuleFinder interface {
	Find(ctx context.Context, ruleKey, subjectID string, timeWhenToCheck, timeWhereToCheck time.Time) (*dto.RuleFetchDetails, error)
}

// ResolutionFinder is used only when more than two facts are available for a
type ResolutionFinder interface {
	FindResolution(ctx context.Context, ruleKey string, resolutionVersion int64) ([]*model.ResolutionTable, error)
}

// InterpretorService determines the decision outcome from the selected rule,
type InterpretorService interface {
	Interpret(ctx context.Context, rule *dto.RuleFetchDetails, resolutions []*model.ResolutionTable) (*InterpretorResult, error)
}

// DecisionWriter persists the models created by the interpreter.
// repository.DecisionRecordRepository satisfies this interface.
type DecisionWriter interface {
	Create(ctx context.Context, record *model.DecisionRecord, details *model.DecisionDetails) error
}

// DecisionResult contains the complete data used to derive an outcome, so a
// caller can persist or explain the decision without repeating time-sensitive
// lookups.
type DecisionResult struct {
	Rule           *dto.RuleFetchDetails    `json:"rule"`
	Resolutions    []*model.ResolutionTable `json:"resolutions,omitempty"`
	Outcome        enums.DecisionOutcome    `json:"outcome"`
	Interpretation *InterpretorResult       `json:"interpretation"`
}

// DecisionService coordinates rule lookup, ambiguity resolution, and outcome
// interpretation.
type DecisionService struct {
	ruleService        RuleFinder
	resolutionService  ResolutionFinder
	interpretorService InterpretorService
	decisionWriter     DecisionWriter
}

func NewDecisionService(
	ruleService RuleFinder,
	resolutionService ResolutionFinder,
	interpretorService InterpretorService,
	decisionWriters ...DecisionWriter,
) *DecisionService {
	var decisionWriter DecisionWriter
	if len(decisionWriters) > 0 {
		decisionWriter = decisionWriters[0]
	}
	return &DecisionService{
		ruleService:        ruleService,
		resolutionService:  resolutionService,
		interpretorService: interpretorService,
		decisionWriter:     decisionWriter,
	}
}

// Decide evaluates ruleKey for subjectID at the supplied knowledge and
// effective times. A resolution policy is fetched only when at least one
// required fact has more than two candidate facts.
func (d *DecisionService) Decide(
	ctx context.Context,
	ruleKey string,
	subjectID string,
	timeWhenToCheck time.Time,
	timeWhereToCheck time.Time,
) (*DecisionResult, error) {
	if d == nil || d.ruleService == nil {
		return nil, fmt.Errorf("decision service rule service is not configured")
	}
	if d.interpretorService == nil {
		return nil, fmt.Errorf("decision service interpretor service is not configured")
	}

	rule, err := d.ruleService.Find(ctx, ruleKey, subjectID, timeWhenToCheck, timeWhereToCheck)
	if err != nil {
		return nil, fmt.Errorf("find rule details for rule %q and subject %q: %w", ruleKey, subjectID, err)
	}

	var resolutions []*model.ResolutionTable
	if hasAmbiguousFacts(rule.FactDetails) {
		if d.resolutionService == nil {
			return nil, fmt.Errorf("decision service resolution service is not configured")
		}

		resolutions, err = d.resolutionService.FindResolution(ctx, rule.RuleKey, rule.ResolutionVersion)
		if err != nil {
			return nil, fmt.Errorf("find resolution for rule %q version %d: %w", rule.RuleKey, rule.ResolutionVersion, err)
		}
	}

	interpretation, err := d.interpretorService.Interpret(ctx, rule, resolutions)
	if err != nil {
		return nil, fmt.Errorf("interpret decision for rule %q and subject %q: %w", ruleKey, subjectID, err)
	}
	interpretation.DecisionRecord.TimeWhenToCheck = timeWhenToCheck
	if d.decisionWriter != nil {
		if err := d.decisionWriter.Create(ctx, &interpretation.DecisionRecord, &interpretation.DecisionDetails); err != nil {
			return nil, fmt.Errorf("store decision for rule %q and subject %q: %w", ruleKey, subjectID, err)
		}
	}

	return &DecisionResult{
		Rule:           rule,
		Resolutions:    resolutions,
		Outcome:        interpretation.DecisionDetails.Status,
		Interpretation: interpretation,
	}, nil
}

func hasAmbiguousFacts(factDetails []dto.RuleFactDetails) bool {
	for _, factDetail := range factDetails {
		if len(factDetail.Facts) > 2 {
			return true
		}
	}
	return false
}
