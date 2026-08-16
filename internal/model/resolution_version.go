package model

type ResolutionVersion struct {
	RuleKey                 string `gorm:"type:text;primaryKey" json:"rule_key"`
	ResolutionLatestVersion int64  `gorm:"not null" json:"resolution_latest_version"`
}

func (ResolutionVersion) TableName() string {
	return "resolution_versions"
}