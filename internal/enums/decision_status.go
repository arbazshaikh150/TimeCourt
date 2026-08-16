package enums

type DecisionOutcome string

const (
	DecisionOutcomeDetermined    DecisionOutcome = "DETERMINED"
	DecisionOutcomeIndeterminate DecisionOutcome = "INDETERMINATE"
	DecisionOutcomeNotApplicable DecisionOutcome = "NOT_APPLICABLE"
	DecisionOutcomeError         DecisionOutcome = "ERROR"
	DecisionOutcomeSuperseded    DecisionOutcome = "SUPERSEDED"
	DecisionOutcomeUnderReview   DecisionOutcome = "UNDER_REVIEW"
)
