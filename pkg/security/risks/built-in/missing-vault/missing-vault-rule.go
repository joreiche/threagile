package missing_vault

import (
	"github.com/threagile/threagile/pkg/model"
	"github.com/threagile/threagile/pkg/security/types"
)

func Rule() model.CustomRiskRule {
	return model.CustomRiskRule{
		Category:      Category,
		SupportedTags: SupportedTags,
		GenerateRisks: GenerateRisks,
	}
}

func Category() model.RiskCategory {
	return model.RiskCategory{
		Id:    "missing-vault",
		Title: "Missing Vault (Secret Storage)",
		Description: "In order to avoid the risk of secret leakage via config files (when attacked through vulnerabilities being able to " +
			"read files like Path-Traversal and others), it is best practice to use a separate hardened process with proper authentication, " +
			"authorization, and audit logging to access config secrets (like credentials, private keys, client certificates, etc.). " +
			"This component is usually some kind of Vault.",
		Impact: "If this risk is unmitigated, attackers might be able to easier steal config secrets (like credentials, private keys, client certificates, etc.) once " +
			"a vulnerability to access files is present and exploited.",
		ASVS:           "V6 - Stored Cryptography Verification Requirements",
		CheatSheet:     "https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html",
		Action:         "Vault (Secret Storage)",
		Mitigation:     "Consider using a Vault (Secret Storage) to securely store and access config secrets (like credentials, private keys, client certificates, etc.).",
		Check:          "Is a Vault (Secret Storage) in place?",
		Function:       types.Architecture,
		STRIDE:         types.InformationDisclosure,
		DetectionLogic: "Models without a Vault (Secret Storage).",
		RiskAssessment: "The risk rating depends on the sensitivity of the technical asset itself and of the data assets processed and stored.",
		FalsePositives: "Models where no technical assets have any kind of sensitive config data to protect " +
			"can be considered as false positives after individual review.",
		ModelFailurePossibleReason: true,
		CWE:                        522,
	}
}

func SupportedTags() []string {
	return []string{}
}

func GenerateRisks(input *model.ParsedModel) []model.Risk {
	risks := make([]model.Risk, 0)
	hasVault := false
	var mostRelevantAsset model.TechnicalAsset
	impact := types.LowImpact
	for _, id := range input.SortedTechnicalAssetIDs() { // use the sorted one to always get the same tech asset with the highest sensitivity as example asset
		techAsset := input.TechnicalAssets[id]
		if techAsset.Technology == types.Vault {
			hasVault = true
		}
		if techAsset.HighestConfidentiality(input) >= types.Confidential ||
			techAsset.HighestIntegrity(input) >= types.Critical ||
			techAsset.HighestAvailability(input) >= types.Critical {
			impact = types.MediumImpact
		}
		if techAsset.Confidentiality >= types.Confidential ||
			techAsset.Integrity >= types.Critical ||
			techAsset.Availability >= types.Critical {
			impact = types.MediumImpact
		}
		// just for referencing the most interesting asset
		if techAsset.HighestSensitivityScore() > mostRelevantAsset.HighestSensitivityScore() {
			mostRelevantAsset = techAsset
		}
	}
	if !hasVault {
		risks = append(risks, createRisk(mostRelevantAsset, impact))
	}
	return risks
}

func createRisk(technicalAsset model.TechnicalAsset, impact types.RiskExploitationImpact) model.Risk {
	title := "<b>Missing Vault (Secret Storage)</b> in the threat model (referencing asset <b>" + technicalAsset.Title + "</b> as an example)"
	risk := model.Risk{
		Category:                     Category(),
		Severity:                     model.CalculateSeverity(types.Unlikely, impact),
		ExploitationLikelihood:       types.Unlikely,
		ExploitationImpact:           impact,
		Title:                        title,
		MostRelevantTechnicalAssetId: technicalAsset.Id,
		DataBreachProbability:        types.Improbable,
		DataBreachTechnicalAssetIDs:  []string{},
	}
	risk.SyntheticId = risk.Category.Id + "@" + technicalAsset.Id
	return risk
}