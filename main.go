package main

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/threagile/threagile/cmd"
	"github.com/threagile/threagile/pkg/colors"
	"github.com/threagile/threagile/pkg/model"
	"github.com/threagile/threagile/pkg/security/types"
)

const (
	keepDiagramSourceFiles = false
	addModelTitle          = false
)

const (
	defaultGraphvizDPI, maxGraphvizDPI = 120, 240
	backupHistoryFilesToKeep           = 50
)

const (
	buildTimestamp                         = ""
	tempDir                                = "/dev/shm" // TODO: make configurable via cmdline arg?
	binDir                                 = "/app"
	appDir                                 = "/app"
	dataDir                                = "/data"
	keyDir                                 = "keys"
	reportFilename                         = "report.pdf"
	excelRisksFilename                     = "risks.xlsx"
	excelTagsFilename                      = "tags.xlsx"
	jsonRisksFilename                      = "risks.json"
	jsonTechnicalAssetsFilename            = "technical-assets.json"
	jsonStatsFilename                      = "stats.json"
	dataFlowDiagramFilenameDOT             = "data-flow-diagram.gv"
	dataFlowDiagramFilenamePNG             = "data-flow-diagram.png"
	dataAssetDiagramFilenameDOT            = "data-asset-diagram.gv"
	dataAssetDiagramFilenamePNG            = "data-asset-diagram.png"
	graphvizDataFlowDiagramConversionCall  = "render-data-flow-diagram.sh"
	graphvizDataAssetDiagramConversionCall = "render-data-asset-diagram.sh"
	inputFile                              = "threagile.yaml"
)

// === Error handling stuff ========================================

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	cmd.Execute()

	// TODO: remove below as soon as refactoring is finished - everything will go through rootCmd.Execute
	// for now it's fine to have as frequently uncommented to see the actual behaviour
	// context := new(Context).Defaults()
	// context.parseCommandlineArgs()
	// if *context.serverPort > 0 {
	// 	context.startServer()
	// } else {
	// 	context.doIt()
	// }
}

const keySize = 32

type timeoutStruct struct {
	xorRand                               []byte
	createdNanoTime, lastAccessedNanoTime int64
}

var mapTokenHashToTimeoutStruct = make(map[string]timeoutStruct)
var mapFolderNameToTokenHash = make(map[string]string)

const extremeShortTimeoutsForTesting = false

func housekeepingTokenMaps() {
	now := time.Now().UnixNano()
	for tokenHash, val := range mapTokenHashToTimeoutStruct {
		if extremeShortTimeoutsForTesting {
			// remove all elements older than 1 minute (= 60000000000 ns) soft
			// and all elements older than 3 minutes (= 180000000000 ns) hard
			if now-val.lastAccessedNanoTime > 60000000000 || now-val.createdNanoTime > 180000000000 {
				fmt.Println("About to remove a token hash from maps")
				deleteTokenHashFromMaps(tokenHash)
			}
		} else {
			// remove all elements older than 30 minutes (= 1800000000000 ns) soft
			// and all elements older than 10 hours (= 36000000000000 ns) hard
			if now-val.lastAccessedNanoTime > 1800000000000 || now-val.createdNanoTime > 36000000000000 {
				deleteTokenHashFromMaps(tokenHash)
			}
		}
	}
}

func deleteTokenHashFromMaps(tokenHash string) {
	delete(mapTokenHashToTimeoutStruct, tokenHash)
	for folderName, check := range mapFolderNameToTokenHash {
		if check == tokenHash {
			delete(mapFolderNameToTokenHash, folderName)
			break
		}
	}
}

func xor(key []byte, xor []byte) []byte {
	if len(key) != len(xor) {
		panic(errors.New("key length not matching XOR length"))
	}
	result := make([]byte, len(xor))
	for i, b := range key {
		result[i] = b ^ xor[i]
	}
	return result
}

type responseType int

const (
	dataFlowDiagram responseType = iota
	dataAssetDiagram
	reportPDF
	risksExcel
	tagsExcel
	risksJSON
	technicalAssetsJSON
	statsJSON
)

type payloadModels struct {
	ID                string    `yaml:"id" json:"id"`
	Title             string    `yaml:"title" json:"title"`
	TimestampCreated  time.Time `yaml:"timestamp_created" json:"timestamp_created"`
	TimestampModified time.Time `yaml:"timestamp_modified" json:"timestamp_modified"`
}

type payloadCover struct {
	Title  string       `yaml:"title" json:"title"`
	Date   time.Time    `yaml:"date" json:"date"`
	Author model.Author `yaml:"author" json:"author"`
}

type payloadOverview struct {
	ManagementSummaryComment string         `yaml:"management_summary_comment" json:"management_summary_comment"`
	BusinessCriticality      string         `yaml:"business_criticality" json:"business_criticality"`
	BusinessOverview         model.Overview `yaml:"business_overview" json:"business_overview"`
	TechnicalOverview        model.Overview `yaml:"technical_overview" json:"technical_overview"`
}

type payloadAbuseCases map[string]string

type payloadSecurityRequirements map[string]string

type payloadDataAsset struct {
	Title                  string   `yaml:"title" json:"title"`
	Id                     string   `yaml:"id" json:"id"`
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

type payloadSharedRuntime struct {
	Title                  string   `yaml:"title" json:"title"`
	Id                     string   `yaml:"id" json:"id"`
	Description            string   `yaml:"description" json:"description"`
	Tags                   []string `yaml:"tags" json:"tags"`
	TechnicalAssetsRunning []string `yaml:"technical_assets_running" json:"technical_assets_running"`
}

var throttlerLock sync.Mutex
var createdObjectsThrottler = make(map[string][]int64)

var locksByFolderName = make(map[string]*sync.Mutex)

type tokenHeader struct {
	Token string `header:"token"`
}
type keyHeader struct {
	Key string `header:"key"`
}

func printTypes(title string, value interface{}) {
	fmt.Println(fmt.Sprintf("  %v: %v", title, value))
}

// explainTypes prints and explanation block and a header
func printExplainTypes(title string, value []types.TypeEnum) {
	fmt.Println(title)
	for _, candidate := range value {
		fmt.Printf("\t %v: %v\n", candidate, candidate.Explain())
	}
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer func() { _ = source.Close() }()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer func() { _ = destination.Close() }()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func lowerCaseAndTrim(tags []string) []string {
	for i := range tags {
		tags[i] = strings.ToLower(strings.TrimSpace(tags[i]))
	}
	return tags
}

func checkTags(tags []string, where string) []string {
	var tagsUsed = make([]string, 0)
	if tags != nil {
		tagsUsed = make([]string, len(tags))
		for i, parsedEntry := range tags {
			referencedTag := fmt.Sprintf("%v", parsedEntry)
			checkTagExists(referencedTag, where)
			tagsUsed[i] = referencedTag
		}
	}
	return tagsUsed
}

// in order to prevent Path-Traversal like stuff...
func removePathElementsFromImageFiles(overview model.Overview) model.Overview {
	for i := range overview.Images {
		newValue := make(map[string]string)
		for file, desc := range overview.Images[i] {
			newValue[filepath.Base(file)] = desc
		}
		overview.Images[i] = newValue
	}
	return overview
}

func hasNotYetAnyDirectNonWildcardRiskTracking(syntheticRiskId string) bool {
	if _, ok := model.ParsedModelRoot.RiskTracking[syntheticRiskId]; ok {
		return false
	}
	return true
}

func withDefault(value string, defaultWhenEmpty string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > 0 && trimmed != "<nil>" {
		return trimmed
	}
	return strings.TrimSpace(defaultWhenEmpty)
}

func createDataFlowId(sourceAssetId, title string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	checkErr(err)
	return sourceAssetId + ">" + strings.Trim(reg.ReplaceAllString(strings.ToLower(title), "-"), "- ")
}

func createSyntheticId(categoryId string,
	mostRelevantDataAssetId, mostRelevantTechnicalAssetId, mostRelevantCommunicationLinkId, mostRelevantTrustBoundaryId, mostRelevantSharedRuntimeId string) string {
	result := categoryId
	if len(mostRelevantTechnicalAssetId) > 0 {
		result += "@" + mostRelevantTechnicalAssetId
	}
	if len(mostRelevantCommunicationLinkId) > 0 {
		result += "@" + mostRelevantCommunicationLinkId
	}
	if len(mostRelevantTrustBoundaryId) > 0 {
		result += "@" + mostRelevantTrustBoundaryId
	}
	if len(mostRelevantSharedRuntimeId) > 0 {
		result += "@" + mostRelevantSharedRuntimeId
	}
	if len(mostRelevantDataAssetId) > 0 {
		result += "@" + mostRelevantDataAssetId
	}
	return result
}

func checkTagExists(referencedTag, where string) {
	if !model.Contains(model.ParsedModelRoot.TagsAvailable, referencedTag) {
		panic(errors.New("missing referenced tag in overall tag list at " + where + ": " + referencedTag))
	}
}

func checkDataAssetTargetExists(referencedAsset, where string) {
	if _, ok := model.ParsedModelRoot.DataAssets[referencedAsset]; !ok {
		panic(errors.New("missing referenced data asset target at " + where + ": " + referencedAsset))
	}
}

func checkTrustBoundaryExists(referencedId, where string) {
	if _, ok := model.ParsedModelRoot.TrustBoundaries[referencedId]; !ok {
		panic(errors.New("missing referenced trust boundary at " + where + ": " + referencedId))
	}
}

func checkSharedRuntimeExists(referencedId, where string) {
	if _, ok := model.ParsedModelRoot.SharedRuntimes[referencedId]; !ok {
		panic(errors.New("missing referenced shared runtime at " + where + ": " + referencedId))
	}
}

func checkCommunicationLinkExists(referencedId, where string) {
	if _, ok := model.CommunicationLinks[referencedId]; !ok {
		panic(errors.New("missing referenced communication link at " + where + ": " + referencedId))
	}
}

func checkTechnicalAssetExists(referencedAsset, where string, onlyForTweak bool) {
	if _, ok := model.ParsedModelRoot.TechnicalAssets[referencedAsset]; !ok {
		suffix := ""
		if onlyForTweak {
			suffix = " (only referenced in diagram tweak)"
		}
		panic(errors.New("missing referenced technical asset target" + suffix + " at " + where + ": " + referencedAsset))
	}
}

func checkNestedTrustBoundariesExisting() {
	for _, trustBoundary := range model.ParsedModelRoot.TrustBoundaries {
		for _, nestedId := range trustBoundary.TrustBoundariesNested {
			if _, ok := model.ParsedModelRoot.TrustBoundaries[nestedId]; !ok {
				panic(errors.New("missing referenced nested trust boundary: " + nestedId))
			}
		}
	}
}

func hash(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%v", h.Sum32())
}

func makeDiagramSameRankNodeTweaks() string {
	// see https://stackoverflow.com/questions/25734244/how-do-i-place-nodes-on-the-same-level-in-dot
	tweak := ""
	if len(model.ParsedModelRoot.DiagramTweakSameRankAssets) > 0 {
		for _, sameRank := range model.ParsedModelRoot.DiagramTweakSameRankAssets {
			assetIDs := strings.Split(sameRank, ":")
			if len(assetIDs) > 0 {
				tweak += "{ rank=same; "
				for _, id := range assetIDs {
					checkTechnicalAssetExists(id, "diagram tweak same-rank", true)
					if len(model.ParsedModelRoot.TechnicalAssets[id].GetTrustBoundaryId()) > 0 {
						panic(errors.New("technical assets (referenced in same rank diagram tweak) are inside trust boundaries: " +
							fmt.Sprintf("%v", model.ParsedModelRoot.DiagramTweakSameRankAssets)))
					}
					tweak += " " + hash(id) + "; "
				}
				tweak += " }"
			}
		}
	}
	return tweak
}

func makeTechAssetNode(technicalAsset model.TechnicalAsset, simplified bool) string {
	if simplified {
		color := colors.RgbHexColorOutOfScope()
		if !technicalAsset.OutOfScope {
			generatedRisks := technicalAsset.GeneratedRisks()
			switch model.HighestSeverityStillAtRisk(generatedRisks) {
			case types.CriticalSeverity:
				color = colors.RgbHexColorCriticalRisk()
			case types.HighSeverity:
				color = colors.RgbHexColorHighRisk()
			case types.ElevatedSeverity:
				color = colors.RgbHexColorElevatedRisk()
			case types.MediumSeverity:
				color = colors.RgbHexColorMediumRisk()
			case types.LowSeverity:
				color = colors.RgbHexColorLowRisk()
			default:
				color = "#444444" // since black is too dark here as fill color
			}
			if len(model.ReduceToOnlyStillAtRisk(generatedRisks)) == 0 {
				color = "#444444" // since black is too dark here as fill color
			}
		}
		return "  " + hash(technicalAsset.Id) + ` [ shape="box" style="filled" fillcolor="` + color + `"
				label=<<b>` + encode(technicalAsset.Title) + `</b>> penwidth="3.0" color="` + color + `" ];
				`
	} else {
		var shape, title string
		var lineBreak = ""
		switch technicalAsset.Type {
		case types.ExternalEntity:
			shape = "box"
			title = technicalAsset.Title
		case types.Process:
			shape = "ellipse"
			title = technicalAsset.Title
		case types.Datastore:
			shape = "cylinder"
			title = technicalAsset.Title
			if technicalAsset.Redundant {
				lineBreak = "<br/>"
			}
		}

		if technicalAsset.UsedAsClientByHuman {
			shape = "octagon"
		}

		// RAA = Relative Attacker Attractiveness
		raa := technicalAsset.RAA
		var attackerAttractivenessLabel string
		if technicalAsset.OutOfScope {
			attackerAttractivenessLabel = "<font point-size=\"15\" color=\"#603112\">RAA: out of scope</font>"
		} else {
			attackerAttractivenessLabel = "<font point-size=\"15\" color=\"#603112\">RAA: " + fmt.Sprintf("%.0f", raa) + " %</font>"
		}

		compartmentBorder := "0"
		if technicalAsset.MultiTenant {
			compartmentBorder = "1"
		}

		return "  " + hash(technicalAsset.Id) + ` [
	label=<<table border="0" cellborder="` + compartmentBorder + `" cellpadding="2" cellspacing="0"><tr><td><font point-size="15" color="` + colors.DarkBlue + `">` + lineBreak + technicalAsset.Technology.String() + `</font><br/><font point-size="15" color="` + colors.LightGray + `">` + technicalAsset.Size.String() + `</font></td></tr><tr><td><b><font color="` + technicalAsset.DetermineLabelColor() + `">` + encode(title) + `</font></b><br/></td></tr><tr><td>` + attackerAttractivenessLabel + `</td></tr></table>>
	shape=` + shape + ` style="` + technicalAsset.DetermineShapeBorderLineStyle() + `,` + technicalAsset.DetermineShapeStyle() + `" penwidth="` + technicalAsset.DetermineShapeBorderPenWidth() + `" fillcolor="` + technicalAsset.DetermineShapeFillColor() + `"
	peripheries=` + strconv.Itoa(technicalAsset.DetermineShapePeripheries()) + `
	color="` + technicalAsset.DetermineShapeBorderColor() + "\"\n  ]; "
	}
}

func makeDataAssetNode(dataAsset model.DataAsset) string {
	var color string
	switch dataAsset.IdentifiedDataBreachProbabilityStillAtRisk() {
	case types.Probable:
		color = colors.RgbHexColorHighRisk()
	case types.Possible:
		color = colors.RgbHexColorMediumRisk()
	case types.Improbable:
		color = colors.RgbHexColorLowRisk()
	default:
		color = "#444444" // since black is too dark here as fill color
	}
	if !dataAsset.IsDataBreachPotentialStillAtRisk() {
		color = "#444444" // since black is too dark here as fill color
	}
	return "  " + hash(dataAsset.Id) + ` [ label=<<b>` + encode(dataAsset.Title) + `</b>> penwidth="3.0" style="filled" fillcolor="` + color + `" color="` + color + "\"\n  ]; "
}

func encode(value string) string {
	return strings.ReplaceAll(value, "&", "&amp;")
}
