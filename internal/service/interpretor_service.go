package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Interpretor checks the facts for a rule and creates the decision data.
type Interpretor struct{}

// NewInterpretorService creates a new rule interpreter.
func NewInterpretorService() *Interpretor {
	return &Interpretor{}
}

// InterpretorResult holds the records that can be saved after a rule is checked.
type InterpretorResult struct {
	DecisionRecord  model.DecisionRecord  `json:"decision_record"`
	DecisionDetails model.DecisionDetails `json:"decision_details"`
	Message         string                `json:"message"`
}

// Interpret checks all facts against the rule and returns the decision records.
func (i *Interpretor) Interpret(
	ctx context.Context,
	rule *dto.RuleFetchDetails,
	resolutions []*model.ResolutionTable,
) (*InterpretorResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, fmt.Errorf("rule details are required")
	}

	result := newInterpretorResult(rule)
	if len(resolutions) > 1 {
		result.DecisionDetails.Status = enums.DecisionOutcomeIndeterminate
		result.Message = "conflicting facts have more than one applicable resolution policy"
		populateDecisionData(&result, rule.FactDetails, rule.FactDetails)
		return &result, nil
	}

	var policy resolutionPriority
	if len(resolutions) == 1 {
		result.DecisionDetails.ResolutionPolicyVersion = resolutions[0].Version
		result.DecisionDetails.ResolutionPolicyRef = resolutions[0].ResolutionID
		policy = parseResolutionPriority(resolutions[0].ResolutionPolicy)
	}

	failedFacts := make([]dto.RuleFactDetails, 0)
	for _, required := range rule.FactDetails {
		matched, determined := evaluateRequiredFact(required, policy)
		if !determined {
			result.DecisionDetails.Status = enums.DecisionOutcomeIndeterminate
			result.Message = "one or more conflicting facts cannot be resolved by the resolution policy"
			failedFacts = append(failedFacts, required)
			continue
		}
		if !matched {
			failedFacts = append(failedFacts, required)
		}
	}
	populateDecisionData(&result, rule.FactDetails, failedFacts)

	if result.DecisionDetails.Status == enums.DecisionOutcomeIndeterminate {
		return &result, nil
	}
	if len(failedFacts) > 0 {
		result.DecisionDetails.Status = enums.DecisionOutcomeNotApplicable
		result.Message = "one or more facts do not satisfy the rule requirements"
		return &result, nil
	}

	result.DecisionDetails.Status = enums.DecisionOutcomeDetermined
	result.Message = "all required facts satisfy the rule"
	return &result, nil
}

// newInterpretorResult creates linked decision records with their default values.
func newInterpretorResult(rule *dto.RuleFetchDetails) InterpretorResult {
	decisionID := uuid.New()
	return InterpretorResult{
		DecisionRecord: model.DecisionRecord{
			DecisionID:       decisionID,
			TenantID:         rule.TenantID,
			TimeOfCompletion: time.Now().UTC(),
			TimeWhenToCheck:  rule.KnowledgeTime,
			Version:          1,
		},
		DecisionDetails: model.DecisionDetails{
			DecisionID:              decisionID,
			SubjectID:               subjectID(rule.FactDetails),
			RuleKey:                 rule.RuleKey,
			RuleUsedVersion:         rule.RuleVersion,
			ResolutionPolicyVersion: 0,
			ResolutionPolicyRef:     uuid.Nil,
			Status:                  enums.DecisionOutcomeDetermined,
		},
	}
}

// subjectID gets the subject from the first available fact.
func subjectID(facts []dto.RuleFactDetails) string {
	for _, fact := range facts {
		if len(fact.Facts) > 0 {
			return fact.Facts[0].SubjectID
		}
	}
	return ""
}

// evaluateRequiredFact selects one fact and checks whether it meets the rule.
// The second result is false when a conflicting fact cannot be selected.
func evaluateRequiredFact(required dto.RuleFactDetails, policy resolutionPriority) (matched, determined bool) {
	if len(required.Facts) == 0 {
		return false, true
	}

	selected, ok := selectFact(required.Facts, policy)
	if !ok {
		return false, false
	}
	return matchesRequiredValue(required.RequiredFactValue, selected.FactValue), true
}

// selectFact picks one fact using the resolution policy when values conflict.
func selectFact(candidates []dto.FactDetails, policy resolutionPriority) (dto.FactDetails, bool) {
	if len(candidates) == 1 || allSameValue(candidates) {
		return candidates[0], true
	}
	if !policy.configured() {
		return dto.FactDetails{}, false
	}

	best, bestScore := candidates[0], policy.score(candidates[0])
	tied := false
	for _, candidate := range candidates[1:] {
		score := policy.score(candidate)
		if score < bestScore {
			best, bestScore, tied = candidate, score, false
		} else if score == bestScore {
			tied = true
		}
	}
	return best, !tied
}

// allSameValue reports whether all candidate facts have the same value.
func allSameValue(candidates []dto.FactDetails) bool {
	for _, candidate := range candidates[1:] {
		if !strings.EqualFold(strings.TrimSpace(candidate.FactValue), strings.TrimSpace(candidates[0].FactValue)) {
			return false
		}
	}
	return true
}

type resolutionPriority struct {
	Confidence []string
	Source     []string
	Authority  []string
}

// configured reports whether the policy has at least one priority list.
func (p resolutionPriority) configured() bool {
	return len(p.Confidence)+len(p.Source)+len(p.Authority) > 0
}

// score gives a lower number to facts preferred by the policy.
func (p resolutionPriority) score(fact dto.FactDetails) int {
	return priorityRank(p.Confidence, fact.Confidence)*1_000_000 +
		priorityRank(p.Source, fact.Source)*1_000 +
		priorityRank(p.Authority, fact.Authority)
}

// priorityRank returns the position of a value in a priority list.
func priorityRank(priorities []string, value string) int {
	for index, priority := range priorities {
		if strings.EqualFold(strings.TrimSpace(priority), strings.TrimSpace(value)) {
			return index
		}
	}
	return len(priorities) + 1
}

// parseResolutionPriority reads priority lists from a resolution policy.
func parseResolutionPriority(raw json.RawMessage) resolutionPriority {
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil {
		return resolutionPriority{}
	}
	return resolutionPriority{
		Confidence: findPriority(value, "confidence"),
		Source:     findPriority(value, "source"),
		Authority:  findPriority(value, "authority"),
	}
}

// findPriority finds one priority list, including lists inside a nested object.
func findPriority(value map[string]json.RawMessage, name string) []string {
	for key, raw := range value {
		if strings.EqualFold(key, name) {
			var priorities []string
			if json.Unmarshal(raw, &priorities) == nil {
				return priorities
			}
		}
	}
	return nil
}

var comparisonPattern = regexp.MustCompile(`(?i)^\s*(<=|>=|<|>|=|==|under|below|over|above)\s*([+-]?[0-9][0-9,]*(?:\.[0-9]+)?)\s*$`)

// matchesRequiredValue compares a fact value with a rule value or number check.
func matchesRequiredValue(required, actual string) bool {
	match := comparisonPattern.FindStringSubmatch(required)
	if len(match) == 0 {
		return strings.EqualFold(strings.TrimSpace(required), strings.TrimSpace(actual))
	}
	want, wantOK := parseNumber(match[2])
	got, gotOK := parseNumber(actual)
	if !wantOK || !gotOK {
		return false
	}
	switch strings.ToLower(match[1]) {
	case "<", "under", "below":
		return got < want
	case "<=":
		return got <= want
	case ">", "over", "above":
		return got > want
	case ">=":
		return got >= want
	default:
		return got == want
	}
}

// parseNumber converts a formatted number into a value that can be compared.
func parseNumber(value string) (float64, bool) {
	value = strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	value = strings.TrimLeft(value, "$₹€£")
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil
}

// populateDecisionData stores the required, available, and failed facts as JSON.
func populateDecisionData(result *InterpretorResult, facts, failedFacts []dto.RuleFactDetails) {
	type requiredFact struct {
		FactKey       string `json:"fact_key"`
		RequiredValue string `json:"required_value"`
	}
	requiredFacts := make([]requiredFact, 0, len(facts))
	presentFacts := make([]dto.FactDetails, 0)
	for _, fact := range facts {
		requiredFacts = append(requiredFacts, requiredFact{
			FactKey:       fact.FactKey,
			RequiredValue: fact.RequiredFactValue,
		})
		presentFacts = append(presentFacts, fact.Facts...)
	}
	failed := make([]requiredFact, 0, len(failedFacts))
	for _, fact := range failedFacts {
		failed = append(failed, requiredFact{
			FactKey:       fact.FactKey,
			RequiredValue: fact.RequiredFactValue,
		})
	}

	required, _ := json.Marshal(requiredFacts)
	present, _ := json.Marshal(presentFacts)
	failedJSON, _ := json.Marshal(failed)
	result.DecisionDetails.FactRequiredData = datatypes.JSON(required)
	result.DecisionDetails.FactPresentData = datatypes.JSON(present)
	result.DecisionDetails.FailedFacts = datatypes.JSON(failedJSON)
}
