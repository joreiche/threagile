package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/threagile/threagile/pkg/colors"
	"github.com/threagile/threagile/pkg/model"
	"github.com/threagile/threagile/pkg/security/types"
)

var CommunicationLinks map[string]model.CommunicationLink // TODO as part of "ParsedModelRoot"?
var IncomingTechnicalCommunicationLinksMappedByTargetId map[string][]model.CommunicationLink
var DirectContainingTrustBoundaryMappedByTechnicalAssetId map[string]model.TrustBoundary
var DirectContainingSharedRuntimeMappedByTechnicalAssetId map[string]model.SharedRuntime

var GeneratedRisksByCategory map[model.RiskCategory][]model.Risk
var GeneratedRisksBySyntheticId map[string]model.Risk

var AllSupportedTags map[string]bool

var (
	_ = SortedKeysOfDataAssets
	_ = SortedKeysOfTechnicalAssets
	_ = SortedDataAssetsByDataBreachProbabilityAndTitleStillAtRisk
	_ = ReduceToOnlyHighRisk
	_ = ReduceToOnlyMediumRisk
	_ = ReduceToOnlyLowRisk
)

func Init() {
	CommunicationLinks = make(map[string]model.CommunicationLink)
	IncomingTechnicalCommunicationLinksMappedByTargetId = make(map[string][]model.CommunicationLink)
	DirectContainingTrustBoundaryMappedByTechnicalAssetId = make(map[string]model.TrustBoundary)
	DirectContainingSharedRuntimeMappedByTechnicalAssetId = make(map[string]model.SharedRuntime)
	GeneratedRisksByCategory = make(map[model.RiskCategory][]model.Risk)
	GeneratedRisksBySyntheticId = make(map[string]model.Risk)
	AllSupportedTags = make(map[string]bool)
}

func AddToListOfSupportedTags(tags []string) {
	for _, tag := range tags {
		AllSupportedTags[tag] = true
	}
}

// === To be used by model macros etc. =======================

type ByOrderAndIdSort []TechnicalAsset

func (what ByOrderAndIdSort) Len() int      { return len(what) }
func (what ByOrderAndIdSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByOrderAndIdSort) Less(i, j int) bool {
	if what[i].DiagramTweakOrder == what[j].DiagramTweakOrder {
		return what[i].Id > what[j].Id
	}
	return what[i].DiagramTweakOrder < what[j].DiagramTweakOrder
}

type ByTrustBoundaryTitleSort []TrustBoundary

func (what ByTrustBoundaryTitleSort) Len() int      { return len(what) }
func (what ByTrustBoundaryTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByTrustBoundaryTitleSort) Less(i, j int) bool {
	return what[i].Title < what[j].Title
}

type BySharedRuntimeTitleSort []SharedRuntime

func (what BySharedRuntimeTitleSort) Len() int      { return len(what) }
func (what BySharedRuntimeTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what BySharedRuntimeTitleSort) Less(i, j int) bool {
	return what[i].Title < what[j].Title
}

type ByDataFormatAcceptedSort []types.DataFormat

func (what ByDataFormatAcceptedSort) Len() int      { return len(what) }
func (what ByDataFormatAcceptedSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByDataFormatAcceptedSort) Less(i, j int) bool {
	return what[i].String() < what[j].String()
}

func SortedTechnicalAssetIDs() []string {
	res := make([]string, 0)
	for id := range ParsedModelRoot.TechnicalAssets {
		res = append(res, id)
	}
	sort.Strings(res)
	return res
}

func TagsActuallyUsed() []string {
	result := make([]string, 0)
	for _, tag := range ParsedModelRoot.TagsAvailable {
		if len(TechnicalAssetsTaggedWithAny(tag)) > 0 ||
			len(CommunicationLinksTaggedWithAny(tag)) > 0 ||
			len(DataAssetsTaggedWithAny(tag)) > 0 ||
			len(TrustBoundariesTaggedWithAny(tag)) > 0 ||
			len(SharedRuntimesTaggedWithAny(tag)) > 0 {
			result = append(result, tag)
		}
	}
	return result
}

// === Sorting stuff =====================================

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfIndividualRiskCategories() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.IndividualRiskCategories {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfSecurityRequirements() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.SecurityRequirements {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfAbuseCases() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.AbuseCases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfQuestions() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.Questions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfDataAssets() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.DataAssets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfTechnicalAssets() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.TechnicalAssets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TechnicalAssetsTaggedWithAny(tags ...string) []TechnicalAsset {
	result := make([]TechnicalAsset, 0)
	for _, candidate := range ParsedModelRoot.TechnicalAssets {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

func CommunicationLinksTaggedWithAny(tags ...string) []CommunicationLink {
	result := make([]CommunicationLink, 0)
	for _, asset := range ParsedModelRoot.TechnicalAssets {
		for _, candidate := range asset.CommunicationLinks {
			if candidate.IsTaggedWithAny(tags...) {
				result = append(result, candidate)
			}
		}
	}
	return result
}

func DataAssetsTaggedWithAny(tags ...string) []DataAsset {
	result := make([]DataAsset, 0)
	for _, candidate := range ParsedModelRoot.DataAssets {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

func TrustBoundariesTaggedWithAny(tags ...string) []TrustBoundary {
	result := make([]TrustBoundary, 0)
	for _, candidate := range ParsedModelRoot.TrustBoundaries {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

func SharedRuntimesTaggedWithAny(tags ...string) []SharedRuntime {
	result := make([]SharedRuntime, 0)
	for _, candidate := range ParsedModelRoot.SharedRuntimes {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedTechnicalAssetsByTitle() []TechnicalAsset {
	assets := make([]TechnicalAsset, 0)
	for _, asset := range ParsedModelRoot.TechnicalAssets {
		assets = append(assets, asset)
	}
	sort.Sort(ByTechnicalAssetTitleSort(assets))
	return assets
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedDataAssetsByTitle() []DataAsset {
	assets := make([]DataAsset, 0)
	for _, asset := range ParsedModelRoot.DataAssets {
		assets = append(assets, asset)
	}
	sort.Sort(byDataAssetTitleSort(assets))
	return assets
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedDataAssetsByDataBreachProbabilityAndTitleStillAtRisk() []DataAsset {
	assets := make([]DataAsset, 0)
	for _, asset := range ParsedModelRoot.DataAssets {
		assets = append(assets, asset)
	}
	sort.Sort(ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk(assets))
	return assets
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedDataAssetsByDataBreachProbabilityAndTitle() []DataAsset {
	assets := make([]DataAsset, 0)
	for _, asset := range ParsedModelRoot.DataAssets {
		assets = append(assets, asset)
	}
	sort.Sort(ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk(assets))
	return assets
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedTechnicalAssetsByRiskSeverityAndTitle() []TechnicalAsset {
	assets := make([]TechnicalAsset, 0)
	for _, asset := range ParsedModelRoot.TechnicalAssets {
		assets = append(assets, asset)
	}
	sort.Sort(ByTechnicalAssetRiskSeverityAndTitleSortStillAtRisk(assets))
	return assets
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedTechnicalAssetsByRAAAndTitle() []TechnicalAsset {
	assets := make([]TechnicalAsset, 0)
	for _, asset := range ParsedModelRoot.TechnicalAssets {
		assets = append(assets, asset)
	}
	sort.Sort(ByTechnicalAssetRAAAndTitleSort(assets))
	return assets
}

/*
// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:
func SortedTechnicalAssetsByQuickWinsAndTitle() []TechnicalAsset {
	assets := make([]TechnicalAsset, 0)
	for _, asset := range ParsedModelRoot.TechnicalAssets {
		if !asset.OutOfScope && asset.QuickWins() > 0 {
			assets = append(assets, asset)
		}
	}
	sort.Sort(ByTechnicalAssetQuickWinsAndTitleSort(assets))
	return assets
}
*/

func OutOfScopeTechnicalAssets() []TechnicalAsset {
	assets := make([]TechnicalAsset, 0)
	for _, asset := range ParsedModelRoot.TechnicalAssets {
		if asset.OutOfScope {
			assets = append(assets, asset)
		}
	}
	sort.Sort(ByTechnicalAssetTitleSort(assets))
	return assets
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfTrustBoundaries() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.TrustBoundaries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func SortedTrustBoundariesByTitle() []TrustBoundary {
	boundaries := make([]TrustBoundary, 0)
	for _, boundary := range ParsedModelRoot.TrustBoundaries {
		boundaries = append(boundaries, boundary)
	}
	sort.Sort(ByTrustBoundaryTitleSort(boundaries))
	return boundaries
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedKeysOfSharedRuntime() []string {
	keys := make([]string, 0)
	for k := range ParsedModelRoot.SharedRuntimes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func SortedSharedRuntimesByTitle() []SharedRuntime {
	result := make([]SharedRuntime, 0)
	for _, runtime := range ParsedModelRoot.SharedRuntimes {
		result = append(result, runtime)
	}
	sort.Sort(BySharedRuntimeTitleSort(result))
	return result
}

func QuestionsUnanswered() int {
	result := 0
	for _, answer := range ParsedModelRoot.Questions {
		if len(strings.TrimSpace(answer)) == 0 {
			result++
		}
	}
	return result
}

// === Style stuff =======================================

// Line Styles:

// dotted when model forgery attempt (i.e. nothing being sent and received)

func (what CommunicationLink) DetermineArrowLineStyle() string {
	if len(what.DataAssetsSent) == 0 && len(what.DataAssetsReceived) == 0 {
		return "dotted" // dotted, because it's strange when too many technical communication links transfer no data... some ok, but many in a diagram ist a sign of model forgery...
	}
	if what.Usage == types.DevOps {
		return "dashed"
	}
	return "solid"
}

// Pen Widths:

func (what CommunicationLink) DetermineArrowPenWidth() string {
	if what.DetermineArrowColor() == colors.Pink {
		return fmt.Sprintf("%f", 3.0)
	}
	if what.DetermineArrowColor() != colors.Black {
		return fmt.Sprintf("%f", 2.5)
	}
	return fmt.Sprintf("%f", 1.5)
}

func (what TechnicalAsset) DetermineShapeBorderPenWidth() string {
	if what.DetermineShapeBorderColor() == colors.Pink {
		return fmt.Sprintf("%f", 3.5)
	}
	if what.DetermineShapeBorderColor() != colors.Black {
		return fmt.Sprintf("%f", 3.0)
	}
	return fmt.Sprintf("%f", 2.0)
}

// Contains tells whether a contains x (in an unsorted slice)
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func (what CommunicationLink) DetermineLabelColor() string {
	// TODO: Just move into main.go and let the generated risk determine the color, don't duplicate the logic here
	/*
		if dataFlow.Protocol.IsEncrypted() {
			return colors.Gray
		} else {*/
	// check for red
	for _, sentDataAsset := range what.DataAssetsSent {
		if ParsedModelRoot.DataAssets[sentDataAsset].Integrity == types.MissionCritical {
			return colors.Red
		}
	}
	for _, receivedDataAsset := range what.DataAssetsReceived {
		if ParsedModelRoot.DataAssets[receivedDataAsset].Integrity == types.MissionCritical {
			return colors.Red
		}
	}
	// check for amber
	for _, sentDataAsset := range what.DataAssetsSent {
		if ParsedModelRoot.DataAssets[sentDataAsset].Integrity == types.Critical {
			return colors.Amber
		}
	}
	for _, receivedDataAsset := range what.DataAssetsReceived {
		if ParsedModelRoot.DataAssets[receivedDataAsset].Integrity == types.Critical {
			return colors.Amber
		}
	}
	// default
	return colors.Gray

}

// pink when model forgery attempt (i.e. nothing being sent and received)

func (what CommunicationLink) DetermineArrowColor() string {
	// TODO: Just move into main.go and let the generated risk determine the color, don't duplicate the logic here
	if len(what.DataAssetsSent) == 0 && len(what.DataAssetsReceived) == 0 ||
		what.Protocol == types.UnknownProtocol {
		return colors.Pink // pink, because it's strange when too many technical communication links transfer no data... some ok, but many in a diagram ist a sign of model forgery...
	}
	if what.Usage == types.DevOps {
		return colors.MiddleLightGray
	} else if what.VPN {
		return colors.DarkBlue
	} else if what.IpFiltered {
		return colors.Brown
	}
	// check for red
	for _, sentDataAsset := range what.DataAssetsSent {
		if ParsedModelRoot.DataAssets[sentDataAsset].Confidentiality == types.StrictlyConfidential {
			return colors.Red
		}
	}
	for _, receivedDataAsset := range what.DataAssetsReceived {
		if ParsedModelRoot.DataAssets[receivedDataAsset].Confidentiality == types.StrictlyConfidential {
			return colors.Red
		}
	}
	// check for amber
	for _, sentDataAsset := range what.DataAssetsSent {
		if ParsedModelRoot.DataAssets[sentDataAsset].Confidentiality == types.Confidential {
			return colors.Amber
		}
	}
	for _, receivedDataAsset := range what.DataAssetsReceived {
		if ParsedModelRoot.DataAssets[receivedDataAsset].Confidentiality == types.Confidential {
			return colors.Amber
		}
	}
	// default
	return colors.Black
	/*
		} else if dataFlow.Authentication != NoneAuthentication {
			return colors.Black
		} else {
			// check for red
			for _, sentDataAsset := range dataFlow.DataAssetsSent { // first check if any red?
				if ParsedModelRoot.DataAssets[sentDataAsset].Integrity == MissionCritical {
					return colors.Red
				}
			}
			for _, receivedDataAsset := range dataFlow.DataAssetsReceived { // first check if any red?
				if ParsedModelRoot.DataAssets[receivedDataAsset].Integrity == MissionCritical {
					return colors.Red
				}
			}
			// check for amber
			for _, sentDataAsset := range dataFlow.DataAssetsSent { // then check if any amber?
				if ParsedModelRoot.DataAssets[sentDataAsset].Integrity == Critical {
					return colors.Amber
				}
			}
			for _, receivedDataAsset := range dataFlow.DataAssetsReceived { // then check if any amber?
				if ParsedModelRoot.DataAssets[receivedDataAsset].Integrity == Critical {
					return colors.Amber
				}
			}
			return colors.Black
		}
	*/
}

func (what TechnicalAsset) DetermineShapeFillColor() string {
	fillColor := colors.VeryLightGray
	if len(what.DataAssetsProcessed) == 0 && len(what.DataAssetsStored) == 0 ||
		what.Technology == types.UnknownTechnology {
		fillColor = colors.LightPink // lightPink, because it's strange when too many technical assets process no data... some ok, but many in a diagram ist a sign of model forgery...
	} else if len(what.CommunicationLinks) == 0 && len(IncomingTechnicalCommunicationLinksMappedByTargetId[what.Id]) == 0 {
		fillColor = colors.LightPink
	} else if what.Internet {
		fillColor = colors.ExtremeLightBlue
	} else if what.OutOfScope {
		fillColor = colors.OutOfScopeFancy
	} else if what.CustomDevelopedParts {
		fillColor = colors.CustomDevelopedParts
	}
	switch what.Machine {
	case types.Physical:
		fillColor = colors.DarkenHexColor(fillColor)
	case types.Container:
		fillColor = colors.BrightenHexColor(fillColor)
	case types.Serverless:
		fillColor = colors.BrightenHexColor(colors.BrightenHexColor(fillColor))
	case types.Virtual:
	}
	return fillColor
}

type ByRiskCategoryTitleSort []risks.RiskCategory

func (what ByRiskCategoryTitleSort) Len() int { return len(what) }
func (what ByRiskCategoryTitleSort) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByRiskCategoryTitleSort) Less(i, j int) bool {
	return what[i].Title < what[j].Title
}

type ByRiskCategoryHighestContainingRiskSeveritySortStillAtRisk []RiskCategory

func (what ByRiskCategoryHighestContainingRiskSeveritySortStillAtRisk) Len() int { return len(what) }
func (what ByRiskCategoryHighestContainingRiskSeveritySortStillAtRisk) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByRiskCategoryHighestContainingRiskSeveritySortStillAtRisk) Less(i, j int) bool {
	risksLeft := ReduceToOnlyStillAtRisk(GeneratedRisksByCategory[what[i]])
	risksRight := ReduceToOnlyStillAtRisk(GeneratedRisksByCategory[what[j]])
	highestLeft := HighestSeverityStillAtRisk(risksLeft)
	highestRight := HighestSeverityStillAtRisk(risksRight)
	if highestLeft == highestRight {
		if len(risksLeft) == 0 && len(risksRight) > 0 {
			return false
		}
		if len(risksLeft) > 0 && len(risksRight) == 0 {
			return true
		}
		return what[i].Title < what[j].Title
	}
	return highestLeft > highestRight
}

type RiskStatistics struct {
	// TODO add also some more like before / after (i.e. with mitigation applied)
	Risks map[string]map[string]int `yaml:"risks" json:"risks"`
}

type ByRiskSeveritySort []Risk

func (what ByRiskSeveritySort) Len() int { return len(what) }
func (what ByRiskSeveritySort) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByRiskSeveritySort) Less(i, j int) bool {
	if what[i].Severity == what[j].Severity {
		trackingStatusLeft := what[i].GetRiskTrackingStatusDefaultingUnchecked()
		trackingStatusRight := what[j].GetRiskTrackingStatusDefaultingUnchecked()
		if trackingStatusLeft == trackingStatusRight {
			impactLeft := what[i].ExploitationImpact
			impactRight := what[j].ExploitationImpact
			if impactLeft == impactRight {
				likelihoodLeft := what[i].ExploitationLikelihood
				likelihoodRight := what[j].ExploitationLikelihood
				if likelihoodLeft == likelihoodRight {
					return what[i].Title < what[j].Title
				} else {
					return likelihoodLeft > likelihoodRight
				}
			} else {
				return impactLeft > impactRight
			}
		} else {
			return trackingStatusLeft < trackingStatusRight
		}
	}
	return what[i].Severity > what[j].Severity
}

type ByDataBreachProbabilitySort []Risk

func (what ByDataBreachProbabilitySort) Len() int { return len(what) }
func (what ByDataBreachProbabilitySort) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByDataBreachProbabilitySort) Less(i, j int) bool {
	if what[i].DataBreachProbability == what[j].DataBreachProbability {
		trackingStatusLeft := what[i].GetRiskTrackingStatusDefaultingUnchecked()
		trackingStatusRight := what[j].GetRiskTrackingStatusDefaultingUnchecked()
		if trackingStatusLeft == trackingStatusRight {
			return what[i].Title < what[j].Title
		} else {
			return trackingStatusLeft < trackingStatusRight
		}
	}
	return what[i].DataBreachProbability > what[j].DataBreachProbability
}

type RiskRule interface {
	Category() RiskCategory
	GenerateRisks(parsedModel ParsedModel) []Risk
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedRiskCategories() []RiskCategory {
	categories := make([]RiskCategory, 0)
	for k := range GeneratedRisksByCategory {
		categories = append(categories, k)
	}
	sort.Sort(ByRiskCategoryHighestContainingRiskSeveritySortStillAtRisk(categories))
	return categories
}
func SortedRisksOfCategory(category RiskCategory) []Risk {
	risks := GeneratedRisksByCategory[category]
	sort.Sort(ByRiskSeveritySort(risks))
	return risks
}

func CountRisks(risksByCategory map[RiskCategory][]Risk) int {
	result := 0
	for _, risks := range risksByCategory {
		result += len(risks)
	}
	return result
}

func RisksOfOnlySTRIDESpoofing(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.STRIDE == types.Spoofing {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDETampering(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.STRIDE == types.Tampering {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDERepudiation(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.STRIDE == types.Repudiation {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDEInformationDisclosure(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.STRIDE == types.InformationDisclosure {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDEDenialOfService(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.STRIDE == types.DenialOfService {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDEElevationOfPrivilege(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.STRIDE == types.ElevationOfPrivilege {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyBusinessSide(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.BusinessSide {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyArchitecture(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.Architecture {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyDevelopment(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.Development {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyOperation(risksByCategory map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.Operations {
				result[risk.Category] = append(result[risk.Category], risk)
			}
		}
	}
	return result
}

func CategoriesOfOnlyRisksStillAtRisk(risksByCategory map[RiskCategory][]Risk) []RiskCategory {
	categories := make(map[RiskCategory]struct{}) // Go's trick of unique elements is a map
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if !risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
				continue
			}
			categories[risk.Category] = struct{}{}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyCriticalRisks(risksByCategory map[RiskCategory][]Risk, initialRisks bool) []RiskCategory {
	categories := make(map[RiskCategory]struct{}) // Go's trick of unique elements is a map
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
				continue
			}
			if risk.Severity == types.CriticalSeverity {
				categories[risk.Category] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyHighRisks(risksByCategory map[RiskCategory][]Risk, initialRisks bool) []RiskCategory {
	categories := make(map[RiskCategory]struct{}) // Go's trick of unique elements is a map
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(GeneratedRisksByCategory[risk.Category])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(GeneratedRisksByCategory[risk.Category])
			}
			if risk.Severity == types.HighSeverity && highest < types.CriticalSeverity {
				categories[risk.Category] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyElevatedRisks(risksByCategory map[RiskCategory][]Risk, initialRisks bool) []RiskCategory {
	categories := make(map[RiskCategory]struct{}) // Go's trick of unique elements is a map
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(GeneratedRisksByCategory[risk.Category])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(GeneratedRisksByCategory[risk.Category])
			}
			if risk.Severity == types.ElevatedSeverity && highest < types.HighSeverity {
				categories[risk.Category] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyMediumRisks(risksByCategory map[RiskCategory][]Risk, initialRisks bool) []RiskCategory {
	categories := make(map[RiskCategory]struct{}) // Go's trick of unique elements is a map
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(GeneratedRisksByCategory[risk.Category])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(GeneratedRisksByCategory[risk.Category])
			}
			if risk.Severity == types.MediumSeverity && highest < types.ElevatedSeverity {
				categories[risk.Category] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyLowRisks(risksByCategory map[RiskCategory][]Risk, initialRisks bool) []RiskCategory {
	categories := make(map[RiskCategory]struct{}) // Go's trick of unique elements is a map
	for _, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(GeneratedRisksByCategory[risk.Category])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(GeneratedRisksByCategory[risk.Category])
			}
			if risk.Severity == types.LowSeverity && highest < types.MediumSeverity {
				categories[risk.Category] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func HighestSeverity(risks []Risk) types.RiskSeverity {
	result := types.LowSeverity
	for _, risk := range risks {
		if risk.Severity > result {
			result = risk.Severity
		}
	}
	return result
}

func HighestSeverityStillAtRisk(risks []Risk) types.RiskSeverity {
	result := types.LowSeverity
	for _, risk := range risks {
		if risk.Severity > result && risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
			result = risk.Severity
		}
	}
	return result
}

func keysAsSlice(categories map[RiskCategory]struct{}) []RiskCategory {
	result := make([]RiskCategory, 0, len(categories))
	for k := range categories {
		result = append(result, k)
	}
	return result
}

func FilteredByOnlyBusinessSide() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.BusinessSide {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyArchitecture() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.Architecture {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyDevelopment() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.Development {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyOperation() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Category.Function == types.Operations {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyCriticalRisks() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.CriticalSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyHighRisks() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.HighSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyElevatedRisks() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.ElevatedSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyMediumRisks() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.MediumSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyLowRisks() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.LowSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilterByModelFailures(risksByCat map[RiskCategory][]Risk) map[RiskCategory][]Risk {
	result := make(map[RiskCategory][]Risk)
	for riskCat, risks := range risksByCat {
		if riskCat.ModelFailurePossibleReason {
			result[riskCat] = risks
		}
	}
	return result
}

func FlattenRiskSlice(risksByCat map[RiskCategory][]Risk) []Risk {
	result := make([]Risk, 0)
	for _, risks := range risksByCat {
		result = append(result, risks...)
	}
	return result
}

func TotalRiskCount() int {
	count := 0
	for _, risks := range GeneratedRisksByCategory {
		count += len(risks)
	}
	return count
}

func FilteredByRiskTrackingUnchecked() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.Unchecked {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingInDiscussion() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.InDiscussion {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingAccepted() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.Accepted {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingInProgress() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.InProgress {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingMitigated() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.Mitigated {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingFalsePositive() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.FalsePositive {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func ReduceToOnlyHighRisk(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.Severity == types.HighSeverity {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyMediumRisk(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.Severity == types.MediumSeverity {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyLowRisk(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.Severity == types.LowSeverity {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingUnchecked(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.Unchecked {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingInDiscussion(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.InDiscussion {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingAccepted(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.Accepted {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingInProgress(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.InProgress {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingMitigated(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.Mitigated {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingFalsePositive(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.GetRiskTrackingStatusDefaultingUnchecked() == types.FalsePositive {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func FilteredByStillAtRisk() []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func OverallRiskStatistics() RiskStatistics {
	result := RiskStatistics{}
	result.Risks = make(map[string]map[string]int)
	result.Risks[types.CriticalSeverity.String()] = make(map[string]int)
	result.Risks[types.CriticalSeverity.String()][types.Unchecked.String()] = 0
	result.Risks[types.CriticalSeverity.String()][types.InDiscussion.String()] = 0
	result.Risks[types.CriticalSeverity.String()][types.Accepted.String()] = 0
	result.Risks[types.CriticalSeverity.String()][types.InProgress.String()] = 0
	result.Risks[types.CriticalSeverity.String()][types.Mitigated.String()] = 0
	result.Risks[types.CriticalSeverity.String()][types.FalsePositive.String()] = 0
	result.Risks[types.HighSeverity.String()] = make(map[string]int)
	result.Risks[types.HighSeverity.String()][types.Unchecked.String()] = 0
	result.Risks[types.HighSeverity.String()][types.InDiscussion.String()] = 0
	result.Risks[types.HighSeverity.String()][types.Accepted.String()] = 0
	result.Risks[types.HighSeverity.String()][types.InProgress.String()] = 0
	result.Risks[types.HighSeverity.String()][types.Mitigated.String()] = 0
	result.Risks[types.HighSeverity.String()][types.FalsePositive.String()] = 0
	result.Risks[types.ElevatedSeverity.String()] = make(map[string]int)
	result.Risks[types.ElevatedSeverity.String()][types.Unchecked.String()] = 0
	result.Risks[types.ElevatedSeverity.String()][types.InDiscussion.String()] = 0
	result.Risks[types.ElevatedSeverity.String()][types.Accepted.String()] = 0
	result.Risks[types.ElevatedSeverity.String()][types.InProgress.String()] = 0
	result.Risks[types.ElevatedSeverity.String()][types.Mitigated.String()] = 0
	result.Risks[types.ElevatedSeverity.String()][types.FalsePositive.String()] = 0
	result.Risks[types.MediumSeverity.String()] = make(map[string]int)
	result.Risks[types.MediumSeverity.String()][types.Unchecked.String()] = 0
	result.Risks[types.MediumSeverity.String()][types.InDiscussion.String()] = 0
	result.Risks[types.MediumSeverity.String()][types.Accepted.String()] = 0
	result.Risks[types.MediumSeverity.String()][types.InProgress.String()] = 0
	result.Risks[types.MediumSeverity.String()][types.Mitigated.String()] = 0
	result.Risks[types.MediumSeverity.String()][types.FalsePositive.String()] = 0
	result.Risks[types.LowSeverity.String()] = make(map[string]int)
	result.Risks[types.LowSeverity.String()][types.Unchecked.String()] = 0
	result.Risks[types.LowSeverity.String()][types.InDiscussion.String()] = 0
	result.Risks[types.LowSeverity.String()][types.Accepted.String()] = 0
	result.Risks[types.LowSeverity.String()][types.InProgress.String()] = 0
	result.Risks[types.LowSeverity.String()][types.Mitigated.String()] = 0
	result.Risks[types.LowSeverity.String()][types.FalsePositive.String()] = 0
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			result.Risks[risk.Severity.String()][risk.GetRiskTrackingStatusDefaultingUnchecked().String()]++
		}
	}
	return result
}
