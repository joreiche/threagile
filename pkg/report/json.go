package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/threagile/threagile/pkg/model"
	"github.com/threagile/threagile/pkg/security/types"
)

func WriteRisksJSON(parsedModel *model.ParsedModel, filename string) error {
	/*
		remainingRisks := make([]model.Risk, 0)
		for _, category := range model.SortedRiskCategories() {
			risks := model.SortedRisksOfCategory(category)
			for _, risk := range model.ReduceToOnlyStillAtRisk(risks) {
				remainingRisks = append(remainingRisks, risk)
			}
		}
	*/
	jsonBytes, err := json.Marshal(model.AllRisks(parsedModel))
	if err != nil {
		return fmt.Errorf("failed to marshal risks to JSON: %w", err)
	}
	err = os.WriteFile(filename, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write risks to JSON file: %w", err)
	}
	return nil
}

// TODO: also a "data assets" json?

func WriteTechnicalAssetsJSON(parsedModel *model.ParsedModel, filename string) error {
	jsonBytes, err := json.Marshal(parsedModel.TechnicalAssets)
	if err != nil {
		return fmt.Errorf("failed to marshal technical assets to JSON: %w", err)
	}
	err = os.WriteFile(filename, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write technical assets to JSON file: %w", err)
	}
	return nil
}

func WriteStatsJSON(parsedModel *model.ParsedModel, filename string) error {
	jsonBytes, err := json.Marshal(overallRiskStatistics(parsedModel))
	if err != nil {
		return fmt.Errorf("failed to marshal stats to JSON: %w", err)
	}
	err = os.WriteFile(filename, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write stats to JSON file: %w", err)
	}
	return nil
}

type riskStatistics struct {
	// TODO add also some more like before / after (i.e. with mitigation applied)
	Risks map[string]map[string]int `yaml:"risks" json:"risks"`
}

func overallRiskStatistics(parsedModel *model.ParsedModel) riskStatistics {
	result := riskStatistics{}
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
	for _, risks := range parsedModel.GeneratedRisksByCategory {
		for _, risk := range risks {
			result.Risks[risk.Severity.String()][model.GetRiskTrackingStatusDefaultingUnchecked(parsedModel, risk).String()]++
		}
	}
	return result
}
