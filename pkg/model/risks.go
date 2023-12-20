/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"log"
	"time"

	"github.com/threagile/threagile/pkg/internal"
	"github.com/threagile/threagile/pkg/security/types"
)

type RiskCategory struct {
	// TODO: refactor all "Id" here and elsewhere to "ID"
	Id                         string
	Title                      string
	Description                string
	Impact                     string
	ASVS                       string
	CheatSheet                 string
	Action                     string
	Mitigation                 string
	Check                      string
	DetectionLogic             string
	RiskAssessment             string
	FalsePositives             string
	Function                   types.RiskFunction
	STRIDE                     types.STRIDE
	ModelFailurePossibleReason bool
	CWE                        int
}

type BuiltInRisk struct {
	Category      func() RiskCategory
	SupportedTags func() []string
	GenerateRisks func(m *ParsedModel) []Risk
}

type CustomRisk struct {
	ID       string
	Category RiskCategory
	Tags     []string
	Runner   *internal.Runner
}

func (r *CustomRisk) GenerateRisks(m *ParsedModel) []Risk {
	if r.Runner == nil {
		return nil
	}

	risks := make([]Risk, 0)
	runError := r.Runner.Run(m, &risks, "-generate-risks")
	if runError != nil {
		log.Fatalf("Failed to generate risks for custom risk rule %q: %v\n", r.Runner.Filename, runError)
	}

	return risks
}

type RiskTracking struct {
	SyntheticRiskId, Justification, Ticket, CheckedBy string
	Status                                            types.RiskStatus
	Date                                              time.Time
}

type Risk struct {
	Category                        RiskCategory                     `yaml:"-" json:"-"`                     // just for navigational convenience... not JSON marshalled
	CategoryId                      string                           `yaml:"category" json:"category"`       // used for better JSON marshalling, is assigned in risk evaluation phase automatically
	RiskStatus                      types.RiskStatus                 `yaml:"risk_status" json:"risk_status"` // used for better JSON marshalling, is assigned in risk evaluation phase automatically
	Severity                        types.RiskSeverity               `yaml:"severity" json:"severity"`
	ExploitationLikelihood          types.RiskExploitationLikelihood `yaml:"exploitation_likelihood" json:"exploitation_likelihood"`
	ExploitationImpact              types.RiskExploitationImpact     `yaml:"exploitation_impact" json:"exploitation_impact"`
	Title                           string                           `yaml:"title" json:"title"`
	SyntheticId                     string                           `yaml:"synthetic_id" json:"synthetic_id"`
	MostRelevantDataAssetId         string                           `yaml:"most_relevant_data_asset" json:"most_relevant_data_asset"`
	MostRelevantTechnicalAssetId    string                           `yaml:"most_relevant_technical_asset" json:"most_relevant_technical_asset"`
	MostRelevantTrustBoundaryId     string                           `yaml:"most_relevant_trust_boundary" json:"most_relevant_trust_boundary"`
	MostRelevantSharedRuntimeId     string                           `yaml:"most_relevant_shared_runtime" json:"most_relevant_shared_runtime"`
	MostRelevantCommunicationLinkId string                           `yaml:"most_relevant_communication_link" json:"most_relevant_communication_link"`
	DataBreachProbability           types.DataBreachProbability      `yaml:"data_breach_probability" json:"data_breach_probability"`
	DataBreachTechnicalAssetIDs     []string                         `yaml:"data_breach_technical_assets" json:"data_breach_technical_assets"`
	// TODO: refactor all "Id" here to "ID"?
}

func (what Risk) GetRiskTracking(model ParsedModel) RiskTracking { // TODO: Unify function naming regarding Get etc.
	var result RiskTracking
	if riskTracking, ok := model.RiskTracking[what.SyntheticId]; ok {
		result = riskTracking
	}
	return result
}

func (what Risk) GetRiskTrackingStatusDefaultingUnchecked(model ParsedModel) types.RiskStatus {
	if riskTracking, ok := model.RiskTracking[what.SyntheticId]; ok {
		return riskTracking.Status
	}
	return types.Unchecked
}

func (what Risk) IsRiskTracked(model ParsedModel) bool {
	if _, ok := model.RiskTracking[what.SyntheticId]; ok {
		return true
	}
	return false
}

func AllRisks() []Risk {
	result := make([]Risk, 0)
	for _, risks := range GeneratedRisksByCategory {
		for _, risk := range risks {
			result = append(result, risk)
		}
	}
	return result
}

func ReduceToOnlyStillAtRisk(risks []Risk) []Risk {
	filteredRisks := make([]Risk, 0)
	for _, risk := range risks {
		if risk.GetRiskTrackingStatusDefaultingUnchecked().IsStillAtRisk() {
			filteredRisks = append(filteredRisks, risk)
		}
	}
	return filteredRisks
}

func HighestExploitationLikelihood(risks []Risk) types.RiskExploitationLikelihood {
	result := types.Unlikely
	for _, risk := range risks {
		if risk.ExploitationLikelihood > result {
			result = risk.ExploitationLikelihood
		}
	}
	return result
}

func HighestExploitationImpact(risks []Risk) types.RiskExploitationImpact {
	result := types.LowImpact
	for _, risk := range risks {
		if risk.ExploitationImpact > result {
			result = risk.ExploitationImpact
		}
	}
	return result
}

type CustomRiskRule struct {
	Category      func() RiskCategory
	SupportedTags func() []string
	GenerateRisks func(input *ParsedModel) []Risk
}
