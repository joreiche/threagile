package model

import "github.com/threagile/threagile/pkg/security/types"

type RiskRule struct {
	Category      func() types.RiskCategory
	SupportedTags func() []string
	GenerateRisks func(input *ParsedModel) []types.Risk
}
