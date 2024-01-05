package model

import (
	"sort"

	"github.com/threagile/threagile/pkg/security/types"
)

func RecursivelyAllTechnicalAssetIDsInside(model *ParsedModel, trustBoundary types.TrustBoundary) []string {
	result := make([]string, 0)
	addAssetIDsRecursively(model, &result, trustBoundary)
	return result
}

func IsTaggedWithAnyTraversingUp(model *ParsedModel, trustBoundary types.TrustBoundary, tags ...string) bool {
	if trustBoundary.IsTaggedWithAny(tags...) {
		return true
	}
	parentID := ParentTrustBoundaryID(model, trustBoundary)
	if len(parentID) > 0 && IsTaggedWithAnyTraversingUp(model, model.TrustBoundaries[parentID], tags...) {
		return true
	}
	return false
}

func ParentTrustBoundaryID(model *ParsedModel, trustBoundary types.TrustBoundary) string {
	var result string
	for _, candidate := range model.TrustBoundaries {
		if contains(candidate.TrustBoundariesNested, trustBoundary.Id) {
			result = candidate.Id
			return result
		}
	}
	return result
}

func HighestConfidentiality(model *ParsedModel, trustBoundary types.TrustBoundary) types.Confidentiality {
	highest := types.Public
	for _, id := range RecursivelyAllTechnicalAssetIDsInside(model, trustBoundary) {
		techAsset := model.TechnicalAssets[id]
		if techAsset.HighestConfidentiality(model) > highest {
			highest = techAsset.HighestConfidentiality(model)
		}
	}
	return highest
}

func HighestIntegrity(model *ParsedModel, trustBoundary types.TrustBoundary) types.Criticality {
	highest := types.Archive
	for _, id := range RecursivelyAllTechnicalAssetIDsInside(model, trustBoundary) {
		techAsset := model.TechnicalAssets[id]
		if techAsset.HighestIntegrity(model) > highest {
			highest = techAsset.HighestIntegrity(model)
		}
	}
	return highest
}

func HighestAvailability(model *ParsedModel, trustBoundary types.TrustBoundary) types.Criticality {
	highest := types.Archive
	for _, id := range RecursivelyAllTechnicalAssetIDsInside(model, trustBoundary) {
		techAsset := model.TechnicalAssets[id]
		if techAsset.HighestAvailability(model) > highest {
			highest = techAsset.HighestAvailability(model)
		}
	}
	return highest
}

func AllParentTrustBoundaryIDs(model *ParsedModel, trustBoundary types.TrustBoundary) []string {
	result := make([]string, 0)
	addTrustBoundaryIDsRecursively(model, &result, trustBoundary)
	return result
}

func addAssetIDsRecursively(model *ParsedModel, result *[]string, trustBoundary types.TrustBoundary) {
	*result = append(*result, trustBoundary.TechnicalAssetsInside...)
	for _, nestedBoundaryID := range trustBoundary.TrustBoundariesNested {
		addAssetIDsRecursively(model, result, model.TrustBoundaries[nestedBoundaryID])
	}
}

func addTrustBoundaryIDsRecursively(model *ParsedModel, result *[]string, trustBoundary types.TrustBoundary) {
	*result = append(*result, trustBoundary.Id)
	parentID := ParentTrustBoundaryID(model, trustBoundary)
	if len(parentID) > 0 {
		addTrustBoundaryIDsRecursively(model, result, model.TrustBoundaries[parentID])
	}
}

// as in Go ranging over map is random order, range over them in sorted (hence reproducible) way:
func SortedKeysOfTrustBoundaries(model *ParsedModel) []string {
	keys := make([]string, 0)
	for k := range model.TrustBoundaries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
