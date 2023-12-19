package unguarded_direct_datastore_access

import (
	"github.com/threagile/threagile/model"
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
		Id:          "unguarded-direct-datastore-access",
		Title:       "Unguarded Direct Datastore Access",
		Description: "Data stores accessed across trust boundaries must be guarded by some protecting service or application.",
		Impact:      "If this risk is unmitigated, attackers might be able to directly attack sensitive data stores without any protecting components in-between.",
		ASVS:        "V1 - Architecture, Design and Threat Modeling Requirements",
		CheatSheet:  "https://cheatsheetseries.owasp.org/cheatsheets/Attack_Surface_Analysis_Cheat_Sheet.html",
		Action:      "Encapsulation of Datastore",
		Mitigation:  "Encapsulate the datastore access behind a guarding service or application.",
		Check:       "Are recommendations from the linked cheat sheet and referenced ASVS chapter applied?",
		Function:    types.Architecture,
		STRIDE:      types.ElevationOfPrivilege,
		DetectionLogic: "In-scope technical assets of type " + types.Datastore.String() + " (except " + types.IdentityStoreLDAP.String() + " when accessed from " + types.IdentityProvider.String() + " and " + types.FileServer.String() + " when accessed via file transfer protocols) with confidentiality rating " +
			"of " + types.Confidential.String() + " (or higher) or with integrity rating of " + types.Critical.String() + " (or higher) " +
			"which have incoming data-flows from assets outside across a network trust-boundary. DevOps config and deployment access is excluded from this risk.", // TODO new rule "missing bastion host"?
		RiskAssessment: "The matching technical assets are at " + types.LowSeverity.String() + " risk. When either the " +
			"confidentiality rating is " + types.StrictlyConfidential.String() + " or the integrity rating " +
			"is " + types.MissionCritical.String() + ", the risk-rating is considered " + types.MediumSeverity.String() + ". " +
			"For assets with RAA values higher than 40 % the risk-rating increases.",
		FalsePositives:             "When the caller is considered fully trusted as if it was part of the datastore itself.",
		ModelFailurePossibleReason: false,
		CWE:                        501,
	}
}

func SupportedTags() []string {
	return []string{}
}

// check for data stores that should not be accessed directly across trust boundaries

func GenerateRisks(input *model.ParsedModel) []model.Risk {
	risks := make([]model.Risk, 0)
	for _, id := range model.SortedTechnicalAssetIDs() {
		technicalAsset := input.TechnicalAssets[id]
		if !technicalAsset.OutOfScope && technicalAsset.Type == types.Datastore {
			for _, incomingAccess := range model.IncomingTechnicalCommunicationLinksMappedByTargetId[technicalAsset.Id] {
				sourceAsset := input.TechnicalAssets[incomingAccess.SourceId]
				if (technicalAsset.Technology == types.IdentityStoreLDAP || technicalAsset.Technology == types.IdentityStoreDatabase) &&
					sourceAsset.Technology == types.IdentityProvider {
					continue
				}
				if technicalAsset.Confidentiality >= types.Confidential || technicalAsset.Integrity >= types.Critical {
					if incomingAccess.IsAcrossTrustBoundaryNetworkOnly() && !FileServerAccessViaFTP(technicalAsset, incomingAccess) &&
						incomingAccess.Usage != types.DevOps && !model.IsSharingSameParentTrustBoundary(technicalAsset, sourceAsset) {
						highRisk := technicalAsset.Confidentiality == types.StrictlyConfidential ||
							technicalAsset.Integrity == types.MissionCritical
						risks = append(risks, createRisk(technicalAsset, incomingAccess,
							input.TechnicalAssets[incomingAccess.SourceId], highRisk))
					}
				}
			}
		}
	}
	return risks
}

func FileServerAccessViaFTP(technicalAsset model.TechnicalAsset, incomingAccess model.CommunicationLink) bool {
	return technicalAsset.Technology == types.FileServer &&
		(incomingAccess.Protocol == types.FTP || incomingAccess.Protocol == types.FTPS || incomingAccess.Protocol == types.SFTP)
}

func createRisk(dataStore model.TechnicalAsset, dataFlow model.CommunicationLink, clientOutsideTrustBoundary model.TechnicalAsset, moreRisky bool) model.Risk {
	impact := types.LowImpact
	if moreRisky || dataStore.RAA > 40 {
		impact = types.MediumImpact
	}
	risk := model.Risk{
		Category:               Category(),
		Severity:               model.CalculateSeverity(types.Likely, impact),
		ExploitationLikelihood: types.Likely,
		ExploitationImpact:     impact,
		Title: "<b>Unguarded Direct Datastore Access</b> of <b>" + dataStore.Title + "</b> by <b>" +
			clientOutsideTrustBoundary.Title + "</b> via <b>" + dataFlow.Title + "</b>",
		MostRelevantTechnicalAssetId:    dataStore.Id,
		MostRelevantCommunicationLinkId: dataFlow.Id,
		DataBreachProbability:           types.Improbable,
		DataBreachTechnicalAssetIDs:     []string{dataStore.Id},
	}
	risk.SyntheticId = risk.Category.Id + "@" + dataFlow.Id + "@" + clientOutsideTrustBoundary.Id + "@" + dataStore.Id
	return risk
}
