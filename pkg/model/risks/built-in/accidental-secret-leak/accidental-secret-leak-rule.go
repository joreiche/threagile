package accidental_secret_leak

import (
	"github.com/threagile/threagile/pkg/model"
	"github.com/threagile/threagile/pkg/security/types"
)

func Rule() model.RiskRule {
	return model.RiskRule{
		Category:      Category,
		SupportedTags: SupportedTags,
		GenerateRisks: GenerateRisks,
	}
}

func Category() types.RiskCategory {
	return types.RiskCategory{
		Id:    "accidental-secret-leak",
		Title: "Accidental Secret Leak",
		Description: "Sourcecode repositories (including their histories) as well as artifact registries can accidentally contain secrets like " +
			"checked-in or packaged-in passwords, API tokens, certificates, crypto keys, etc.",
		Impact: "If this risk is unmitigated, attackers which have access to affected sourcecode repositories or artifact registries might " +
			"find secrets accidentally checked-in.",
		ASVS:       "V14 - Configuration Verification Requirements",
		CheatSheet: "https://cheatsheetseries.owasp.org/cheatsheets/Attack_Surface_Analysis_Cheat_Sheet.html",
		Action:     "Build Pipeline Hardening",
		Mitigation: "Establish measures preventing accidental check-in or package-in of secrets into sourcecode repositories " +
			"and artifact registries. This starts by using good .gitignore and .dockerignore files, but does not stop there. " +
			"See for example tools like <i>\"git-secrets\" or \"Talisman\"</i> to have check-in preventive measures for secrets. " +
			"Consider also to regularly scan your repositories for secrets accidentally checked-in using scanning tools like <i>\"gitleaks\" or \"gitrob\"</i>.",
		Check:                      "Are recommendations from the linked cheat sheet and referenced ASVS chapter applied?",
		Function:                   types.Operations,
		STRIDE:                     types.InformationDisclosure,
		DetectionLogic:             "In-scope sourcecode repositories and artifact registries.",
		RiskAssessment:             "The risk rating depends on the sensitivity of the technical asset itself and of the data assets processed and stored.",
		FalsePositives:             "Usually no false positives.",
		ModelFailurePossibleReason: false,
		CWE:                        200,
	}
}

func SupportedTags() []string {
	return []string{"git", "nexus"}
}

func GenerateRisks(parsedModel *model.ParsedModel) []types.Risk {
	risks := make([]types.Risk, 0)
	for _, id := range parsedModel.SortedTechnicalAssetIDs() {
		techAsset := parsedModel.TechnicalAssets[id]
		if !techAsset.OutOfScope &&
			(techAsset.Technology == types.SourcecodeRepository || techAsset.Technology == types.ArtifactRegistry) {
			var risk types.Risk
			if techAsset.IsTaggedWithAny("git") {
				risk = createRisk(parsedModel, techAsset, "Git", "Git Leak Prevention")
			} else {
				risk = createRisk(parsedModel, techAsset, "", "")
			}
			risks = append(risks, risk)
		}
	}
	return risks
}

func createRisk(parsedModel *model.ParsedModel, technicalAsset types.TechnicalAsset, prefix, details string) types.Risk {
	if len(prefix) > 0 {
		prefix = " (" + prefix + ")"
	}
	title := "<b>Accidental Secret Leak" + prefix + "</b> risk at <b>" + technicalAsset.Title + "</b>"
	if len(details) > 0 {
		title += ": <u>" + details + "</u>"
	}
	impact := types.LowImpact
	if technicalAsset.HighestConfidentiality(parsedModel) >= types.Confidential ||
		technicalAsset.HighestIntegrity(parsedModel) >= types.Critical ||
		technicalAsset.HighestAvailability(parsedModel) >= types.Critical {
		impact = types.MediumImpact
	}
	if technicalAsset.HighestConfidentiality(parsedModel) == types.StrictlyConfidential ||
		technicalAsset.HighestIntegrity(parsedModel) == types.MissionCritical ||
		technicalAsset.HighestAvailability(parsedModel) == types.MissionCritical {
		impact = types.HighImpact
	}
	// create risk
	risk := types.Risk{
		CategoryId:                   Category().Id,
		Severity:                     types.CalculateSeverity(types.Unlikely, impact),
		ExploitationLikelihood:       types.Unlikely,
		ExploitationImpact:           impact,
		Title:                        title,
		MostRelevantTechnicalAssetId: technicalAsset.Id,
		DataBreachProbability:        types.Probable,
		DataBreachTechnicalAssetIDs:  []string{technicalAsset.Id},
	}
	risk.SyntheticId = risk.CategoryId + "@" + technicalAsset.Id
	return risk
}
