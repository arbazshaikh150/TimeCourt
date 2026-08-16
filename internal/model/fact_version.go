package model

type FactVersion struct {
	FactKey       string `gorm:"type:text;not null;primaryKey" json:"fact_key"`
	SubjectID     string `gorm:"type:text;not null;primaryKey" json:"subject"`
	LatestVersion int64  `gorm:"not null" json:"version"`
}

func (FactVersion) TableName() string {
	return "fact_versions"
}