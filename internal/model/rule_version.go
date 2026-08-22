package model

type RuleVersion struct {
	RuleKey       string `gorm:"type:text;primaryKey" json:"rule_key"`
	RuleVersionID int64  `gorm:"not null" json:"rule_version_id"`
}

func (RuleVersion) TableName() string {
	return "rule_versions"
}
