package dto

import "encoding/json"

type CreateResolutionRequest struct {
	RuleKey          string          `json:"rule_key"`
	TenantID         string          `json:"tenant_id"`
	ResolutionPolicy json.RawMessage `json:"resolution_policy"`
	Source           string          `json:"source"`
	Version          int64           `json:"version"`
}
