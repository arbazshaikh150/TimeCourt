package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

/*
Finding the latest version (since version are increased atomically ) per tenant
and checking the resolution based on the resolution version and then passed to interpretor
*/
const findApplicableRuleQuery = `
	WITH applicable_rule AS (
		SELECT
			rule_id,
			tenant_id,
			rule_key,
			rule_version,
			knowledge_time,
			effective_start_time,
			effective_end_time,
			resolution_version,
			interpretor_version,
			source
		FROM rules
		WHERE rule_key = $1
			AND effective_period @> $2::timestamptz
			AND knowledge_time <= $3
		ORDER BY knowledge_time DESC
		LIMIT 1
	),
	latest_facts AS (
		SELECT DISTINCT ON (
			fi.tenant_id,
			fi.fact_key,
			fi.subject_id
		)
			fi.fact_information_id,
			fi.tenant_id,
			fi.fact_key,
			fi.fact_version,
			fi.subject_id,
			fi.fact_effective_start_time,
			fi.fact_effective_end_time,
			fi.knowledge_time,
			fi.fact_value,
			fi.authority,
			fi.confidence,
			fi.source
		FROM fact_information fi
		WHERE fi.subject_id = $4
			AND fi.effective_period @> $2::timestamptz
			AND fi.knowledge_time <= $3
		ORDER BY
			fi.tenant_id,
			fi.fact_key,
			fi.subject_id,
			fi.fact_version DESC,
			fi.knowledge_time DESC
	)
	SELECT
		r.rule_id,
		r.tenant_id,
		r.rule_key,
		r.rule_version,
		r.knowledge_time,
		r.effective_start_time,
		r.effective_end_time,
		r.resolution_version,
		r.interpretor_version,
		r.source,

		rd.fact_key,
		rd.fact_required_value,

		fi.fact_information_id,
		fi.tenant_id,
		fi.fact_key,
		fi.fact_version,
		fi.subject_id,
		fi.fact_effective_start_time,
		fi.fact_effective_end_time,
		fi.knowledge_time,
		fi.fact_value,
		fi.authority,
		fi.confidence,
		fi.source

	FROM applicable_rule r
	INNER JOIN rule_details rd
		ON rd.rule_id = r.rule_id
	LEFT JOIN latest_facts fi
		ON fi.fact_key = rd.fact_key
		AND fi.subject_id = $4
	ORDER BY rd.fact_key
`

// Find returns the most recently known rule applicable at timeWhereToCheck,
// together with the facts for subjectID known at timeWhenToCheck.
func (r *PgxRuleTableRepository) Find(
	ctx context.Context,
	ruleKey, subjectID string,
	timeWhenToCheck, timeWhereToCheck time.Time,
) (*dto.RuleFetchDetails, error) {
	rows, err := r.db.Query(ctx, findApplicableRuleQuery, ruleKey, timeWhereToCheck, timeWhenToCheck, subjectID)
	if err != nil {
		return nil, fmt.Errorf("find applicable rule: %w", err)
	}
	defer rows.Close()

	result, err := collectRuleFetchDetails(rows)
	if err != nil {
		return nil, fmt.Errorf("collect applicable rule: %w", err)
	}
	return result, nil
}

func collectRuleFetchDetails(rows pgx.Rows) (*dto.RuleFetchDetails, error) {
	var result *dto.RuleFetchDetails
	factDetailIndexes := make(map[string]int)

	for rows.Next() {
		row, err := scanRuleSearchRow(rows)
		if err != nil {
			return nil, err
		}
		if result == nil {
			result = row.ruleFetchDetails()
		}
		addRuleFact(result, factDetailIndexes, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, pgx.ErrNoRows
	}
	return result, nil
}

func addRuleFact(result *dto.RuleFetchDetails, factDetailIndexes map[string]int, row ruleSearchRow) {
	index, exists := factDetailIndexes[row.factKey]
	if !exists {
		index = len(result.FactDetails)
		factDetailIndexes[row.factKey] = index
		result.FactDetails = append(result.FactDetails, dto.RuleFactDetails{
			FactKey:           row.factKey,
			RequiredFactValue: row.requiredFactValue,
			Facts:             make([]dto.FactDetails, 0),
		})
	}
	if row.factInformationID == nil {
		return
	}
	result.FactDetails[index].Facts = append(result.FactDetails[index].Facts, row.factDetails())
}

type ruleSearchRow struct {
	ruleID             uuid.UUID
	tenantID           string
	ruleKey            string
	ruleVersion        int64
	knowledgeTime      time.Time
	effectiveStartTime time.Time
	effectiveEndTime   time.Time
	resolutionVersion  int64
	interpretorVersion int64
	source             string

	factKey           string
	requiredFactValue string

	factInformationID      *uuid.UUID
	factTenantID           *string
	actualFactKey          *string
	factVersion            *int64
	factSubjectID          *string
	factEffectiveStartTime *time.Time
	factEffectiveEndTime   *time.Time
	factKnowledgeTime      *time.Time
	factValue              *string
	factAuthority          *string
	factConfidence         *string
	factSource             *string
}

func scanRuleSearchRow(rows pgx.Rows) (ruleSearchRow, error) {
	var row ruleSearchRow
	err := rows.Scan(
		&row.ruleID, &row.tenantID, &row.ruleKey, &row.ruleVersion,
		&row.knowledgeTime, &row.effectiveStartTime, &row.effectiveEndTime,
		&row.resolutionVersion, &row.interpretorVersion, &row.source,
		&row.factKey, &row.requiredFactValue,
		&row.factInformationID, &row.factTenantID, &row.actualFactKey,
		&row.factVersion, &row.factSubjectID, &row.factEffectiveStartTime,
		&row.factEffectiveEndTime, &row.factKnowledgeTime, &row.factValue,
		&row.factAuthority, &row.factConfidence, &row.factSource,
	)
	return row, err
}

func (r ruleSearchRow) ruleFetchDetails() *dto.RuleFetchDetails {
	return &dto.RuleFetchDetails{
		RuleID:             r.ruleID,
		TenantID:           r.tenantID,
		RuleKey:            r.ruleKey,
		RuleVersion:        r.ruleVersion,
		KnowledgeTime:      r.knowledgeTime,
		EffectiveStartTime: r.effectiveStartTime,
		EffectiveEndTime:   r.effectiveEndTime,
		ResolutionVersion:  r.resolutionVersion,
		InterpretorVersion: r.interpretorVersion,
		Source:             r.source,
		FactDetails:        make([]dto.RuleFactDetails, 0),
	}
}

func (r ruleSearchRow) factDetails() dto.FactDetails {
	return dto.FactDetails{
		FactInformationID:      *r.factInformationID,
		TenantID:               *r.factTenantID,
		FactKey:                *r.actualFactKey,
		FactVersion:            *r.factVersion,
		SubjectID:              *r.factSubjectID,
		FactEffectiveStartTime: *r.factEffectiveStartTime,
		FactEffectiveEndTime:   *r.factEffectiveEndTime,
		KnowledgeTime:          *r.factKnowledgeTime,
		FactValue:              *r.factValue,
		Authority:              *r.factAuthority,
		Confidence:             *r.factConfidence,
		Source:                 *r.factSource,
	}
}
