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

// TODO : Cleanup is pending
// Interpretor evaluates the facts selected for a rule. Persisting the returned
// DecisionRecord and DecisionDetails is deliberately kept with the caller's
// decision repository, so evaluation remains deterministic and testable.
type Interpretor struct{}

func NewInterpretorService() *Interpretor {
	return &Interpretor{}
}

// FactTrace is the explainable, per-fact evaluation trace. It is returned in
// addition to the JSONB fields on DecisionDetails for API consumers.
type FactTrace struct {
	FactKey       string            `json:"fact_key"`
	RequiredValue string            `json:"required_value"`
	SelectedFact  *dto.FactDetails  `json:"selected_fact,omitempty"`
	Candidates    []dto.FactDetails `json:"candidates"`
	Matched       bool              `json:"matched"`
	Reason        string            `json:"reason"`
}

// InterpretorResult is ready to be stored in decision_records and
// decision_details. Message and Trace provide the detailed explanation that is
// not represented by columns in the current schema.
type InterpretorResult struct {
	DecisionRecord  model.DecisionRecord  `json:"decision_record"`
	DecisionDetails model.DecisionDetails `json:"decision_details"`
	Message         string                `json:"message"`
	Trace           []FactTrace           `json:"trace"`
}

// Interpret resolves conflicting facts, evaluates each required fact value,
// and produces database-ready decision models.
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
		result.Trace = unresolvedTraces(rule.FactDetails, result.Message)
		populateTraceData(&result, rule.FactDetails)
		return &result, nil
	}
	fmt.Println("The resolution is :", resolutions)
	var policy resolutionPriority
	if len(resolutions) == 1 {
		result.DecisionDetails.ResolutionPolicyVersion = resolutions[0].Version
		result.DecisionDetails.ResolutionPolicyRef = resolutions[0].ResolutionID
		policy = parseResolutionPriority(resolutions[0].ResolutionPolicy)
	}

	traces := make([]FactTrace, 0, len(rule.FactDetails))
	for _, required := range rule.FactDetails {
		trace, determined := evaluateRequiredFact(required, policy)
		traces = append(traces, trace)
		if !determined {
			result.DecisionDetails.Status = enums.DecisionOutcomeIndeterminate
			result.Message = "one or more conflicting facts cannot be resolved by the resolution policy"
		}
	}
	result.Trace = traces
	populateTraceData(&result, rule.FactDetails)

	if result.DecisionDetails.Status == enums.DecisionOutcomeIndeterminate {
		return &result, nil
	}
	for _, trace := range traces {
		if !trace.Matched {
			result.DecisionDetails.Status = enums.DecisionOutcomeNotApplicable
			result.Message = "one or more facts do not satisfy the rule requirements"
			return &result, nil
		}
	}

	result.DecisionDetails.Status = enums.DecisionOutcomeDetermined
	result.Message = "all required facts satisfy the rule"
	return &result, nil
}

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

func subjectID(facts []dto.RuleFactDetails) string {
	for _, fact := range facts {
		if len(fact.Facts) > 0 {
			return fact.Facts[0].SubjectID
		}
	}
	return ""
}

func evaluateRequiredFact(required dto.RuleFactDetails, policy resolutionPriority) (FactTrace, bool) {
	trace := FactTrace{FactKey: required.FactKey, RequiredValue: required.RequiredFactValue, Candidates: required.Facts}
	if len(required.Facts) == 0 {
		trace.Reason = "no fact is available for this requirement"
		return trace, true
	}

	selected, ok := selectFact(required.Facts, policy)
	if !ok {
		trace.Reason = "multiple conflicting facts exist and no resolution priority selects one"
		return trace, false
	}
	trace.SelectedFact = &selected
	trace.Matched = matchesRequiredValue(required.RequiredFactValue, selected.FactValue)
	if trace.Matched {
		trace.Reason = "fact satisfies the required value"
	} else {
		trace.Reason = "fact value violates the required value"
	}
	return trace, true
}

func selectFact(candidates []dto.FactDetails, policy resolutionPriority) (dto.FactDetails, bool) {
	fmt.Println("The policy is : ", policy)
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

func (p resolutionPriority) configured() bool {
	return len(p.Confidence)+len(p.Source)+len(p.Authority) > 0
}

func (p resolutionPriority) score(fact dto.FactDetails) int {
	return priorityRank(p.Confidence, fact.Confidence)*1_000_000 +
		priorityRank(p.Source, fact.Source)*1_000 +
		priorityRank(p.Authority, fact.Authority)
}

func priorityRank(priorities []string, value string) int {
	for index, priority := range priorities {
		if strings.EqualFold(strings.TrimSpace(priority), strings.TrimSpace(value)) {
			return index
		}
	}
	return len(priorities) + 1
}

// parseResolutionPriority accepts either {"confidence":[...],"source":[...]}
// or the same fields nested under "priority" or "priorities".
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

func findPriority(value map[string]json.RawMessage, name string) []string {
	for key, raw := range value {
		if strings.EqualFold(key, name) {
			var priorities []string
			if json.Unmarshal(raw, &priorities) == nil {
				return priorities
			}
		}
	}
	for key, raw := range value {
		if !strings.EqualFold(key, "priority") && !strings.EqualFold(key, "priorities") {
			continue
		}
		var nested map[string]json.RawMessage
		if json.Unmarshal(raw, &nested) == nil {
			return findPriority(nested, name)
		}
	}
	return nil
}

var comparisonPattern = regexp.MustCompile(`(?i)^\s*(<=|>=|<|>|=|==|under|below|over|above)\s*([+-]?[0-9][0-9,]*(?:\.[0-9]+)?)\s*$`)

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

func parseNumber(value string) (float64, bool) {
	value = strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	value = strings.TrimLeft(value, "$₹€£")
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil
}

func unresolvedTraces(facts []dto.RuleFactDetails, reason string) []FactTrace {
	traces := make([]FactTrace, 0, len(facts))
	for _, fact := range facts {
		traces = append(traces, FactTrace{FactKey: fact.FactKey, RequiredValue: fact.RequiredFactValue, Candidates: fact.Facts, Reason: reason})
	}
	return traces
}

func populateTraceData(result *InterpretorResult, facts []dto.RuleFactDetails) {
	type requiredFact struct {
		FactKey       string `json:"fact_key"`
		RequiredValue string `json:"required_value"`
	}
	requiredFacts := make([]requiredFact, 0, len(facts))
	for _, fact := range facts {
		requiredFacts = append(requiredFacts, requiredFact{
			FactKey:       fact.FactKey,
			RequiredValue: fact.RequiredFactValue,
		})
	}
	required, _ := json.Marshal(requiredFacts)
	present, _ := json.Marshal(result.Trace)
	failed := make([]FactTrace, 0)
	for _, trace := range result.Trace {
		if !trace.Matched || strings.Contains(trace.Reason, "conflicting") || strings.Contains(trace.Reason, "no fact") {
			failed = append(failed, trace)
		}
	}
	failedJSON, _ := json.Marshal(failed)
	result.DecisionDetails.FactRequiredData = datatypes.JSON(required)
	result.DecisionDetails.FactPresentData = datatypes.JSON(present)
	result.DecisionDetails.FailedFacts = datatypes.JSON(failedJSON)
}
