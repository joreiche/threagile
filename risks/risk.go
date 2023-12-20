package risks

import (
	"log"

	"github.com/threagile/threagile/pkg/internal"
	"github.com/threagile/threagile/pkg/model"
)

type BuiltInRisk struct {
	Category      func() model.RiskCategory
	SupportedTags func() []string
	GenerateRisks func(m *model.ParsedModel) []model.Risk
}

type CustomRisk struct {
	ID       string
	Category model.RiskCategory
	Tags     []string
	Runner   *internal.Runner
}

func (r *CustomRisk) GenerateRisks(m *model.ParsedModel) []model.Risk {
	if r.Runner == nil {
		return nil
	}

	risks := make([]model.Risk, 0)
	runError := r.Runner.Run(m, &risks, "-generate-risks")
	if runError != nil {
		log.Fatalf("Failed to generate risks for custom risk rule %q: %v\n", r.Runner.Filename, runError)
	}

	return risks
}
