package dto

import "github.com/arbazshaikh150/TimeCourt/internal/enums"

type IdempotentResult struct {
	Source   string
	Acquired bool
	Status   enums.IdempotentStatus
}
