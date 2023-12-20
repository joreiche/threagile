/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"github.com/threagile/threagile/pkg/security/types"
)

type SharedRuntime struct {
	Id, Title, Description string
	Tags                   []string
	TechnicalAssetsRunning []string
}

func (what SharedRuntime) IsTaggedWithAny(tags ...string) bool {
	return containsCaseInsensitiveAny(what.Tags, tags...)
}

func (what SharedRuntime) IsTaggedWithBaseTag(baseTag string) bool {
	return isTaggedWithBaseTag(what.Tags, baseTag)
}

func (what SharedRuntime) HighestConfidentiality() types.Confidentiality {
	highest := types.Public
	for _, id := range what.TechnicalAssetsRunning {
		techAsset := ParsedModelRoot.TechnicalAssets[id]
		if techAsset.HighestConfidentiality() > highest {
			highest = techAsset.HighestConfidentiality()
		}
	}
	return highest
}

func (what SharedRuntime) HighestIntegrity() types.Criticality {
	highest := types.Archive
	for _, id := range what.TechnicalAssetsRunning {
		techAsset := ParsedModelRoot.TechnicalAssets[id]
		if techAsset.HighestIntegrity() > highest {
			highest = techAsset.HighestIntegrity()
		}
	}
	return highest
}

func (what SharedRuntime) HighestAvailability() types.Criticality {
	highest := types.Archive
	for _, id := range what.TechnicalAssetsRunning {
		techAsset := ParsedModelRoot.TechnicalAssets[id]
		if techAsset.HighestAvailability() > highest {
			highest = techAsset.HighestAvailability()
		}
	}
	return highest
}

func (what SharedRuntime) TechnicalAssetWithHighestRAA() TechnicalAsset {
	result := ParsedModelRoot.TechnicalAssets[what.TechnicalAssetsRunning[0]]
	for _, asset := range what.TechnicalAssetsRunning {
		candidate := ParsedModelRoot.TechnicalAssets[asset]
		if candidate.RAA > result.RAA {
			result = candidate
		}
	}
	return result
}
