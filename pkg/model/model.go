/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"time"

	"github.com/threagile/threagile/pkg/input"
	"github.com/threagile/threagile/pkg/security/types"
)

type ParsedModel struct {
	Author                                        input.Author
	Title                                         string
	Date                                          time.Time
	ManagementSummaryComment                      string
	BusinessOverview                              input.Overview
	TechnicalOverview                             input.Overview
	BusinessCriticality                           types.Criticality
	SecurityRequirements                          map[string]string
	Questions                                     map[string]string
	AbuseCases                                    map[string]string
	TagsAvailable                                 []string
	DataAssets                                    map[string]DataAsset
	TechnicalAssets                               map[string]TechnicalAsset
	TrustBoundaries                               map[string]TrustBoundary
	SharedRuntimes                                map[string]SharedRuntime
	IndividualRiskCategories                      map[string]RiskCategory
	RiskTracking                                  map[string]RiskTracking
	DiagramTweakNodesep, DiagramTweakRanksep      int
	DiagramTweakEdgeLayout                        string
	DiagramTweakSuppressEdgeLabels                bool
	DiagramTweakLayoutLeftToRight                 bool
	DiagramTweakInvisibleConnectionsBetweenAssets []string
	DiagramTweakSameRankAssets                    []string
}

func CalculateSeverity(likelihood types.RiskExploitationLikelihood, impact types.RiskExploitationImpact) types.RiskSeverity {
	result := likelihood.Weight() * impact.Weight()
	if result <= 1 {
		return types.LowSeverity
	}
	if result <= 3 {
		return types.MediumSeverity
	}
	if result <= 8 {
		return types.ElevatedSeverity
	}
	if result <= 12 {
		return types.HighSeverity
	}
	return types.CriticalSeverity
}

func (model ParsedModel) InScopeTechnicalAssets() []TechnicalAsset {
	result := make([]TechnicalAsset, 0)
	for _, asset := range model.TechnicalAssets {
		if !asset.OutOfScope {
			result = append(result, asset)
		}
	}
	return result
}
