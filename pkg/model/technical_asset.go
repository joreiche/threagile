/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"fmt"
	"sort"

	"github.com/threagile/threagile/pkg/colors"
	"github.com/threagile/threagile/pkg/security/types"
)

type TechnicalAsset struct {
	Id, Title, Description                                                                  string
	Usage                                                                                   types.Usage
	Type                                                                                    types.TechnicalAssetType
	Size                                                                                    types.TechnicalAssetSize
	Technology                                                                              types.TechnicalAssetTechnology
	Machine                                                                                 types.TechnicalAssetMachine
	Internet, MultiTenant, Redundant, CustomDevelopedParts, OutOfScope, UsedAsClientByHuman bool
	Encryption                                                                              types.EncryptionStyle
	JustificationOutOfScope                                                                 string
	Owner                                                                                   string
	Confidentiality                                                                         types.Confidentiality
	Integrity, Availability                                                                 types.Criticality
	JustificationCiaRating                                                                  string
	Tags, DataAssetsProcessed, DataAssetsStored                                             []string
	DataFormatsAccepted                                                                     []types.DataFormat
	CommunicationLinks                                                                      []CommunicationLink
	DiagramTweakOrder                                                                       int
	// will be set by separate calculation step:
	RAA float64
}

func (what TechnicalAsset) IsTaggedWithAny(tags ...string) bool {
	return containsCaseInsensitiveAny(what.Tags, tags...)
}

func (what TechnicalAsset) IsTaggedWithBaseTag(baseTag string) bool {
	return isTaggedWithBaseTag(what.Tags, baseTag)
}

// first use the tag(s) of the asset itself, then their trust boundaries (recursively up) and then their shared runtime

func (what TechnicalAsset) IsTaggedWithAnyTraversingUp(model ParsedModel, tags ...string) bool {
	if containsCaseInsensitiveAny(what.Tags, tags...) {
		return true
	}
	tbID := what.GetTrustBoundaryId()
	if len(tbID) > 0 {
		if model.TrustBoundaries[tbID].IsTaggedWithAnyTraversingUp(tags...) {
			return true
		}
	}
	for _, sr := range model.SharedRuntimes {
		if Contains(sr.TechnicalAssetsRunning, what.Id) && sr.IsTaggedWithAny(tags...) {
			return true
		}
	}
	return false
}

func (what TechnicalAsset) IsSameTrustBoundary(otherAssetId string) bool {
	trustBoundaryOfMyAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[what.Id]
	trustBoundaryOfOtherAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[otherAssetId]
	return trustBoundaryOfMyAsset.Id == trustBoundaryOfOtherAsset.Id
}

func (what TechnicalAsset) IsSameExecutionEnvironment(otherAssetId string) bool {
	trustBoundaryOfMyAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[what.Id]
	trustBoundaryOfOtherAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[otherAssetId]
	if trustBoundaryOfMyAsset.Type == types.ExecutionEnvironment && trustBoundaryOfOtherAsset.Type == types.ExecutionEnvironment {
		return trustBoundaryOfMyAsset.Id == trustBoundaryOfOtherAsset.Id
	}
	return false
}

func (what TechnicalAsset) IsSameTrustBoundaryNetworkOnly(otherAssetId string) bool {
	trustBoundaryOfMyAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[what.Id]
	if !trustBoundaryOfMyAsset.Type.IsNetworkBoundary() { // find and use the parent boundary then
		trustBoundaryOfMyAsset = ParsedModelRoot.TrustBoundaries[trustBoundaryOfMyAsset.ParentTrustBoundaryID()]
	}
	trustBoundaryOfOtherAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[otherAssetId]
	if !trustBoundaryOfOtherAsset.Type.IsNetworkBoundary() { // find and use the parent boundary then
		trustBoundaryOfOtherAsset = ParsedModelRoot.TrustBoundaries[trustBoundaryOfOtherAsset.ParentTrustBoundaryID()]
	}
	return trustBoundaryOfMyAsset.Id == trustBoundaryOfOtherAsset.Id
}

func (what TechnicalAsset) HighestSensitivityScore() float64 {
	return what.Confidentiality.AttackerAttractivenessForAsset() +
		what.Integrity.AttackerAttractivenessForAsset() +
		what.Availability.AttackerAttractivenessForAsset()
}

func (what TechnicalAsset) HighestConfidentiality() types.Confidentiality {
	highest := what.Confidentiality
	for _, dataId := range what.DataAssetsProcessed {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Confidentiality > highest {
			highest = dataAsset.Confidentiality
		}
	}
	for _, dataId := range what.DataAssetsStored {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Confidentiality > highest {
			highest = dataAsset.Confidentiality
		}
	}
	return highest
}

func (what TechnicalAsset) DataAssetsProcessedSorted() []DataAsset {
	result := make([]DataAsset, 0)
	for _, assetID := range what.DataAssetsProcessed {
		result = append(result, ParsedModelRoot.DataAssets[assetID])
	}
	sort.Sort(ByDataAssetTitleSort(result))
	return result
}

func (what TechnicalAsset) DataAssetsStoredSorted() []DataAsset {
	result := make([]DataAsset, 0)
	for _, assetID := range what.DataAssetsStored {
		result = append(result, ParsedModelRoot.DataAssets[assetID])
	}
	sort.Sort(ByDataAssetTitleSort(result))
	return result
}

func (what TechnicalAsset) DataFormatsAcceptedSorted() []types.DataFormat {
	result := make([]types.DataFormat, 0)
	for _, format := range what.DataFormatsAccepted {
		result = append(result, format)
	}
	sort.Sort(ByDataFormatAcceptedSort(result))
	return result
}

func (what TechnicalAsset) CommunicationLinksSorted() []CommunicationLink {
	result := make([]CommunicationLink, 0)
	for _, format := range what.CommunicationLinks {
		result = append(result, format)
	}
	sort.Sort(ByTechnicalCommunicationLinkTitleSort(result))
	return result
}

func (what TechnicalAsset) HighestIntegrity(model ParsedModel) types.Criticality {
	highest := what.Integrity
	for _, dataId := range what.DataAssetsProcessed {
		dataAsset := model.DataAssets[dataId]
		if dataAsset.Integrity > highest {
			highest = dataAsset.Integrity
		}
	}
	for _, dataId := range what.DataAssetsStored {
		dataAsset := model.DataAssets[dataId]
		if dataAsset.Integrity > highest {
			highest = dataAsset.Integrity
		}
	}
	return highest
}

func (what TechnicalAsset) HighestAvailability(model ParsedModel) types.Criticality {
	highest := what.Availability
	for _, dataId := range what.DataAssetsProcessed {
		dataAsset := model.DataAssets[dataId]
		if dataAsset.Availability > highest {
			highest = dataAsset.Availability
		}
	}
	for _, dataId := range what.DataAssetsStored {
		dataAsset := model.DataAssets[dataId]
		if dataAsset.Availability > highest {
			highest = dataAsset.Availability
		}
	}
	return highest
}

func (what TechnicalAsset) HasDirectConnection(otherAssetId string) bool {
	for _, dataFlow := range IncomingTechnicalCommunicationLinksMappedByTargetId[what.Id] {
		if dataFlow.SourceId == otherAssetId {
			return true
		}
	}
	// check both directions, hence two times, just reversed
	for _, dataFlow := range IncomingTechnicalCommunicationLinksMappedByTargetId[otherAssetId] {
		if dataFlow.SourceId == what.Id {
			return true
		}
	}
	return false
}

func (what TechnicalAsset) GeneratedRisks() []Risk {
	resultingRisks := make([]Risk, 0)
	if len(SortedRiskCategories()) == 0 {
		fmt.Println("Uh, strange, no risks generated (yet?) and asking for them by tech asset...")
	}
	for _, category := range SortedRiskCategories() {
		risks := SortedRisksOfCategory(category)
		for _, risk := range risks {
			if risk.MostRelevantTechnicalAssetId == what.Id {
				resultingRisks = append(resultingRisks, risk)
			}
		}
	}
	sort.Sort(ByRiskSeveritySort(resultingRisks))
	return resultingRisks
}

/*
func (what TechnicalAsset) HighestRiskSeverity() RiskSeverity {
	highest := Low
	for _, risk := range what.GeneratedRisks() {
		if risk.Severity > highest {
			highest = risk.Severity
		}
	}
	return highest
}
*/

func (what TechnicalAsset) IsZero() bool {
	return len(what.Id) == 0
}

func (what TechnicalAsset) ProcessesOrStoresDataAsset(dataAssetId string) bool {
	if Contains(what.DataAssetsProcessed, dataAssetId) {
		return true
	}
	if Contains(what.DataAssetsStored, dataAssetId) {
		return true
	}
	return false
}

// red when >= confidential data stored in unencrypted technical asset

func (what TechnicalAsset) DetermineLabelColor() string {
	// TODO: Just move into main.go and let the generated risk determine the color, don't duplicate the logic here
	// Check for red
	if what.Integrity == types.MissionCritical {
		return colors.Red
	}
	for _, storedDataAsset := range what.DataAssetsStored {
		if ParsedModelRoot.DataAssets[storedDataAsset].Integrity == types.MissionCritical {
			return colors.Red
		}
	}
	for _, processedDataAsset := range what.DataAssetsProcessed {
		if ParsedModelRoot.DataAssets[processedDataAsset].Integrity == types.MissionCritical {
			return colors.Red
		}
	}
	// Check for amber
	if what.Integrity == types.Critical {
		return colors.Amber
	}
	for _, storedDataAsset := range what.DataAssetsStored {
		if ParsedModelRoot.DataAssets[storedDataAsset].Integrity == types.Critical {
			return colors.Amber
		}
	}
	for _, processedDataAsset := range what.DataAssetsProcessed {
		if ParsedModelRoot.DataAssets[processedDataAsset].Integrity == types.Critical {
			return colors.Amber
		}
	}
	return colors.Black
	/*
		if what.Encrypted {
			return colors.Black
		} else {
			if what.Confidentiality == StrictlyConfidential {
				return colors.Red
			}
			for _, storedDataAsset := range what.DataAssetsStored {
				if ParsedModelRoot.DataAssets[storedDataAsset].Confidentiality == StrictlyConfidential {
					return colors.Red
				}
			}
			if what.Confidentiality == Confidential {
				return colors.Amber
			}
			for _, storedDataAsset := range what.DataAssetsStored {
				if ParsedModelRoot.DataAssets[storedDataAsset].Confidentiality == Confidential {
					return colors.Amber
				}
			}
			return colors.Black
		}
	*/
}

// red when mission-critical integrity, but still unauthenticated (non-readonly) channels access it
// amber when critical integrity, but still unauthenticated (non-readonly) channels access it
// pink when model forgery attempt (i.e. nothing being processed or stored)

func (what TechnicalAsset) DetermineShapeBorderColor() string {
	// TODO: Just move into main.go and let the generated risk determine the color, don't duplicate the logic here
	// Check for red
	if what.Confidentiality == types.StrictlyConfidential {
		return colors.Red
	}
	for _, storedDataAsset := range what.DataAssetsStored {
		if ParsedModelRoot.DataAssets[storedDataAsset].Confidentiality == types.StrictlyConfidential {
			return colors.Red
		}
	}
	for _, processedDataAsset := range what.DataAssetsProcessed {
		if ParsedModelRoot.DataAssets[processedDataAsset].Confidentiality == types.StrictlyConfidential {
			return colors.Red
		}
	}
	// Check for amber
	if what.Confidentiality == types.Confidential {
		return colors.Amber
	}
	for _, storedDataAsset := range what.DataAssetsStored {
		if ParsedModelRoot.DataAssets[storedDataAsset].Confidentiality == types.Confidential {
			return colors.Amber
		}
	}
	for _, processedDataAsset := range what.DataAssetsProcessed {
		if ParsedModelRoot.DataAssets[processedDataAsset].Confidentiality == types.Confidential {
			return colors.Amber
		}
	}
	return colors.Black
	/*
		if what.Integrity == MissionCritical {
			for _, dataFlow := range IncomingTechnicalCommunicationLinksMappedByTargetId[what.Id] {
				if !dataFlow.Readonly && dataFlow.Authentication == NoneAuthentication {
					return colors.Red
				}
			}
		}

		if what.Integrity == Critical {
			for _, dataFlow := range IncomingTechnicalCommunicationLinksMappedByTargetId[what.Id] {
				if !dataFlow.Readonly && dataFlow.Authentication == NoneAuthentication {
					return colors.Amber
				}
			}
		}

		if len(what.DataAssetsProcessed) == 0 && len(what.DataAssetsStored) == 0 {
			return colors.Pink // pink, because it's strange when too many technical assets process no data... some are ok, but many in a diagram is a sign of model forgery...
		}

		return colors.Black
	*/
}

/*
// Loops over all data assets (stored and processed by this technical asset) and determines for each
// data asset, how many percentage of the data risk is reduced when this technical asset has all risks mitigated.
// Example: This means if the data asset is loosing a risk and thus getting from red to amber it counts as 1.
// Other example: When only one out of four lines (see data risk mapping) leading to red tech assets are removed by
// the mitigations, then this counts as 0.25. The overall sum is returned.
func (what TechnicalAsset) QuickWins() float64 {
	result := 0.0
	uniqueDataAssetsStoredAndProcessed := make(map[string]interface{})
	for _, dataAssetId := range what.DataAssetsStored {
		uniqueDataAssetsStoredAndProcessed[dataAssetId] = true
	}
	for _, dataAssetId := range what.DataAssetsProcessed {
		uniqueDataAssetsStoredAndProcessed[dataAssetId] = true
	}
	highestSeverity := HighestSeverityStillAtRisk(what.GeneratedRisks())
	for dataAssetId, _ := range uniqueDataAssetsStoredAndProcessed {
		dataAsset := ParsedModelRoot.DataAssets[dataAssetId]
		if dataAsset.IdentifiedRiskSeverityStillAtRisk() <= highestSeverity {
			howManySameLevelCausingUsagesOfThisData := 0.0
			for techAssetId, risks := range dataAsset.IdentifiedRisksByResponsibleTechnicalAssetId() {
				if !ParsedModelRoot.TechnicalAssets[techAssetId].OutOfScope {
					for _, risk := range risks {
						if len(risk.MostRelevantTechnicalAssetId) > 0 { // T O D O caching of generated risks inside the method?
							if HighestSeverityStillAtRisk(ParsedModelRoot.TechnicalAssets[risk.MostRelevantTechnicalAssetId].GeneratedRisks()) == highestSeverity {
								howManySameLevelCausingUsagesOfThisData++
								break
							}
						}
					}
				}
			}
			if howManySameLevelCausingUsagesOfThisData > 0 {
				result += 1.0 / howManySameLevelCausingUsagesOfThisData
			}
		}
	}
	return result
}
*/

// dotted when model forgery attempt (i.e. nothing being processed or stored)

func (what TechnicalAsset) DetermineShapeBorderLineStyle() string {
	if len(what.DataAssetsProcessed) == 0 && len(what.DataAssetsStored) == 0 || what.OutOfScope {
		return "dotted" // dotted, because it's strange when too many technical communication links transfer no data... some ok, but many in a diagram ist a sign of model forgery...
	}
	return "solid"
}

// 3 when redundant

func (what TechnicalAsset) DetermineShapePeripheries() int {
	if what.Redundant {
		return 2
	}
	return 1
}

func (what TechnicalAsset) DetermineShapeStyle() string {
	return "filled"
}

func (what TechnicalAsset) GetTrustBoundaryId() string {
	for _, trustBoundary := range ParsedModelRoot.TrustBoundaries {
		for _, techAssetInside := range trustBoundary.TechnicalAssetsInside {
			if techAssetInside == what.Id {
				return trustBoundary.Id
			}
		}
	}
	return ""
}

type ByTechnicalAssetRiskSeverityAndTitleSortStillAtRisk []TechnicalAsset

func (what ByTechnicalAssetRiskSeverityAndTitleSortStillAtRisk) Len() int { return len(what) }
func (what ByTechnicalAssetRiskSeverityAndTitleSortStillAtRisk) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByTechnicalAssetRiskSeverityAndTitleSortStillAtRisk) Less(i, j int) bool {
	risksLeft := ReduceToOnlyStillAtRisk(what[i].GeneratedRisks())
	risksRight := ReduceToOnlyStillAtRisk(what[j].GeneratedRisks())
	highestSeverityLeft := HighestSeverityStillAtRisk(risksLeft)
	highestSeverityRight := HighestSeverityStillAtRisk(risksRight)
	var result bool
	if highestSeverityLeft == highestSeverityRight {
		if len(risksLeft) == 0 && len(risksRight) > 0 {
			return false
		} else if len(risksLeft) > 0 && len(risksRight) == 0 {
			return true
		} else {
			result = what[i].Title < what[j].Title
		}
	} else {
		result = highestSeverityLeft > highestSeverityRight
	}
	if what[i].OutOfScope && what[j].OutOfScope {
		result = what[i].Title < what[j].Title
	} else if what[i].OutOfScope {
		result = false
	} else if what[j].OutOfScope {
		result = true
	}
	return result
}

type ByTechnicalAssetRAAAndTitleSort []TechnicalAsset

func (what ByTechnicalAssetRAAAndTitleSort) Len() int      { return len(what) }
func (what ByTechnicalAssetRAAAndTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByTechnicalAssetRAAAndTitleSort) Less(i, j int) bool {
	raaLeft := what[i].RAA
	raaRight := what[j].RAA
	if raaLeft == raaRight {
		return what[i].Title < what[j].Title
	}
	return raaLeft > raaRight
}

/*
type ByTechnicalAssetQuickWinsAndTitleSort []TechnicalAsset

func (what ByTechnicalAssetQuickWinsAndTitleSort) Len() int      { return len(what) }
func (what ByTechnicalAssetQuickWinsAndTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByTechnicalAssetQuickWinsAndTitleSort) Less(i, j int) bool {
	qwLeft := what[i].QuickWins()
	qwRight := what[j].QuickWins()
	if qwLeft == qwRight {
		return what[i].Title < what[j].Title
	}
	return qwLeft > qwRight
}
*/

type ByTechnicalAssetTitleSort []TechnicalAsset

func (what ByTechnicalAssetTitleSort) Len() int      { return len(what) }
func (what ByTechnicalAssetTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByTechnicalAssetTitleSort) Less(i, j int) bool {
	return what[i].Title < what[j].Title
}
