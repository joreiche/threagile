/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package macros

import (
	addbuildpipeline "github.com/threagile/threagile/pkg/macros/built-in/add-build-pipeline"
	addvault "github.com/threagile/threagile/pkg/macros/built-in/add-vault"
	prettyprint "github.com/threagile/threagile/pkg/macros/built-in/pretty-print"
	removeunusedtags "github.com/threagile/threagile/pkg/macros/built-in/remove-unused-tags"
	seedrisktracking "github.com/threagile/threagile/pkg/macros/built-in/seed-risk-tracking"
	seedtags "github.com/threagile/threagile/pkg/macros/built-in/seed-tags"

	"github.com/threagile/threagile/model"
)

func ListBuiltInMacros() []model.MacroDetails {
	return []model.MacroDetails{
		addbuildpipeline.GetMacroDetails(),
		addvault.GetMacroDetails(),
		prettyprint.GetMacroDetails(),
		removeunusedtags.GetMacroDetails(),
		seedrisktracking.GetMacroDetails(),
		seedtags.GetMacroDetails(),
	}
}

func ListCustomeMacros() []model.MacroDetails {
	// TODO: implement
	return []model.MacroDetails{}
}
