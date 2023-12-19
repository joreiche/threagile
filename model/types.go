package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/threagile/threagile/colors"
	"github.com/threagile/threagile/pkg/security/types"
)

var ParsedModelRoot ParsedModel

var CommunicationLinks map[string]CommunicationLink // TODO as part of "ParsedModelRoot"?
var IncomingTechnicalCommunicationLinksMappedByTargetId map[string][]CommunicationLink
var DirectContainingTrustBoundaryMappedByTechnicalAssetId map[string]TrustBoundary
var DirectContainingSharedRuntimeMappedByTechnicalAssetId map[string]SharedRuntime

var GeneratedRisksByCategory map[RiskCategory][]Risk
var GeneratedRisksBySyntheticId map[string]Risk

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
	CommunicationLinks = make(map[string]CommunicationLink)
	IncomingTechnicalCommunicationLinksMappedByTargetId = make(map[string][]CommunicationLink)
	DirectContainingTrustBoundaryMappedByTechnicalAssetId = make(map[string]TrustBoundary)
	DirectContainingSharedRuntimeMappedByTechnicalAssetId = make(map[string]SharedRuntime)
	GeneratedRisksByCategory = make(map[RiskCategory][]Risk)
	GeneratedRisksBySyntheticId = make(map[string]Risk)
	AllSupportedTags = make(map[string]bool)
}

func AddToListOfSupportedTags(tags []string) {
	for _, tag := range tags {
		AllSupportedTags[tag] = true
	}
}

type CustomRiskRule struct {
	Category      func() RiskCategory
	SupportedTags func() []string
	GenerateRisks func(input *ParsedModel) []Risk
}

// === To be used by model macros etc. =======================

func AddTagToModelInput(modelInput *ModelInput, tag string, dryRun bool, changes *[]string) {
	tag = NormalizeTag(tag)
	if !Contains(modelInput.TagsAvailable, tag) {
		*changes = append(*changes, "adding tag: "+tag)
		if !dryRun {
			modelInput.TagsAvailable = append(modelInput.TagsAvailable, tag)
		}
	}
}

func NormalizeTag(tag string) string {
	return strings.TrimSpace(strings.ToLower(tag))
}

func MakeID(val string) string {
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")
	return strings.Trim(reg.ReplaceAllString(strings.ToLower(val), "-"), "- ")
}

// === Model Type Stuff ======================================

type ModelInput struct { // TODO: Eventually remove this and directly use ParsedModelRoot? But then the error messages for model errors are not quite as good anymore...
	Includes                                      []string                               `yaml:"includes" json:"includes"`
	ThreagileVersion                              string                                 `yaml:"threagile_version" json:"threagile_version"`
	Title                                         string                                 `yaml:"title" json:"title"`
	Author                                        Author                                 `yaml:"author" json:"author"`
	Date                                          string                                 `yaml:"date" json:"date"`
	BusinessOverview                              Overview                               `yaml:"business_overview" json:"business_overview"`
	TechnicalOverview                             Overview                               `yaml:"technical_overview" json:"technical_overview"`
	BusinessCriticality                           string                                 `yaml:"business_criticality" json:"business_criticality"`
	ManagementSummaryComment                      string                                 `yaml:"management_summary_comment" json:"management_summary_comment"`
	Questions                                     map[string]string                      `yaml:"questions" json:"questions"`
	AbuseCases                                    map[string]string                      `yaml:"abuse_cases" json:"abuse_cases"`
	SecurityRequirements                          map[string]string                      `yaml:"security_requirements" json:"security_requirements"`
	TagsAvailable                                 []string                               `yaml:"tags_available" json:"tags_available"`
	DataAssets                                    map[string]InputDataAsset              `yaml:"data_assets" json:"data_assets"`
	TechnicalAssets                               map[string]InputTechnicalAsset         `yaml:"technical_assets" json:"technical_assets"`
	TrustBoundaries                               map[string]InputTrustBoundary          `yaml:"trust_boundaries" json:"trust_boundaries"`
	SharedRuntimes                                map[string]InputSharedRuntime          `yaml:"shared_runtimes" json:"shared_runtimes"`
	IndividualRiskCategories                      map[string]InputIndividualRiskCategory `yaml:"individual_risk_categories" json:"individual_risk_categories"`
	RiskTracking                                  map[string]InputRiskTracking           `yaml:"risk_tracking" json:"risk_tracking"`
	DiagramTweakNodesep                           int                                    `yaml:"diagram_tweak_nodesep" json:"diagram_tweak_nodesep"`
	DiagramTweakRanksep                           int                                    `yaml:"diagram_tweak_ranksep" json:"diagram_tweak_ranksep"`
	DiagramTweakEdgeLayout                        string                                 `yaml:"diagram_tweak_edge_layout" json:"diagram_tweak_edge_layout"`
	DiagramTweakSuppressEdgeLabels                bool                                   `yaml:"diagram_tweak_suppress_edge_labels" json:"diagram_tweak_suppress_edge_labels"`
	DiagramTweakLayoutLeftToRight                 bool                                   `yaml:"diagram_tweak_layout_left_to_right" json:"diagram_tweak_layout_left_to_right"`
	DiagramTweakInvisibleConnectionsBetweenAssets []string                               `yaml:"diagram_tweak_invisible_connections_between_assets" json:"diagram_tweak_invisible_connections_between_assets"`
	DiagramTweakSameRankAssets                    []string                               `yaml:"diagram_tweak_same_rank_assets" json:"diagram_tweak_same_rank_assets"`
}

func (model *ModelInput) Defaults() *ModelInput {
	*model = ModelInput{
		Questions:                make(map[string]string),
		AbuseCases:               make(map[string]string),
		SecurityRequirements:     make(map[string]string),
		DataAssets:               make(map[string]InputDataAsset),
		TechnicalAssets:          make(map[string]InputTechnicalAsset),
		TrustBoundaries:          make(map[string]InputTrustBoundary),
		SharedRuntimes:           make(map[string]InputSharedRuntime),
		IndividualRiskCategories: make(map[string]InputIndividualRiskCategory),
		RiskTracking:             make(map[string]InputRiskTracking),
	}

	return model
}

func (model *ModelInput) Load(inputFilename string) error {
	modelYaml, readError := os.ReadFile(inputFilename)
	if readError != nil {
		log.Fatal("Unable to read model file: ", readError)
	}

	unmarshalError := yaml.Unmarshal(modelYaml, &model)
	if unmarshalError != nil {
		log.Fatal("Unable to parse model yaml: ", unmarshalError)
	}

	for _, includeFile := range model.Includes {
		mergeError := model.Merge(filepath.Dir(inputFilename), includeFile)
		if mergeError != nil {
			log.Fatalf("Unable to merge model include %q: %v", includeFile, mergeError)
		}
	}

	return nil
}

type UniqueStringSlice []string

func (slice UniqueStringSlice) Merge(otherSlice []string) []string {
	valueMap := make(map[string]bool)
	for _, value := range slice {
		valueMap[value] = true
	}

	for _, value := range otherSlice {
		valueMap[value] = true
	}

	valueSlice := make(UniqueStringSlice, 0)
	for key := range valueMap {
		valueSlice = append(valueSlice, key)
	}

	return valueSlice
}

func (model *ModelInput) Merge(dir string, includeFilename string) error {
	modelYaml, readError := os.ReadFile(filepath.Join(dir, includeFilename))
	if readError != nil {
		return fmt.Errorf("unable to read model file: %v", readError)
	}

	var fileStructure map[string]any
	unmarshalStructureError := yaml.Unmarshal(modelYaml, &fileStructure)
	if unmarshalStructureError != nil {
		return fmt.Errorf("unable to parse model structure: %v", unmarshalStructureError)
	}

	var includedModel ModelInput
	unmarshalError := yaml.Unmarshal(modelYaml, &includedModel)
	if unmarshalError != nil {
		return fmt.Errorf("unable to parse model yaml: %v", unmarshalError)
	}

	for item := range fileStructure {
		switch strings.ToLower(item) {
		case strings.ToLower("includes"):
			for _, includeFile := range includedModel.Includes {
				mergeError := model.Merge(filepath.Join(dir, filepath.Dir(includeFilename)), includeFile)
				if mergeError != nil {
					return fmt.Errorf("unable to merge model include %q: %v", includeFile, mergeError)
				}
			}
			break

		case strings.ToLower("threagile_version"):
			model.ThreagileVersion = includedModel.ThreagileVersion
			break

		case strings.ToLower("title"):
			model.Title = includedModel.Title
			break

		case strings.ToLower("author"):
			model.Author = includedModel.Author
			break

		case strings.ToLower("date"):
			model.Date = includedModel.Date
			break

		case strings.ToLower("business_overview"):
			model.BusinessOverview = includedModel.BusinessOverview
			break

		case strings.ToLower("technical_overview"):
			model.TechnicalOverview = includedModel.TechnicalOverview
			break

		case strings.ToLower("business_criticality"):
			model.BusinessCriticality = includedModel.BusinessCriticality
			break

		case strings.ToLower("management_summary_comment"):
			model.ManagementSummaryComment = includedModel.ManagementSummaryComment
			break

		case strings.ToLower("questions"):
			for mapKey, mapValue := range includedModel.Questions {
				model.Questions[mapKey] = mapValue
			}
			break

		case strings.ToLower("abuse_cases"):
			for mapKey, mapValue := range includedModel.AbuseCases {
				model.AbuseCases[mapKey] = mapValue
			}
			break

		case strings.ToLower("security_requirements"):
			for mapKey, mapValue := range includedModel.SecurityRequirements {
				model.SecurityRequirements[mapKey] = mapValue
			}
			break

		case strings.ToLower("tags_available"):
			model.TagsAvailable = UniqueStringSlice(model.TagsAvailable).Merge(includedModel.TagsAvailable)
			break

		case strings.ToLower("data_assets"):
			for mapKey, mapValue := range includedModel.DataAssets {
				model.DataAssets[mapKey] = mapValue
			}
			break

		case strings.ToLower("technical_assets"):
			for mapKey, mapValue := range includedModel.TechnicalAssets {
				model.TechnicalAssets[mapKey] = mapValue
			}
			break

		case strings.ToLower("trust_boundaries"):
			for mapKey, mapValue := range includedModel.TrustBoundaries {
				model.TrustBoundaries[mapKey] = mapValue
			}
			break

		case strings.ToLower("shared_runtimes"):
			for mapKey, mapValue := range includedModel.SharedRuntimes {
				model.SharedRuntimes[mapKey] = mapValue
			}
			break

		case strings.ToLower("individual_risk_categories"):
			for mapKey, mapValue := range includedModel.IndividualRiskCategories {
				model.IndividualRiskCategories[mapKey] = mapValue
			}
			break

		case strings.ToLower("risk_tracking"):
			for mapKey, mapValue := range includedModel.RiskTracking {
				model.RiskTracking[mapKey] = mapValue
			}
			break

		case "diagram_tweak_nodesep":
			model.DiagramTweakNodesep = includedModel.DiagramTweakNodesep
			break

		case "diagram_tweak_ranksep":
			model.DiagramTweakRanksep = includedModel.DiagramTweakRanksep
			break

		case "diagram_tweak_edge_layout":
			model.DiagramTweakEdgeLayout = includedModel.DiagramTweakEdgeLayout
			break

		case "diagram_tweak_suppress_edge_labels":
			model.DiagramTweakSuppressEdgeLabels = includedModel.DiagramTweakSuppressEdgeLabels
			break

		case "diagram_tweak_layout_left_to_right":
			model.DiagramTweakLayoutLeftToRight = includedModel.DiagramTweakLayoutLeftToRight
			break

		case "diagram_tweak_invisible_connections_between_assets":
			model.DiagramTweakInvisibleConnectionsBetweenAssets = append(model.DiagramTweakInvisibleConnectionsBetweenAssets, includedModel.DiagramTweakInvisibleConnectionsBetweenAssets...)
			break

		case "diagram_tweak_same_rank_assets":
			model.DiagramTweakSameRankAssets = append(model.DiagramTweakSameRankAssets, includedModel.DiagramTweakSameRankAssets...)
		}
	}

	return nil
}

type InputDataAsset struct {
	ID                     string   `yaml:"id" json:"id"`
	Description            string   `yaml:"description" json:"description"`
	Usage                  string   `yaml:"usage" json:"usage"`
	Tags                   []string `yaml:"tags" json:"tags"`
	Origin                 string   `yaml:"origin" json:"origin"`
	Owner                  string   `yaml:"owner" json:"owner"`
	Quantity               string   `yaml:"quantity" json:"quantity"`
	Confidentiality        string   `yaml:"confidentiality" json:"confidentiality"`
	Integrity              string   `yaml:"integrity" json:"integrity"`
	Availability           string   `yaml:"availability" json:"availability"`
	JustificationCiaRating string   `yaml:"justification_cia_rating" json:"justification_cia_rating"`
}

type InputTechnicalAsset struct {
	ID                      string                            `yaml:"id" json:"id"`
	Description             string                            `yaml:"description" json:"description"`
	Type                    string                            `yaml:"type" json:"type"`
	Usage                   string                            `yaml:"usage" json:"usage"`
	UsedAsClientByHuman     bool                              `yaml:"used_as_client_by_human" json:"used_as_client_by_human"`
	OutOfScope              bool                              `yaml:"out_of_scope" json:"out_of_scope"`
	JustificationOutOfScope string                            `yaml:"justification_out_of_scope" json:"justification_out_of_scope"`
	Size                    string                            `yaml:"size" json:"size"`
	Technology              string                            `yaml:"technology" json:"technology"`
	Tags                    []string                          `yaml:"tags" json:"tags"`
	Internet                bool                              `yaml:"internet" json:"internet"`
	Machine                 string                            `yaml:"machine" json:"machine"`
	Encryption              string                            `yaml:"encryption" json:"encryption"`
	Owner                   string                            `yaml:"owner" json:"owner"`
	Confidentiality         string                            `yaml:"confidentiality" json:"confidentiality"`
	Integrity               string                            `yaml:"integrity" json:"integrity"`
	Availability            string                            `yaml:"availability" json:"availability"`
	JustificationCiaRating  string                            `yaml:"justification_cia_rating" json:"justification_cia_rating"`
	MultiTenant             bool                              `yaml:"multi_tenant" json:"multi_tenant"`
	Redundant               bool                              `yaml:"redundant" json:"redundant"`
	CustomDevelopedParts    bool                              `yaml:"custom_developed_parts" json:"custom_developed_parts"`
	DataAssetsProcessed     []string                          `yaml:"data_assets_processed" json:"data_assets_processed"`
	DataAssetsStored        []string                          `yaml:"data_assets_stored" json:"data_assets_stored"`
	DataFormatsAccepted     []string                          `yaml:"data_formats_accepted" json:"data_formats_accepted"`
	DiagramTweakOrder       int                               `yaml:"diagram_tweak_order" json:"diagram_tweak_order"`
	CommunicationLinks      map[string]InputCommunicationLink `yaml:"communication_links" json:"communication_links"`
}

type InputCommunicationLink struct {
	Target                 string   `yaml:"target" json:"target"`
	Description            string   `yaml:"description" json:"description"`
	Protocol               string   `yaml:"protocol" json:"protocol"`
	Authentication         string   `yaml:"authentication" json:"authentication"`
	Authorization          string   `yaml:"authorization" json:"authorization"`
	Tags                   []string `yaml:"tags" json:"tags"`
	VPN                    bool     `yaml:"vpn" json:"vpn"`
	IpFiltered             bool     `yaml:"ip_filtered" json:"ip_filtered"`
	Readonly               bool     `yaml:"readonly" json:"readonly"`
	Usage                  string   `yaml:"usage" json:"usage"`
	DataAssetsSent         []string `yaml:"data_assets_sent" json:"data_assets_sent"`
	DataAssetsReceived     []string `yaml:"data_assets_received" json:"data_assets_received"`
	DiagramTweakWeight     int      `yaml:"diagram_tweak_weight" json:"diagram_tweak_weight"`
	DiagramTweakConstraint bool     `yaml:"diagram_tweak_constraint" json:"diagram_tweak_constraint"`
}

type InputSharedRuntime struct {
	ID                     string   `yaml:"id" json:"id"`
	Description            string   `yaml:"description" json:"description"`
	Tags                   []string `yaml:"tags" json:"tags"`
	TechnicalAssetsRunning []string `yaml:"technical_assets_running" json:"technical_assets_running"`
}

type InputTrustBoundary struct {
	ID                    string   `yaml:"id" json:"id"`
	Description           string   `yaml:"description" json:"description"`
	Type                  string   `yaml:"type" json:"type"`
	Tags                  []string `yaml:"tags" json:"tags"`
	TechnicalAssetsInside []string `yaml:"technical_assets_inside" json:"technical_assets_inside"`
	TrustBoundariesNested []string `yaml:"trust_boundaries_nested" json:"trust_boundaries_nested"`
}

type InputIndividualRiskCategory struct {
	ID                         string                         `yaml:"id" json:"id"`
	Description                string                         `yaml:"description" json:"description"`
	Impact                     string                         `yaml:"impact" json:"impact"`
	ASVS                       string                         `yaml:"asvs" json:"asvs"`
	CheatSheet                 string                         `yaml:"cheat_sheet" json:"cheat_sheet"`
	Action                     string                         `yaml:"action" json:"action"`
	Mitigation                 string                         `yaml:"mitigation" json:"mitigation"`
	Check                      string                         `yaml:"check" json:"check"`
	Function                   string                         `yaml:"function" json:"function"`
	STRIDE                     string                         `yaml:"stride" json:"stride"`
	DetectionLogic             string                         `yaml:"detection_logic" json:"detection_logic"`
	RiskAssessment             string                         `yaml:"risk_assessment" json:"risk_assessment"`
	FalsePositives             string                         `yaml:"false_positives" json:"false_positives"`
	ModelFailurePossibleReason bool                           `yaml:"model_failure_possible_reason" json:"model_failure_possible_reason"`
	CWE                        int                            `yaml:"cwe" json:"cwe"`
	RisksIdentified            map[string]InputRiskIdentified `yaml:"risks_identified" json:"risks_identified"`
}

type InputRiskIdentified struct {
	Severity                      string   `yaml:"severity" json:"severity"`
	ExploitationLikelihood        string   `yaml:"exploitation_likelihood" json:"exploitation_likelihood"`
	ExploitationImpact            string   `yaml:"exploitation_impact" json:"exploitation_impact"`
	DataBreachProbability         string   `yaml:"data_breach_probability" json:"data_breach_probability"`
	DataBreachTechnicalAssets     []string `yaml:"data_breach_technical_assets" json:"data_breach_technical_assets"`
	MostRelevantDataAsset         string   `yaml:"most_relevant_data_asset" json:"most_relevant_data_asset"`
	MostRelevantTechnicalAsset    string   `yaml:"most_relevant_technical_asset" json:"most_relevant_technical_asset"`
	MostRelevantCommunicationLink string   `yaml:"most_relevant_communication_link" json:"most_relevant_communication_link"`
	MostRelevantTrustBoundary     string   `yaml:"most_relevant_trust_boundary" json:"most_relevant_trust_boundary"`
	MostRelevantSharedRuntime     string   `yaml:"most_relevant_shared_runtime" json:"most_relevant_shared_runtime"`
}

type InputRiskTracking struct {
	Status        string `yaml:"status" json:"status"`
	Justification string `yaml:"justification" json:"justification"`
	Ticket        string `yaml:"ticket" json:"ticket"`
	Date          string `yaml:"date" json:"date"`
	CheckedBy     string `yaml:"checked_by" json:"checked_by"`
}

// TODO: pass ParsedModelRoot as parameter instead of using global variable
func (what TrustBoundary) addAssetIDsRecursively(result *[]string) {
	*result = append(*result, what.TechnicalAssetsInside...)
	for _, nestedBoundaryID := range what.TrustBoundariesNested {
		ParsedModelRoot.TrustBoundaries[nestedBoundaryID].addAssetIDsRecursively(result)
	}
}

// TODO: pass ParsedModelRoot as parameter instead of using global variable
func (what TrustBoundary) addTrustBoundaryIDsRecursively(result *[]string) {
	*result = append(*result, what.Id)
	parentID := what.ParentTrustBoundaryID()
	if len(parentID) > 0 {
		ParsedModelRoot.TrustBoundaries[parentID].addTrustBoundaryIDsRecursively(result)
	}
}

func IsSharingSameParentTrustBoundary(left, right TechnicalAsset) bool {
	tbIDLeft, tbIDRight := left.GetTrustBoundaryId(), right.GetTrustBoundaryId()
	if len(tbIDLeft) == 0 && len(tbIDRight) > 0 {
		return false
	}
	if len(tbIDLeft) > 0 && len(tbIDRight) == 0 {
		return false
	}
	if len(tbIDLeft) == 0 && len(tbIDRight) == 0 {
		return true
	}
	if tbIDLeft == tbIDRight {
		return true
	}
	tbLeft, tbRight := ParsedModelRoot.TrustBoundaries[tbIDLeft], ParsedModelRoot.TrustBoundaries[tbIDRight]
	tbParentsLeft, tbParentsRight := tbLeft.AllParentTrustBoundaryIDs(), tbRight.AllParentTrustBoundaryIDs()
	for _, parentLeft := range tbParentsLeft {
		for _, parentRight := range tbParentsRight {
			if parentLeft == parentRight {
				return true
			}
		}
	}
	return false
}

func (what TrustBoundary) AllParentTrustBoundaryIDs() []string {
	result := make([]string, 0)
	what.addTrustBoundaryIDsRecursively(&result)
	return result
}

type DataAsset struct {
	Id                     string                `yaml:"id" json:"id"`                   // TODO: tag here still required?
	Title                  string                `yaml:"title" json:"title"`             // TODO: tag here still required?
	Description            string                `yaml:"description" json:"description"` // TODO: tag here still required?
	Usage                  types.Usage           `yaml:"usage" json:"usage"`
	Tags                   []string              `yaml:"tags" json:"tags"`
	Origin                 string                `yaml:"origin" json:"origin"`
	Owner                  string                `yaml:"owner" json:"owner"`
	Quantity               types.Quantity        `yaml:"quantity" json:"quantity"`
	Confidentiality        types.Confidentiality `yaml:"confidentiality" json:"confidentiality"`
	Integrity              types.Criticality     `yaml:"integrity" json:"integrity"`
	Availability           types.Criticality     `yaml:"availability" json:"availability"`
	JustificationCiaRating string                `yaml:"justification_cia_rating" json:"justification_cia_rating"`
}

func (what DataAsset) IsTaggedWithAny(tags ...string) bool {
	return ContainsCaseInsensitiveAny(what.Tags, tags...)
}

func (what DataAsset) IsTaggedWithBaseTag(baseTag string) bool {
	return IsTaggedWithBaseTag(what.Tags, baseTag)
}

/*
func (what DataAsset) IsAtRisk() bool {
	for _, techAsset := range what.ProcessedByTechnicalAssetsSorted() {
		if len(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks())) > 0 {
			return true
		}
	}
	for _, techAsset := range what.StoredByTechnicalAssetsSorted() {
		if len(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks())) > 0 {
			return true
		}
	}
	return false
}
*/

/*
func (what DataAsset) IdentifiedRiskSeverityStillAtRisk() RiskSeverity {
	highestRiskSeverity := Low
	for _, techAsset := range what.ProcessedByTechnicalAssetsSorted() {
		candidateSeverity := HighestSeverityStillAtRisk(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks()))
		if candidateSeverity > highestRiskSeverity {
			highestRiskSeverity = candidateSeverity
		}
	}
	for _, techAsset := range what.StoredByTechnicalAssetsSorted() {
		candidateSeverity := HighestSeverityStillAtRisk(ReduceToOnlyStillAtRisk(techAsset.GeneratedRisks()))
		if candidateSeverity > highestRiskSeverity {
			highestRiskSeverity = candidateSeverity
		}
	}
	return highestRiskSeverity
}
*/

func (what DataAsset) IdentifiedRisksByResponsibleTechnicalAssetId() map[string][]Risk {
	uniqueTechAssetIDsResponsibleForThisDataAsset := make(map[string]interface{})
	for _, techAsset := range what.ProcessedByTechnicalAssetsSorted() {
		if len(techAsset.GeneratedRisks()) > 0 {
			uniqueTechAssetIDsResponsibleForThisDataAsset[techAsset.Id] = true
		}
	}
	for _, techAsset := range what.StoredByTechnicalAssetsSorted() {
		if len(techAsset.GeneratedRisks()) > 0 {
			uniqueTechAssetIDsResponsibleForThisDataAsset[techAsset.Id] = true
		}
	}

	result := make(map[string][]Risk)
	for techAssetId := range uniqueTechAssetIDsResponsibleForThisDataAsset {
		result[techAssetId] = append(result[techAssetId], ParsedModelRoot.TechnicalAssets[techAssetId].GeneratedRisks()...)
	}
	return result
}

func (what DataAsset) IsDataBreachPotentialStillAtRisk() bool {
	for _, risk := range FilteredByStillAtRisk() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				return true
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				return true
			}
		}
	}
	return false
}

func (what DataAsset) IdentifiedDataBreachProbability() types.DataBreachProbability {
	highestProbability := types.Improbable
	for _, risk := range AllRisks() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
		}
	}
	return highestProbability
}

func (what DataAsset) IdentifiedDataBreachProbabilityStillAtRisk() types.DataBreachProbability {
	highestProbability := types.Improbable
	for _, risk := range FilteredByStillAtRisk() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				if risk.DataBreachProbability > highestProbability {
					highestProbability = risk.DataBreachProbability
					break
				}
			}
		}
	}
	return highestProbability
}

func (what DataAsset) IdentifiedDataBreachProbabilityRisksStillAtRisk() []Risk {
	result := make([]Risk, 0)
	for _, risk := range FilteredByStillAtRisk() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				result = append(result, risk)
				break
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				result = append(result, risk)
				break
			}
		}
	}
	return result
}

func (what DataAsset) IdentifiedDataBreachProbabilityRisks() []Risk {
	result := make([]Risk, 0)
	for _, risk := range AllRisks() {
		for _, techAsset := range risk.DataBreachTechnicalAssetIDs {
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsProcessed, what.Id) {
				result = append(result, risk)
				break
			}
			if Contains(ParsedModelRoot.TechnicalAssets[techAsset].DataAssetsStored, what.Id) {
				result = append(result, risk)
				break
			}
		}
	}
	return result
}

func (what DataAsset) ProcessedByTechnicalAssetsSorted() []TechnicalAsset {
	result := make([]TechnicalAsset, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, candidateID := range technicalAsset.DataAssetsProcessed {
			if candidateID == what.Id {
				result = append(result, technicalAsset)
			}
		}
	}
	sort.Sort(ByTechnicalAssetTitleSort(result))
	return result
}

func (what DataAsset) StoredByTechnicalAssetsSorted() []TechnicalAsset {
	result := make([]TechnicalAsset, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, candidateID := range technicalAsset.DataAssetsStored {
			if candidateID == what.Id {
				result = append(result, technicalAsset)
			}
		}
	}
	sort.Sort(ByTechnicalAssetTitleSort(result))
	return result
}

func (what DataAsset) SentViaCommLinksSorted() []CommunicationLink {
	result := make([]CommunicationLink, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, commLink := range technicalAsset.CommunicationLinks {
			for _, candidateID := range commLink.DataAssetsSent {
				if candidateID == what.Id {
					result = append(result, commLink)
				}
			}
		}
	}
	sort.Sort(ByTechnicalCommunicationLinkTitleSort(result))
	return result
}

func (what DataAsset) ReceivedViaCommLinksSorted() []CommunicationLink {
	result := make([]CommunicationLink, 0)
	for _, technicalAsset := range ParsedModelRoot.TechnicalAssets {
		for _, commLink := range technicalAsset.CommunicationLinks {
			for _, candidateID := range commLink.DataAssetsReceived {
				if candidateID == what.Id {
					result = append(result, commLink)
				}
			}
		}
	}
	sort.Sort(ByTechnicalCommunicationLinkTitleSort(result))
	return result
}

func IsTaggedWithBaseTag(tags []string, baseTag string) bool { // base tags are before the colon ":" like in "aws:ec2" it's "aws". The subtag is after the colon. Also, a pure "aws" tag matches the base tag "aws"
	baseTag = strings.ToLower(strings.TrimSpace(baseTag))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == baseTag || strings.HasPrefix(tag, baseTag+":") {
			return true
		}
	}
	return false
}

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
	return ContainsCaseInsensitiveAny(what.Tags, tags...)
}

func (what TechnicalAsset) IsTaggedWithBaseTag(baseTag string) bool {
	return IsTaggedWithBaseTag(what.Tags, baseTag)
}

// first use the tag(s) of the asset itself, then their trust boundaries (recursively up) and then their shared runtime

func (what TechnicalAsset) IsTaggedWithAnyTraversingUp(tags ...string) bool {
	if ContainsCaseInsensitiveAny(what.Tags, tags...) {
		return true
	}
	tbID := what.GetTrustBoundaryId()
	if len(tbID) > 0 {
		if ParsedModelRoot.TrustBoundaries[tbID].IsTaggedWithAnyTraversingUp(tags...) {
			return true
		}
	}
	for _, sr := range ParsedModelRoot.SharedRuntimes {
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

func (what TechnicalAsset) HighestIntegrity() types.Criticality {
	highest := what.Integrity
	for _, dataId := range what.DataAssetsProcessed {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Integrity > highest {
			highest = dataAsset.Integrity
		}
	}
	for _, dataId := range what.DataAssetsStored {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Integrity > highest {
			highest = dataAsset.Integrity
		}
	}
	return highest
}

func (what TechnicalAsset) HighestAvailability() types.Criticality {
	highest := what.Availability
	for _, dataId := range what.DataAssetsProcessed {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Availability > highest {
			highest = dataAsset.Availability
		}
	}
	for _, dataId := range what.DataAssetsStored {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
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

type ByDataAssetDataBreachProbabilityAndTitleSort []DataAsset

func (what ByDataAssetDataBreachProbabilityAndTitleSort) Len() int { return len(what) }
func (what ByDataAssetDataBreachProbabilityAndTitleSort) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByDataAssetDataBreachProbabilityAndTitleSort) Less(i, j int) bool {
	highestDataBreachProbabilityLeft := what[i].IdentifiedDataBreachProbability()
	highestDataBreachProbabilityRight := what[j].IdentifiedDataBreachProbability()
	if highestDataBreachProbabilityLeft == highestDataBreachProbabilityRight {
		return what[i].Title < what[j].Title
	}
	return highestDataBreachProbabilityLeft > highestDataBreachProbabilityRight
}

type ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk []DataAsset

func (what ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk) Len() int { return len(what) }
func (what ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk) Swap(i, j int) {
	what[i], what[j] = what[j], what[i]
}
func (what ByDataAssetDataBreachProbabilityAndTitleSortStillAtRisk) Less(i, j int) bool {
	risksLeft := what[i].IdentifiedDataBreachProbabilityRisksStillAtRisk()
	risksRight := what[j].IdentifiedDataBreachProbabilityRisksStillAtRisk()
	highestDataBreachProbabilityLeft := what[i].IdentifiedDataBreachProbabilityStillAtRisk()
	highestDataBreachProbabilityRight := what[j].IdentifiedDataBreachProbabilityStillAtRisk()
	if highestDataBreachProbabilityLeft == highestDataBreachProbabilityRight {
		if len(risksLeft) == 0 && len(risksRight) > 0 {
			return false
		}
		if len(risksLeft) > 0 && len(risksRight) == 0 {
			return true
		}
		return what[i].Title < what[j].Title
	}
	return highestDataBreachProbabilityLeft > highestDataBreachProbabilityRight
}

type ByOrderAndIdSort []TechnicalAsset

func (what ByOrderAndIdSort) Len() int      { return len(what) }
func (what ByOrderAndIdSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByOrderAndIdSort) Less(i, j int) bool {
	if what[i].DiagramTweakOrder == what[j].DiagramTweakOrder {
		return what[i].Id > what[j].Id
	}
	return what[i].DiagramTweakOrder < what[j].DiagramTweakOrder
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

type ByDataAssetTitleSort []DataAsset

func (what ByDataAssetTitleSort) Len() int      { return len(what) }
func (what ByDataAssetTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByDataAssetTitleSort) Less(i, j int) bool {
	return what[i].Title < what[j].Title
}

type ByDataFormatAcceptedSort []types.DataFormat

func (what ByDataFormatAcceptedSort) Len() int      { return len(what) }
func (what ByDataFormatAcceptedSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByDataFormatAcceptedSort) Less(i, j int) bool {
	return what[i].String() < what[j].String()
}

type CommunicationLink struct {
	Id, SourceId, TargetId, Title, Description string
	Protocol                                   types.Protocol
	Tags                                       []string
	VPN, IpFiltered, Readonly                  bool
	Authentication                             types.Authentication
	Authorization                              types.Authorization
	Usage                                      types.Usage
	DataAssetsSent, DataAssetsReceived         []string
	DiagramTweakWeight                         int
	DiagramTweakConstraint                     bool
}

func (what CommunicationLink) IsTaggedWithAny(tags ...string) bool {
	return ContainsCaseInsensitiveAny(what.Tags, tags...)
}

func (what CommunicationLink) IsTaggedWithBaseTag(baseTag string) bool {
	return IsTaggedWithBaseTag(what.Tags, baseTag)
}

type ByTechnicalCommunicationLinkIdSort []CommunicationLink

func (what ByTechnicalCommunicationLinkIdSort) Len() int      { return len(what) }
func (what ByTechnicalCommunicationLinkIdSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByTechnicalCommunicationLinkIdSort) Less(i, j int) bool {
	return what[i].Id > what[j].Id
}

type ByTechnicalCommunicationLinkTitleSort []CommunicationLink

func (what ByTechnicalCommunicationLinkTitleSort) Len() int      { return len(what) }
func (what ByTechnicalCommunicationLinkTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByTechnicalCommunicationLinkTitleSort) Less(i, j int) bool {
	return what[i].Title > what[j].Title
}

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
	return ContainsCaseInsensitiveAny(what.Tags, tags...)
}

func (what TrustBoundary) IsTaggedWithBaseTag(baseTag string) bool {
	return IsTaggedWithBaseTag(what.Tags, baseTag)
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

type SharedRuntime struct {
	Id, Title, Description string
	Tags                   []string
	TechnicalAssetsRunning []string
}

func (what SharedRuntime) IsTaggedWithAny(tags ...string) bool {
	return ContainsCaseInsensitiveAny(what.Tags, tags...)
}

func (what SharedRuntime) IsTaggedWithBaseTag(baseTag string) bool {
	return IsTaggedWithBaseTag(what.Tags, baseTag)
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

func (what CommunicationLink) IsAcrossTrustBoundary() bool {
	trustBoundaryOfSourceAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[what.SourceId]
	trustBoundaryOfTargetAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[what.TargetId]
	return trustBoundaryOfSourceAsset.Id != trustBoundaryOfTargetAsset.Id
}

func (what CommunicationLink) IsAcrossTrustBoundaryNetworkOnly() bool {
	trustBoundaryOfSourceAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[what.SourceId]
	if !trustBoundaryOfSourceAsset.Type.IsNetworkBoundary() { // find and use the parent boundary then
		trustBoundaryOfSourceAsset = ParsedModelRoot.TrustBoundaries[trustBoundaryOfSourceAsset.ParentTrustBoundaryID()]
	}
	trustBoundaryOfTargetAsset := DirectContainingTrustBoundaryMappedByTechnicalAssetId[what.TargetId]
	if !trustBoundaryOfTargetAsset.Type.IsNetworkBoundary() { // find and use the parent boundary then
		trustBoundaryOfTargetAsset = ParsedModelRoot.TrustBoundaries[trustBoundaryOfTargetAsset.ParentTrustBoundaryID()]
	}
	return trustBoundaryOfSourceAsset.Id != trustBoundaryOfTargetAsset.Id && trustBoundaryOfTargetAsset.Type.IsNetworkBoundary()
}

func (what CommunicationLink) HighestConfidentiality() types.Confidentiality {
	highest := types.Public
	for _, dataId := range what.DataAssetsSent {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Confidentiality > highest {
			highest = dataAsset.Confidentiality
		}
	}
	for _, dataId := range what.DataAssetsReceived {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Confidentiality > highest {
			highest = dataAsset.Confidentiality
		}
	}
	return highest
}

func (what CommunicationLink) HighestIntegrity() types.Criticality {
	highest := types.Archive
	for _, dataId := range what.DataAssetsSent {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Integrity > highest {
			highest = dataAsset.Integrity
		}
	}
	for _, dataId := range what.DataAssetsReceived {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Integrity > highest {
			highest = dataAsset.Integrity
		}
	}
	return highest
}

func (what CommunicationLink) HighestAvailability() types.Criticality {
	highest := types.Archive
	for _, dataId := range what.DataAssetsSent {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Availability > highest {
			highest = dataAsset.Availability
		}
	}
	for _, dataId := range what.DataAssetsReceived {
		dataAsset := ParsedModelRoot.DataAssets[dataId]
		if dataAsset.Availability > highest {
			highest = dataAsset.Availability
		}
	}
	return highest
}

func (what CommunicationLink) DataAssetsSentSorted() []DataAsset {
	result := make([]DataAsset, 0)
	for _, assetID := range what.DataAssetsSent {
		result = append(result, ParsedModelRoot.DataAssets[assetID])
	}
	sort.Sort(ByDataAssetTitleSort(result))
	return result
}

func (what CommunicationLink) DataAssetsReceivedSorted() []DataAsset {
	result := make([]DataAsset, 0)
	for _, assetID := range what.DataAssetsReceived {
		result = append(result, ParsedModelRoot.DataAssets[assetID])
	}
	sort.Sort(ByDataAssetTitleSort(result))
	return result
}

type Author struct {
	Name     string `yaml:"name" json:"name"`
	Homepage string `yaml:"homepage" json:"homepage"`
}

type Overview struct {
	Description string              `yaml:"description" json:"description"`
	Images      []map[string]string `yaml:"images" json:"images"` // yes, array of map here, as array keeps the order of the image keys
}

type ParsedModel struct {
	Author                                        Author
	Title                                         string
	Date                                          time.Time
	ManagementSummaryComment                      string
	BusinessOverview                              Overview
	TechnicalOverview                             Overview
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
	sort.Sort(ByDataAssetTitleSort(assets))
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

func (what CommunicationLink) IsBidirectional() bool {
	return len(what.DataAssetsSent) > 0 && len(what.DataAssetsReceived) > 0
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

func ContainsCaseInsensitiveAny(a []string, x ...string) bool {
	for _, n := range a {
		for _, c := range x {
			if strings.TrimSpace(strings.ToLower(c)) == strings.TrimSpace(strings.ToLower(n)) {
				return true
			}
		}
	}
	return false
}

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

type MacroDetails struct {
	ID, Title, Description string
}

type MacroQuestion struct {
	ID, Title, Description string
	PossibleAnswers        []string
	MultiSelect            bool
	DefaultAnswer          string
}

const NoMoreQuestionsID = ""

func NoMoreQuestions() MacroQuestion {
	return MacroQuestion{
		ID:              NoMoreQuestionsID,
		Title:           "",
		Description:     "",
		PossibleAnswers: nil,
		MultiSelect:     false,
		DefaultAnswer:   "",
	}
}

func (what MacroQuestion) NoMoreQuestions() bool {
	return what.ID == NoMoreQuestionsID
}

func (what MacroQuestion) IsValueConstrained() bool {
	return what.PossibleAnswers != nil && len(what.PossibleAnswers) > 0
}

func (what MacroQuestion) IsMatchingValueConstraint(answer string) bool {
	if what.IsValueConstrained() {
		for _, val := range what.PossibleAnswers {
			if strings.ToLower(val) == strings.ToLower(answer) {
				return true
			}
		}
		return false
	}
	return true
}

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

type ByRiskCategoryTitleSort []RiskCategory

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

func (what Risk) GetRiskTracking() RiskTracking { // TODO: Unify function naming regarding Get etc.
	var result RiskTracking
	if riskTracking, ok := ParsedModelRoot.RiskTracking[what.SyntheticId]; ok {
		result = riskTracking
	}
	return result
}

func (what Risk) GetRiskTrackingStatusDefaultingUnchecked() types.RiskStatus {
	if riskTracking, ok := ParsedModelRoot.RiskTracking[what.SyntheticId]; ok {
		return riskTracking.Status
	}
	return types.Unchecked
}

func (what Risk) IsRiskTracked() bool {
	if _, ok := ParsedModelRoot.RiskTracking[what.SyntheticId]; ok {
		return true
	}
	return false
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

type RiskTracking struct {
	SyntheticRiskId, Justification, Ticket, CheckedBy string
	Status                                            types.RiskStatus
	Date                                              time.Time
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

func InScopeTechnicalAssets() []TechnicalAsset {
	result := make([]TechnicalAsset, 0)
	for _, asset := range ParsedModelRoot.TechnicalAssets {
		if !asset.OutOfScope {
			result = append(result, asset)
		}
	}
	return result
}
