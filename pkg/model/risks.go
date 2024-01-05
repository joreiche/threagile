/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

package model

import (
	"sort"

	"github.com/threagile/threagile/pkg/security/types"
)

// TODO: consider about moving this logic to be part of parsedModel methods or check usage of function and move it inside the consumer

func GetRiskCategory(parsedModel *ParsedModel, categoryID string) *types.RiskCategory {
	if len(parsedModel.IndividualRiskCategories) > 0 {
		custom, customOk := parsedModel.IndividualRiskCategories[categoryID]
		if customOk {
			return &custom
		}
	}

	if len(parsedModel.BuiltInRiskCategories) > 0 {
		builtIn, builtInOk := parsedModel.BuiltInRiskCategories[categoryID]
		if builtInOk {
			return &builtIn
		}
	}

	return nil
}

func GetRiskCategories(parsedModel *ParsedModel, categoryIDs []string) []types.RiskCategory {
	categoryMap := make(map[string]types.RiskCategory)
	for _, categoryId := range categoryIDs {
		category := GetRiskCategory(parsedModel, categoryId)
		if category != nil {
			categoryMap[categoryId] = *category
		}
	}

	categories := make([]types.RiskCategory, 0)
	for categoryId := range categoryMap {
		categories = append(categories, categoryMap[categoryId])
	}

	return categories
}

func AllRisks(parsedModel *ParsedModel) []types.Risk {
	result := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			result = append(result, risk)
		}
	}
	return result
}

func ReduceToOnlyStillAtRisk(parsedModel *ParsedModel, risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func HighestSeverityStillAtRisk(model *ParsedModel, risks []types.Risk) types.RiskSeverity {
	result := types.LowSeverity
	for _, risk := range risks {
		if risk.Severity > result && GetRiskTrackingStatusDefaultingUnchecked(model, risk).IsStillAtRisk() {
			result = risk.Severity
		}
	}
	return result
}

func SortByRiskCategoryHighestContainingRiskSeveritySortStillAtRisk(parsedModel *ParsedModel, riskCategories []types.RiskCategory) {
	sort.Slice(riskCategories, func(i, j int) bool {
		risksLeft := ReduceToOnlyStillAtRisk(parsedModel, parsedModel.GeneratedRisksByCategory[riskCategories[i].Id])
		risksRight := ReduceToOnlyStillAtRisk(parsedModel, parsedModel.GeneratedRisksByCategory[riskCategories[j].Id])
		highestLeft := HighestSeverityStillAtRisk(parsedModel, risksLeft)
		highestRight := HighestSeverityStillAtRisk(parsedModel, risksRight)
		if highestLeft == highestRight {
			if len(risksLeft) == 0 && len(risksRight) > 0 {
				return false
			}
			if len(risksLeft) > 0 && len(risksRight) == 0 {
				return true
			}
			return riskCategories[i].Title < riskCategories[j].Title
		}
		return highestLeft > highestRight
	})
}

func SortByRiskSeverity(risks []types.Risk, parsedModel *ParsedModel) {
	sort.Slice(risks, func(i, j int) bool {
		if risks[i].Severity == risks[j].Severity {
			trackingStatusLeft := GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risks[i])
			trackingStatusRight := GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risks[j])
			if trackingStatusLeft == trackingStatusRight {
				impactLeft := risks[i].ExploitationImpact
				impactRight := risks[j].ExploitationImpact
				if impactLeft == impactRight {
					likelihoodLeft := risks[i].ExploitationLikelihood
					likelihoodRight := risks[j].ExploitationLikelihood
					if likelihoodLeft == likelihoodRight {
						return risks[i].Title < risks[j].Title
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
		return risks[i].Severity > risks[j].Severity

	})
}

func SortByDataBreachProbability(risks []types.Risk, parsedModel *ParsedModel) {
	sort.Slice(risks, func(i, j int) bool {

		if risks[i].DataBreachProbability == risks[j].DataBreachProbability {
			trackingStatusLeft := GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risks[i])
			trackingStatusRight := GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risks[j])
			if trackingStatusLeft == trackingStatusRight {
				return risks[i].Title < risks[j].Title
			} else {
				return trackingStatusLeft < trackingStatusRight
			}
		}
		return risks[i].DataBreachProbability > risks[j].DataBreachProbability
	})
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:

func SortedRiskCategories(parsedModel *ParsedModel) []types.RiskCategory {
	categoryMap := make(map[string]types.RiskCategory)
	for categoryId := range parsedModel.GeneratedRisksByCategory {
		category := GetRiskCategory(parsedModel, categoryId)
		if category != nil {
			categoryMap[categoryId] = *category
		}
	}

	categories := make([]types.RiskCategory, 0)
	for categoryId := range categoryMap {
		categories = append(categories, categoryMap[categoryId])
	}

	SortByRiskCategoryHighestContainingRiskSeveritySortStillAtRisk(parsedModel, categories)
	return categories
}

func SortedRisksOfCategory(parsedModel *ParsedModel, category types.RiskCategory) []types.Risk {
	risks := parsedModel.GeneratedRisksByCategory[category.Id]
	SortByRiskSeverity(risks, parsedModel)
	return risks
}

func CountRisks(risksByCategory map[string][]types.Risk) int {
	result := 0
	for _, risks := range risksByCategory {
		result += len(risks)
	}
	return result
}

func RisksOfOnlySTRIDESpoofing(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category != nil {
				if category.STRIDE == types.Spoofing {
					result[categoryId] = append(result[categoryId], risk)
				}
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDETampering(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category != nil {
				if category.STRIDE == types.Tampering {
					result[categoryId] = append(result[categoryId], risk)
				}
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDERepudiation(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.STRIDE == types.Repudiation {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDEInformationDisclosure(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.STRIDE == types.InformationDisclosure {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDEDenialOfService(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.STRIDE == types.DenialOfService {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func RisksOfOnlySTRIDEElevationOfPrivilege(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.STRIDE == types.ElevationOfPrivilege {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyBusinessSide(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.BusinessSide {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyArchitecture(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Architecture {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyDevelopment(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Development {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func RisksOfOnlyOperation(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Operations {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func CategoriesOfOnlyRisksStillAtRisk(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk) []string {
	categories := make(map[string]struct{}) // Go's trick of unique elements is a map
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			if !GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
				continue
			}
			categories[categoryId] = struct{}{}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyCriticalRisks(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk, initialRisks bool) []string {
	categories := make(map[string]struct{}) // Go's trick of unique elements is a map
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
				continue
			}
			if risk.Severity == types.CriticalSeverity {
				categories[categoryId] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyHighRisks(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk, initialRisks bool) []string {
	categories := make(map[string]struct{}) // Go's trick of unique elements is a map
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(parsedModel.GeneratedRisksByCategory[categoryId])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(parsedModel, parsedModel.GeneratedRisksByCategory[categoryId])
			}
			if risk.Severity == types.HighSeverity && highest < types.CriticalSeverity {
				categories[categoryId] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyElevatedRisks(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk, initialRisks bool) []string {
	categories := make(map[string]struct{}) // Go's trick of unique elements is a map
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(parsedModel.GeneratedRisksByCategory[categoryId])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(parsedModel, parsedModel.GeneratedRisksByCategory[categoryId])
			}
			if risk.Severity == types.ElevatedSeverity && highest < types.HighSeverity {
				categories[categoryId] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyMediumRisks(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk, initialRisks bool) []string {
	categories := make(map[string]struct{}) // Go's trick of unique elements is a map
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(parsedModel.GeneratedRisksByCategory[categoryId])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(parsedModel, parsedModel.GeneratedRisksByCategory[categoryId])
			}
			if risk.Severity == types.MediumSeverity && highest < types.ElevatedSeverity {
				categories[categoryId] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func CategoriesOfOnlyLowRisks(parsedModel *ParsedModel, risksByCategory map[string][]types.Risk, initialRisks bool) []string {
	categories := make(map[string]struct{}) // Go's trick of unique elements is a map
	for categoryId, risks := range risksByCategory {
		for _, risk := range risks {
			if !initialRisks && !GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
				continue
			}
			highest := HighestSeverity(parsedModel.GeneratedRisksByCategory[categoryId])
			if !initialRisks {
				highest = HighestSeverityStillAtRisk(parsedModel, parsedModel.GeneratedRisksByCategory[categoryId])
			}
			if risk.Severity == types.LowSeverity && highest < types.MediumSeverity {
				categories[categoryId] = struct{}{}
			}
		}
	}
	// return as slice (of now unique values)
	return keysAsSlice(categories)
}

func HighestSeverity(risks []types.Risk) types.RiskSeverity {
	result := types.LowSeverity
	for _, risk := range risks {
		if risk.Severity > result {
			result = risk.Severity
		}
	}
	return result
}

func keysAsSlice(categories map[string]struct{}) []string {
	result := make([]string, 0, len(categories))
	for k := range categories {
		result = append(result, k)
	}
	return result
}

func FilteredByOnlyBusinessSide(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for categoryId, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.BusinessSide {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyArchitecture(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for categoryId, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Architecture {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyDevelopment(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for categoryId, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Development {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyOperation(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for categoryId, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Operations {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyCriticalRisks(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.CriticalSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyHighRisks(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.HighSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyElevatedRisks(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.ElevatedSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyMediumRisks(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.MediumSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByOnlyLowRisks(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if risk.Severity == types.LowSeverity {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilterByModelFailures(parsedModel *ParsedModel, risksByCat map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, risks := range risksByCat {
		category := GetRiskCategory(parsedModel, categoryId)
		if category.ModelFailurePossibleReason {
			result[categoryId] = risks
		}
	}

	return result
}

func FlattenRiskSlice(risksByCat map[string][]types.Risk) []types.Risk {
	result := make([]types.Risk, 0)
	for _, risks := range risksByCat {
		result = append(result, risks...)
	}
	return result
}

func TotalRiskCount(parsedModel *ParsedModel) int {
	count := 0
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		count += len(risks)
	}
	return count
}

func FilteredByRiskTrackingUnchecked(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.Unchecked {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingInDiscussion(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.InDiscussion {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingAccepted(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.Accepted {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingInProgress(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.InProgress {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingMitigated(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.Mitigated {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func FilteredByRiskTrackingFalsePositive(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.FalsePositive {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}

func ReduceToOnlyHighRisk(risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if risk.Severity == types.HighSeverity {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyMediumRisk(risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if risk.Severity == types.MediumSeverity {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyLowRisk(risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if risk.Severity == types.LowSeverity {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingUnchecked(parsedModel *ParsedModel, risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.Unchecked {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingInDiscussion(parsedModel *ParsedModel, risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.InDiscussion {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingAccepted(parsedModel *ParsedModel, risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.Accepted {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingInProgress(parsedModel *ParsedModel, risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.InProgress {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingMitigated(parsedModel *ParsedModel, risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.Mitigated {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func ReduceToOnlyRiskTrackingFalsePositive(parsedModel *ParsedModel, risks []types.Risk) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risk := range risks {
		if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk) == types.FalsePositive {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func FilteredByStillAtRisk(parsedModel *ParsedModel) []types.Risk {
	filteredRisks := make([]types.Risk, 0)
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			if GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).IsStillAtRisk() {
				filteredRisks = append(filteredRisks, risk)
			}
		}
	}
	return filteredRisks
}
