package wrong_trust_boundary_content

import (
	"github.com/threagile/threagile/pkg/security/types"
)

func Rule() types.RiskRule {
	return types.RiskRule{
		Category:      Category,
		SupportedTags: SupportedTags,
		GenerateRisks: GenerateRisks,
	}
}

func Category() types.RiskCategory {
	return types.RiskCategory{
		Id:    "wrong-trust-boundary-content",
		Title: "Wrong Trust Boundary Content",
		Description: "When a trust boundary of type " + types.NetworkPolicyNamespaceIsolation.String() + " contains " +
			"non-container assets it is likely to be a model failure.",
		Impact:                     "If this potential model error is not fixed, some risks might not be visible.",
		ASVS:                       "V1 - Architecture, Design and Threat Modeling Requirements",
		CheatSheet:                 "https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html",
		Action:                     "Model Consistency",
		Mitigation:                 "Try to model the correct types of trust boundaries and data assets.",
		Check:                      "Are recommendations from the linked cheat sheet and referenced ASVS chapter applied?",
		Function:                   types.Architecture,
		STRIDE:                     types.ElevationOfPrivilege,
		DetectionLogic:             "Trust boundaries which should only contain containers, but have different assets inside.",
		RiskAssessment:             types.LowSeverity.String(),
		FalsePositives:             "Usually no false positives as this looks like an incomplete model.",
		ModelFailurePossibleReason: true,
		CWE:                        1008,
	}
}

func SupportedTags() []string {
	return []string{}
}

func GenerateRisks(input *types.ParsedModel) []types.Risk {
	risks := make([]types.Risk, 0)
	for _, trustBoundary := range input.TrustBoundaries {
		if trustBoundary.Type == types.NetworkPolicyNamespaceIsolation {
			for _, techAssetID := range trustBoundary.TechnicalAssetsInside {
				techAsset := input.TechnicalAssets[techAssetID]
				if techAsset.Machine != types.Container && techAsset.Machine != types.Serverless {
					risks = append(risks, createRisk(techAsset))
				}
			}
		}
	}
	return risks
}

func createRisk(technicalAsset types.TechnicalAsset) types.Risk {
	title := "<b>Wrong Trust Boundary Content</b> (non-container asset inside container trust boundary) at <b>" + technicalAsset.Title + "</b>"
	risk := types.Risk{
		CategoryId:                   Category().Id,
		Severity:                     types.CalculateSeverity(types.Unlikely, types.LowImpact),
		ExploitationLikelihood:       types.Unlikely,
		ExploitationImpact:           types.LowImpact,
		Title:                        title,
		MostRelevantTechnicalAssetId: technicalAsset.Id,
		DataBreachProbability:        types.Improbable,
		DataBreachTechnicalAssetIDs:  []string{technicalAsset.Id},
	}
	risk.SyntheticId = risk.CategoryId + "@" + technicalAsset.Id
	return risk
}
