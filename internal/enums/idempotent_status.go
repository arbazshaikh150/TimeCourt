package enums

type IdempotentStatus string

const (
	IdempotentSuccess    IdempotentStatus = "SUCCESS"
	IdempotentFailure    IdempotentStatus = "FAILED"
	IdempotentProcessing IdempotentStatus = "PROCESSING"
)
