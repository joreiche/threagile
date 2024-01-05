/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

package model

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/threagile/threagile/pkg/input"
	"github.com/threagile/threagile/pkg/security/types"
)

// TODO: move model out of types package and
// rename parsedModel to model or something like this to emphasize that it's just a model
// maybe
type ParsedModel struct {
	Author                                        input.Author                       `json:"author" yaml:"author"`
	Title                                         string                             `json:"title,omitempty" yaml:"title"`
	Date                                          time.Time                          `json:"date" yaml:"date"`
	ManagementSummaryComment                      string                             `json:"management_summary_comment,omitempty" yaml:"management_summary_comment"`
	BusinessOverview                              input.Overview                     `json:"business_overview" yaml:"business_overview"`
	TechnicalOverview                             input.Overview                     `json:"technical_overview" yaml:"technical_overview"`
	BusinessCriticality                           types.Criticality                  `json:"business_criticality,omitempty" yaml:"business_criticality"`
	SecurityRequirements                          map[string]string                  `json:"security_requirements,omitempty" yaml:"security_requirements"`
	Questions                                     map[string]string                  `json:"questions,omitempty" yaml:"questions"`
	AbuseCases                                    map[string]string                  `json:"abuse_cases,omitempty" yaml:"abuse_cases"`
	TagsAvailable                                 []string                           `json:"tags_available,omitempty" yaml:"tags_available"`
	DataAssets                                    map[string]types.DataAsset         `json:"data_assets,omitempty" yaml:"data_assets"`
	TechnicalAssets                               map[string]types.TechnicalAsset    `json:"technical_assets,omitempty" yaml:"technical_assets"`
	TrustBoundaries                               map[string]types.TrustBoundary     `json:"trust_boundaries,omitempty" yaml:"trust_boundaries"`
	SharedRuntimes                                map[string]types.SharedRuntime     `json:"shared_runtimes,omitempty" yaml:"shared_runtimes"`
	IndividualRiskCategories                      map[string]types.RiskCategory      `json:"individual_risk_categories,omitempty" yaml:"individual_risk_categories"`
	BuiltInRiskCategories                         map[string]types.RiskCategory      `json:"built_in_risk_categories,omitempty" yaml:"built_in_risk_categories"`
	RiskTracking                                  map[string]types.RiskTracking      `json:"risk_tracking,omitempty" yaml:"risk_tracking"`
	CommunicationLinks                            map[string]types.CommunicationLink `json:"communication_links,omitempty" yaml:"communication_links"`
	AllSupportedTags                              map[string]bool                    `json:"all_supported_tags,omitempty" yaml:"all_supported_tags"`
	DiagramTweakNodesep                           int                                `json:"diagram_tweak_nodesep,omitempty" yaml:"diagram_tweak_nodesep"`
	DiagramTweakRanksep                           int                                `json:"diagram_tweak_ranksep,omitempty" yaml:"diagram_tweak_ranksep"`
	DiagramTweakEdgeLayout                        string                             `json:"diagram_tweak_edge_layout,omitempty" yaml:"diagram_tweak_edge_layout"`
	DiagramTweakSuppressEdgeLabels                bool                               `json:"diagram_tweak_suppress_edge_labels,omitempty" yaml:"diagram_tweak_suppress_edge_labels"`
	DiagramTweakLayoutLeftToRight                 bool                               `json:"diagram_tweak_layout_left_to_right,omitempty" yaml:"diagram_tweak_layout_left_to_right"`
	DiagramTweakInvisibleConnectionsBetweenAssets []string                           `json:"diagram_tweak_invisible_connections_between_assets,omitempty" yaml:"diagram_tweak_invisible_connections_between_assets"`
	DiagramTweakSameRankAssets                    []string                           `json:"diagram_tweak_same_rank_assets,omitempty" yaml:"diagram_tweak_same_rank_assets"`

	// TODO: those are generated based on items above and needs to be private
	IncomingTechnicalCommunicationLinksMappedByTargetId   map[string][]types.CommunicationLink `json:"incoming_technical_communication_links_mapped_by_target_id,omitempty" yaml:"incoming_technical_communication_links_mapped_by_target_id"`
	DirectContainingTrustBoundaryMappedByTechnicalAssetId map[string]types.TrustBoundary       `json:"direct_containing_trust_boundary_mapped_by_technical_asset_id,omitempty" yaml:"direct_containing_trust_boundary_mapped_by_technical_asset_id"`
	GeneratedRisksByCategory                              map[string][]types.Risk              `json:"generated_risks_by_category,omitempty" yaml:"generated_risks_by_category"`
	GeneratedRisksBySyntheticId                           map[string]types.Risk                `json:"generated_risks_by_synthetic_id,omitempty" yaml:"generated_risks_by_synthetic_id"`
}

func (parsedModel *ParsedModel) AddToListOfSupportedTags(tags []string) {
	for _, tag := range tags {
		parsedModel.AllSupportedTags[tag] = true
	}
}

func (parsedModel *ParsedModel) GetDeferredRiskTrackingDueToWildcardMatching() map[string]types.RiskTracking {
	deferredRiskTrackingDueToWildcardMatching := make(map[string]types.RiskTracking)
	for syntheticRiskId, riskTracking := range parsedModel.RiskTracking {
		if strings.Contains(syntheticRiskId, "*") { // contains a wildcard char
			deferredRiskTrackingDueToWildcardMatching[syntheticRiskId] = riskTracking
		}
	}

	return deferredRiskTrackingDueToWildcardMatching
}

func (parsedModel *ParsedModel) HasNotYetAnyDirectNonWildcardRiskTracking(syntheticRiskId string) bool {
	if _, ok := parsedModel.RiskTracking[syntheticRiskId]; ok {
		return false
	}
	return true
}

func (parsedModel *ParsedModel) ApplyRisk(rule RiskRule, skippedRules *map[string]bool) {
	id := rule.Category().Id
	_, ok := (*skippedRules)[id]

	if ok {
		fmt.Printf("Skipping risk rule %q\n", rule.Category().Id)
		delete(*skippedRules, rule.Category().Id)
	} else {
		parsedModel.AddToListOfSupportedTags(rule.SupportedTags())
		generatedRisks := rule.GenerateRisks(parsedModel)
		if generatedRisks != nil {
			if len(generatedRisks) > 0 {
				parsedModel.GeneratedRisksByCategory[rule.Category().Id] = generatedRisks
			}
		} else {
			fmt.Printf("Failed to generate risks for %q\n", id)
		}
	}
}

func (parsedModel *ParsedModel) CheckTags(tags []string, where string) ([]string, error) {
	var tagsUsed = make([]string, 0)
	if tags != nil {
		tagsUsed = make([]string, len(tags))
		for i, parsedEntry := range tags {
			referencedTag := fmt.Sprintf("%v", parsedEntry)
			err := parsedModel.CheckTagExists(referencedTag, where)
			if err != nil {
				return nil, err
			}
			tagsUsed[i] = referencedTag
		}
	}
	return tagsUsed, nil
}

// TODO: refactor skipRiskRules to be a string array instead of a comma-separated string
func (parsedModel *ParsedModel) ApplyRiskGeneration(customRiskRules map[string]*types.CustomRisk,
	builtinRiskRules map[string]RiskRule,
	skipRiskRules string,
	progressReporter progressReporter) {
	progressReporter.Info("Applying risk generation")

	skippedRules := make(map[string]bool)
	if len(skipRiskRules) > 0 {
		for _, id := range strings.Split(skipRiskRules, ",") {
			skippedRules[id] = true
		}
	}

	for _, rule := range builtinRiskRules {
		parsedModel.ApplyRisk(rule, &skippedRules)
	}

	// NOW THE CUSTOM RISK RULES (if any)
	for id, customRule := range customRiskRules {
		_, ok := skippedRules[id]
		if ok {
			progressReporter.Info("Skipping custom risk rule:", id)
			delete(skippedRules, id)
		} else {
			progressReporter.Info("Executing custom risk rule:", id)
			parsedModel.AddToListOfSupportedTags(customRule.Tags)
			customRisks := customRule.GenerateRisks(parsedModel)
			if len(customRisks) > 0 {
				parsedModel.GeneratedRisksByCategory[customRule.Category.Id] = customRisks
			}

			progressReporter.Info("Added custom risks:", len(customRisks))
		}
	}

	if len(skippedRules) > 0 {
		keys := make([]string, 0)
		for k := range skippedRules {
			keys = append(keys, k)
		}
		if len(keys) > 0 {
			progressReporter.Info("Unknown risk rules to skip:", keys)
		}
	}

	// save also in map keyed by synthetic risk-id
	for _, category := range SortedRiskCategories(parsedModel) {
		someRisks := SortedRisksOfCategory(parsedModel, category)
		for _, risk := range someRisks {
			parsedModel.GeneratedRisksBySyntheticId[strings.ToLower(risk.SyntheticId)] = risk
		}
	}
}

func (parsedModel *ParsedModel) ApplyWildcardRiskTrackingEvaluation(ignoreOrphanedRiskTracking bool, progressReporter progressReporter) error {
	progressReporter.Info("Executing risk tracking evaluation")
	for syntheticRiskIdPattern, riskTracking := range parsedModel.GetDeferredRiskTrackingDueToWildcardMatching() {
		progressReporter.Info("Applying wildcard risk tracking for risk id: " + syntheticRiskIdPattern)

		foundSome := false
		var matchingRiskIdExpression = regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(syntheticRiskIdPattern), `\*`, `[^@]+`))
		for syntheticRiskId := range parsedModel.GeneratedRisksBySyntheticId {
			if matchingRiskIdExpression.Match([]byte(syntheticRiskId)) && parsedModel.HasNotYetAnyDirectNonWildcardRiskTracking(syntheticRiskId) {
				foundSome = true
				parsedModel.RiskTracking[syntheticRiskId] = types.RiskTracking{
					SyntheticRiskId: strings.TrimSpace(syntheticRiskId),
					Justification:   riskTracking.Justification,
					CheckedBy:       riskTracking.CheckedBy,
					Ticket:          riskTracking.Ticket,
					Status:          riskTracking.Status,
					Date:            riskTracking.Date,
				}
			}
		}

		if !foundSome {
			if ignoreOrphanedRiskTracking {
				progressReporter.Warn("WARNING: Wildcard risk tracking does not match any risk id: " + syntheticRiskIdPattern)
			} else {
				return errors.New("wildcard risk tracking does not match any risk id: " + syntheticRiskIdPattern)
			}
		}
	}
	return nil
}

func (parsedModel *ParsedModel) CheckRiskTracking(ignoreOrphanedRiskTracking bool, progressReporter progressReporter) error {
	progressReporter.Info("Checking risk tracking")
	for _, tracking := range parsedModel.RiskTracking {
		if _, ok := parsedModel.GeneratedRisksBySyntheticId[tracking.SyntheticRiskId]; !ok {
			if ignoreOrphanedRiskTracking {
				progressReporter.Info("Risk tracking references unknown risk (risk id not found): " + tracking.SyntheticRiskId)
			} else {
				return errors.New("Risk tracking references unknown risk (risk id not found) - you might want to use the option -ignore-orphaned-risk-tracking: " + tracking.SyntheticRiskId +
					"\n\nNOTE: For risk tracking each risk-id needs to be defined (the string with the @ sign in it). " +
					"These unique risk IDs are visible in the PDF report (the small grey string under each risk), " +
					"the Excel (column \"ID\"), as well as the JSON responses. Some risk IDs have only one @ sign in them, " +
					"while others multiple. The idea is to allow for unique but still speaking IDs. Therefore each risk instance " +
					"creates its individual ID by taking all affected elements causing the risk to be within an @-delimited part. " +
					"Using wildcards (the * sign) for parts delimited by @ signs allows to handle groups of certain risks at once. " +
					"Best is to lookup the IDs to use in the created Excel file. Alternatively a model macro \"seed-risk-tracking\" " +
					"is available that helps in initially seeding the risk tracking part here based on already identified and not yet handled risks.")
			}
		}
	}

	// save also the risk-category-id and risk-status directly in the risk for better JSON marshalling
	for category := range parsedModel.GeneratedRisksByCategory {
		for i := range parsedModel.GeneratedRisksByCategory[category] {
			//			context.parsedModel.GeneratedRisksByCategory[category][i].CategoryId = category
			parsedModel.GeneratedRisksByCategory[category][i].RiskStatus = GetRiskTrackingStatusDefaultingUnchecked(parsedModel, parsedModel.GeneratedRisksByCategory[category][i])
		}
	}
	return nil
}

func (parsedModel *ParsedModel) CheckTagExists(referencedTag, where string) error {
	if !contains(parsedModel.TagsAvailable, referencedTag) {
		return errors.New("missing referenced tag in overall tag list at " + where + ": " + referencedTag)
	}
	return nil
}

func (parsedModel *ParsedModel) CheckDataAssetTargetExists(referencedAsset, where string) error {
	if _, ok := parsedModel.DataAssets[referencedAsset]; !ok {
		return errors.New("missing referenced data asset target at " + where + ": " + referencedAsset)
	}
	return nil
}

func (parsedModel *ParsedModel) CheckTrustBoundaryExists(referencedId, where string) error {
	if _, ok := parsedModel.TrustBoundaries[referencedId]; !ok {
		return errors.New("missing referenced trust boundary at " + where + ": " + referencedId)
	}
	return nil
}

func (parsedModel *ParsedModel) CheckSharedRuntimeExists(referencedId, where string) error {
	if _, ok := parsedModel.SharedRuntimes[referencedId]; !ok {
		return errors.New("missing referenced shared runtime at " + where + ": " + referencedId)
	}
	return nil
}

func (parsedModel *ParsedModel) CheckCommunicationLinkExists(referencedId, where string) error {
	if _, ok := parsedModel.CommunicationLinks[referencedId]; !ok {
		return errors.New("missing referenced communication link at " + where + ": " + referencedId)
	}
	return nil
}

func (parsedModel *ParsedModel) CheckTechnicalAssetExists(referencedAsset, where string, onlyForTweak bool) error {
	if _, ok := parsedModel.TechnicalAssets[referencedAsset]; !ok {
		suffix := ""
		if onlyForTweak {
			suffix = " (only referenced in diagram tweak)"
		}
		return errors.New("missing referenced technical asset target" + suffix + " at " + where + ": " + referencedAsset)
	}
	return nil
}

func (parsedModel *ParsedModel) CheckNestedTrustBoundariesExisting() error {
	for _, trustBoundary := range parsedModel.TrustBoundaries {
		for _, nestedId := range trustBoundary.TrustBoundariesNested {
			if _, ok := parsedModel.TrustBoundaries[nestedId]; !ok {
				return errors.New("missing referenced nested trust boundary: " + nestedId)
			}
		}
	}
	return nil
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

func (parsedModel *ParsedModel) InScopeTechnicalAssets() []types.TechnicalAsset {
	result := make([]types.TechnicalAsset, 0)
	for _, asset := range parsedModel.TechnicalAssets {
		if !asset.OutOfScope {
			result = append(result, asset)
		}
	}
	return result
}

func (parsedModel *ParsedModel) SortedTechnicalAssetIDs() []string {
	res := make([]string, 0)
	for id := range parsedModel.TechnicalAssets {
		res = append(res, id)
	}
	sort.Strings(res)
	return res
}

func (parsedModel *ParsedModel) TagsActuallyUsed() []string {
	result := make([]string, 0)
	for _, tag := range parsedModel.TagsAvailable {
		if len(parsedModel.TechnicalAssetsTaggedWithAny(tag)) > 0 ||
			len(parsedModel.CommunicationLinksTaggedWithAny(tag)) > 0 ||
			len(parsedModel.DataAssetsTaggedWithAny(tag)) > 0 ||
			len(parsedModel.TrustBoundariesTaggedWithAny(tag)) > 0 ||
			len(parsedModel.SharedRuntimesTaggedWithAny(tag)) > 0 {
			result = append(result, tag)
		}
	}
	return result
}

func (parsedModel *ParsedModel) TechnicalAssetsTaggedWithAny(tags ...string) []types.TechnicalAsset {
	result := make([]types.TechnicalAsset, 0)
	for _, candidate := range parsedModel.TechnicalAssets {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

func (parsedModel *ParsedModel) CommunicationLinksTaggedWithAny(tags ...string) []types.CommunicationLink {
	result := make([]types.CommunicationLink, 0)
	for _, asset := range parsedModel.TechnicalAssets {
		for _, candidate := range asset.CommunicationLinks {
			if candidate.IsTaggedWithAny(tags...) {
				result = append(result, candidate)
			}
		}
	}
	return result
}

func (parsedModel *ParsedModel) DataAssetsTaggedWithAny(tags ...string) []types.DataAsset {
	result := make([]types.DataAsset, 0)
	for _, candidate := range parsedModel.DataAssets {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

func (parsedModel *ParsedModel) TrustBoundariesTaggedWithAny(tags ...string) []types.TrustBoundary {
	result := make([]types.TrustBoundary, 0)
	for _, candidate := range parsedModel.TrustBoundaries {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

func (parsedModel *ParsedModel) SharedRuntimesTaggedWithAny(tags ...string) []types.SharedRuntime {
	result := make([]types.SharedRuntime, 0)
	for _, candidate := range parsedModel.SharedRuntimes {
		if candidate.IsTaggedWithAny(tags...) {
			result = append(result, candidate)
		}
	}
	return result
}

func (parsedModel *ParsedModel) OutOfScopeTechnicalAssets() []types.TechnicalAsset {
	assets := make([]types.TechnicalAsset, 0)
	for _, asset := range parsedModel.TechnicalAssets {
		if asset.OutOfScope {
			assets = append(assets, asset)
		}
	}
	sort.Sort(types.ByTechnicalAssetTitleSort(assets))
	return assets
}

func (parsedModel *ParsedModel) RisksOfOnlySTRIDEInformationDisclosure(risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, categoryRisks := range risksByCategory {
		for _, risk := range categoryRisks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.STRIDE == types.InformationDisclosure {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func (parsedModel *ParsedModel) RisksOfOnlySTRIDEDenialOfService(risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, categoryRisks := range risksByCategory {
		for _, risk := range categoryRisks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.STRIDE == types.DenialOfService {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func (parsedModel *ParsedModel) RisksOfOnlySTRIDEElevationOfPrivilege(risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, categoryRisks := range risksByCategory {
		for _, risk := range categoryRisks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.STRIDE == types.ElevationOfPrivilege {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func (parsedModel *ParsedModel) RisksOfOnlyBusinessSide(risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, categoryRisks := range risksByCategory {
		for _, risk := range categoryRisks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.BusinessSide {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func (parsedModel *ParsedModel) RisksOfOnlyArchitecture(risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, categoryRisks := range risksByCategory {
		for _, risk := range categoryRisks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Architecture {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func (parsedModel *ParsedModel) RisksOfOnlyDevelopment(risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, categoryRisks := range risksByCategory {
		for _, risk := range categoryRisks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Development {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func (parsedModel *ParsedModel) RisksOfOnlyOperation(risksByCategory map[string][]types.Risk) map[string][]types.Risk {
	result := make(map[string][]types.Risk)
	for categoryId, categoryRisks := range risksByCategory {
		for _, risk := range categoryRisks {
			category := GetRiskCategory(parsedModel, categoryId)
			if category.Function == types.Operations {
				result[categoryId] = append(result[categoryId], risk)
			}
		}
	}
	return result
}

func GetRiskTracking(model *ParsedModel, risk types.Risk) types.RiskTracking { // TODO: Unify function naming regarding Get etc.
	var result types.RiskTracking
	if riskTracking, ok := model.RiskTracking[risk.SyntheticId]; ok {
		result = riskTracking
	}
	return result
}

func GetRiskTrackingStatusDefaultingUnchecked(model *ParsedModel, risk types.Risk) types.RiskStatus {
	if riskTracking, ok := model.RiskTracking[risk.SyntheticId]; ok {
		return riskTracking.Status
	}
	return types.Unchecked
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
