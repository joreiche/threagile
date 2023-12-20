/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"sort"

	"github.com/threagile/threagile/pkg/security/types"
)

type DataAsset struct {
	Id                     string                `yaml:"id" json:"id"`                   // TODO: tag here still required?
	Title                  string                `yaml:"title" json:"title"`             // TODO: tag here still required?
	Description            string                `yaml:"description" json:"description"` // TODO: tag here still required?
	Usage                  types.Usage           `yaml:"usage" json:"usage"`
	Tags                   []string              `yaml:"tags" json:"tags"`
	Origin                 string                `yaml:"origin" json:"origin"`
	Owner                  string                `yaml:"owner" json:"owner"`
	Quantity               types.Quantity        `yaml:"quantity" json:"quantity"`
	Confidentiality        types.Confidentiality `yaml:"confidentiality" json:"confidentiality"`
	Integrity              types.Criticality     `yaml:"integrity" json:"integrity"`
	Availability           types.Criticality     `yaml:"availability" json:"availability"`
	JustificationCiaRating string                `yaml:"justification_cia_rating" json:"justification_cia_rating"`
}

func (what DataAsset) IsTaggedWithAny(tags ...string) bool {
	return containsCaseInsensitiveAny(what.Tags, tags...)
}

func (what DataAsset) IsTaggedWithBaseTag(baseTag string) bool {
	return isTaggedWithBaseTag(what.Tags, baseTag)
}

/*
func (what DataAsset) IsAtRisk() bool {
	for _, techAsset := range what.ProcessedByTechnicalAssetsSorted() {
		if len(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks())) > 0 {
			return true
		}
	}
	for _, techAsset := range what.StoredByTechnicalAssetsSorted() {
		if len(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks())) > 0 {
			return true
		}
	}
	return false
}
*/

/*
func (what DataAsset) IdentifiedRiskSeverityStillAtRisk() RiskSeverity {
	highestRiskSeverity := Low
	for _, techAsset := range what.ProcessedByTechnicalAssetsSorted() {
		candidateSeverity := HighestSeverityStillAtRisk(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks()))
		if candidateSeverity > highestRiskSeverity {
			highestRiskSeverity = candidateSeverity
		}
	}
	for _, techAsset := range what.StoredByTechnicalAssetsSorted() {
		candidateSeverity := HighestSeverityStillAtRisk(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks()))
		if candidateSeverity > highestRiskSeverity {
			highestRiskSeverity = candidateSeverity
		}
	}
	return highestRiskSeverity
}
*/

func (what DataAsset) IdentifiedRisksByResponsibleTechnicalAssetId(model ParsedModel) map[string][]Risk {
	uniqueTechAssetIDsResponsibleForThisDataAsset := make(map[string]interface{})
	for _, techAsset := range what.ProcessedByTechnicalAssetsSorted() {
		if len(techAsset.GeneratedRisks()) > 0 {
			uniqueTechAssetIDsResponsibleForThisDataAsset[techAsset.Id] = true
		}
	}
	for _, techAsset := range what.StoredByTechnicalAssetsSorted() {
		if len(techAsset.GeneratedRisks()) > 0 {
			uniqueTechAssetIDsResponsibleForThisDataAsset[techAsset.Id] = true
		}
	}

	result := make(map[string][]Risk)
	for techAssetId := range uniqueTechAssetIDsResponsibleForThisDataAsset {
		result[techAssetId] = append(result[techAssetId], model.TechnicalAssets[techAssetId].GeneratedRisks()...)
	}
	return result
}

func (what DataAsset) IsDataBreachPotentialStillAtRisk() bool {
	for _, risk := range FilteredByStillAtRisk() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				return true
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				return true
			}
		}
	}
	return false
}

func (what DataAsset) IdentifiedDataBreachProbability() types.DataBreachProbability {
	highestProbability := types.Improbable
	for _, risk := range AllRisks() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
		}
	}
	return highestProbability
}

func (what DataAsset) IdentifiedDataBreachProbabilityStillAtRisk() types.DataBreachProbability {
	highestProbability := types.Improbable
	for _, risk := range FilteredByStillAtRisk() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
		}
	}
	return highestProbability
}

func (what DataAsset) IdentifiedDataBreachProbabilityRisksStillAtRisk() []Risk {
	result := make([]Risk, 0)
	for _, risk := range FilteredByStillAtRisk() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				result = append(result, risk)
				break
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				result = append(result, risk)
				break
			}
		}
	}
	return result
}

func (what DataAsset) IdentifiedDataBreachProbabilityRisks() []Risk {
	result := make([]Risk, 0)
	for _, risk := range AllRisks() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				result = append(result, risk)
				break
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				result = append(result, risk)
				break
			}
		}
	}
	return result
}

func (what DataAsset) ProcessedByTechnicalAssetsSorted() []TechnicalAsset {
	result := make([]TechnicalAsset, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, candidateID := range technicalAsset.DataAssetsProcessed {
			if candidateID == what.Id {
				result = append(result, technicalAsset)
			}
		}
	}
	sort.Sort(ByTechnicalAssetTitleSort(result))
	return result
}

func (what DataAsset) StoredByTechnicalAssetsSorted() []TechnicalAsset {
	result := make([]TechnicalAsset, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, candidateID := range technicalAsset.DataAssetsStored {
			if candidateID == what.Id {
				result = append(result, technicalAsset)
			}
		}
	}
	sort.Sort(ByTechnicalAssetTitleSort(result))
	return result
}

func (what DataAsset) SentViaCommLinksSorted() []CommunicationLink {
	result := make([]CommunicationLink, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, commLink := range technicalAsset.CommunicationLinks {
			for _, candidateID := range commLink.DataAssetsSent {
				if candidateID == what.Id {
					result = append(result, commLink)
				}
			}
		}
	}
	sort.Sort(ByTechnicalCommunicationLinkTitleSort(result))
	return result
}

func (what DataAsset) ReceivedViaCommLinksSorted() []CommunicationLink {
	result := make([]CommunicationLink, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, commLink := range technicalAsset.CommunicationLinks {
			for _, candidateID := range commLink.DataAssetsReceived {
				if candidateID == what.Id {
					result = append(result, commLink)
				}
			}
		}
	}
	sort.Sort(ByTechnicalCommunicationLinkTitleSort(result))
	return result
}

type ByDataAssetDataBreachProbabilityAndTitleSort []DataAsset

func (what ByDataAssetDataBreachProbabilityAndTitleSort) Len() int { return len(what) }
func (what ByDataAssetDataBreachProbabilityAndTitleSort) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByDataAssetDataBreachProbabilityAndTitleSort) Less(i, j int) bool {
	highestDataBreachProbabilityLeft := what[i].IdentifiedDataBreachProbability()
	highestDataBreachProbabilityRight := what[j].IdentifiedDataBreachProbability()
	if highestDataBreachProbabilityLeft == highestDataBreachProbabilityRight {
		return what[i].Title < what[j].Title
	}
	return highestDataBreachProbabilityLeft > highestDataBreachProbabilityRight
}

type ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk []DataAsset

func (what ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk) Len() int { return len(what) }
func (what ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk) Less(i, j int) bool {
	risksLeft := what[i].IdentifiedDataBreachProbabilityRisksStillAtRisk()
	risksRight := what[j].IdentifiedDataBreachProbabilityRisksStillAtRisk()
	highestDataBreachProbabilityLeft := what[i].IdentifiedDataBreachProbabilityStillAtRisk()
	highestDataBreachProbabilityRight := what[j].IdentifiedDataBreachProbabilityStillAtRisk()
	if highestDataBreachProbabilityLeft == highestDataBreachProbabilityRight {
		if len(risksLeft) == 0 && len(risksRight) > 0 {
			return false
		}
		if len(risksLeft) > 0 && len(risksRight) == 0 {
			return true
		}
		return what[i].Title < what[j].Title
	}
	return highestDataBreachProbabilityLeft > highestDataBreachProbabilityRight
}

type ByDataAssetTitleSort []DataAsset

func (what ByDataAssetTitleSort) Len() int      { return len(what) }
func (what ByDataAssetTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByDataAssetTitleSort) Less(i, j int) bool {
	return what[i].Title < what[j].Title
}
