/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"github.com/threagile/threagile/pkg/security/types"
)

type TrustBoundary struct {
	Id, Title, Description string
	Type                   types.TrustBoundaryType
	Tags                   []string
	TechnicalAssetsInside  []string
	TrustBoundariesNested  []string
}

func (what TrustBoundary) RecursivelyAllTechnicalAssetIDsInside() []string {
	result := make([]string, 0)
	what.addAssetIDsRecursively(&result)
	return result
}

func (what TrustBoundary) IsTaggedWithAny(tags ...string) bool {
	return containsCaseInsensitiveAny(what.Tags, tags...)
}

func (what TrustBoundary) IsTaggedWithBaseTag(baseTag string) bool {
	return isTaggedWithBaseTag(what.Tags, baseTag)
}

func (what TrustBoundary) IsTaggedWithAnyTraversingUp(tags ...string) bool {
	if what.IsTaggedWithAny(tags...) {
		return true
	}
	parentID := what.ParentTrustBoundaryID()
	if len(parentID) > 0 && ParsedModelRoot.TrustBoundaries[parentID].IsTaggedWithAnyTraversingUp(tags...) {
		return true
	}
	return false
}

func (what TrustBoundary) ParentTrustBoundaryID() string {
	var result string
	for _, candidate := range ParsedModelRoot.TrustBoundaries {
		if Contains(candidate.TrustBoundariesNested, what.Id) {
			result = candidate.Id
			return result
		}
	}
	return result
}

func (what TrustBoundary) HighestConfidentiality() types.Confidentiality {
	highest := types.Public
	for _, id := range what.RecursivelyAllTechnicalAssetIDsInside() {
		techAsset := ParsedModelRoot.TechnicalAssets[id]
		if techAsset.HighestConfidentiality() > highest {
			highest = techAsset.HighestConfidentiality()
		}
	}
	return highest
}

func (what TrustBoundary) HighestIntegrity() types.Criticality {
	highest := types.Archive
	for _, id := range what.RecursivelyAllTechnicalAssetIDsInside() {
		techAsset := ParsedModelRoot.TechnicalAssets[id]
		if techAsset.HighestIntegrity() > highest {
			highest = techAsset.HighestIntegrity()
		}
	}
	return highest
}

func (what TrustBoundary) HighestAvailability() types.Criticality {
	highest := types.Archive
	for _, id := range what.RecursivelyAllTechnicalAssetIDsInside() {
		techAsset := ParsedModelRoot.TechnicalAssets[id]
		if techAsset.HighestAvailability() > highest {
			highest = techAsset.HighestAvailability()
		}
	}
	return highest
}

func (what TrustBoundary) AllParentTrustBoundaryIDs() []string {
	result := make([]string, 0)
	what.addTrustBoundaryIDsRecursively(&result)
	return result
}


func (what TrustBoundary) addAssetIDsRecursively(model ParsedModel, result *[]string) {
	*result = append(*result, what.TechnicalAssetsInside...)
	for _, nestedBoundaryID := range what.TrustBoundariesNested {
		model.TrustBoundaries[nestedBoundaryID].addAssetIDsRecursively(result)
	}
}

// TODO: pass ParsedModelRoot as parameter instead of using global variable
func (what TrustBoundary) addTrustBoundaryIDsRecursively(model ParsedModel,  *[]string) {
	*result = append(*result, what.Id)
	parentID := what.ParentTrustBoundaryID()
	if len(parentID) > 0 {
		model.TrustBoundaries[parentID].addTrustBoundaryIDsRecursively(result)
	}
}
