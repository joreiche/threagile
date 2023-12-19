/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/threagile/threagile/pkg/docs"
	"github.com/threagile/threagile/pkg/macros"
)

var listMacrosCmd = &cobra.Command{
	Use:   "list-model-macros",
	Short: "Print model macros",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(docs.Logo + "\n\n" + docs.VersionText)
		cmd.Println("The following model macros are available (can be extended via custom model macros):")
		cmd.Println()
		/* TODO finish plugin stuff
		cmd.Println("Custom model macros:")
		for id, customModelMacro := range macros.ListCustomMacros() {
			cmd.Println(id, "-->", customModelMacro.GetMacroDetails().Title)
		}
		cmd.Println()
		*/
		cmd.Println("----------------------")
		cmd.Println("Built-in model macros:")
		cmd.Println("----------------------")
		for _, macros := range macros.ListBuiltInMacros() {
			cmd.Println(macros.ID, "-->", macros.Title)
		}
		cmd.Println()
	},
}
