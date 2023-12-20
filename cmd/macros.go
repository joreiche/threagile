/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/threagile/threagile/pkg/docs"
	builinmacros "github.com/threagile/threagile/pkg/macros/built-in"
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
		for _, macros := range builinmacros.ListBuiltInMacros() {
			cmd.Println(macros.ID, "-->", macros.Title)
		}
		cmd.Println()
	},
}

var explainMacrosCmd = &cobra.Command{
	Use:   "explain-model-macros",
	Short: "Explain model macros",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(docs.Logo + "\n\n" + docs.VersionText)
		cmd.Println("Explanation for the model macros:")
		cmd.Println()
		/* TODO finish plugin stuff
		cmd.Println("Custom model macros:")
		for id, customModelMacro := range macros.ListCustomMacros() {
			cmd.Printf("%v: %v\n", macros.ID, macros.Title)
		}
		cmd.Println()
		*/
		cmd.Println("----------------------")
		cmd.Println("Built-in model macros:")
		cmd.Println("----------------------")
		for _, macros := range builinmacros.ListBuiltInMacros() {
			cmd.Printf("%v: %v\n", macros.ID, macros.Title)
		}

		cmd.Println()
		/*
			fmt.Printf("%v: %v\n", addbuildpipeline.GetMacroDetails().ID, addbuildpipeline.GetMacroDetails().Description)
			fmt.Printf("%v: %v\n", addvault.GetMacroDetails().ID, addvault.GetMacroDetails().Description)
			fmt.Printf("%v: %v\n", prettyprint.GetMacroDetails().ID, prettyprint.GetMacroDetails().Description)
			fmt.Printf("%v: %v\n", removeunusedtags.GetMacroDetails().ID, removeunusedtags.GetMacroDetails().Description)
			fmt.Printf("%v: %v\n", seedrisktracking.GetMacroDetails().ID, seedrisktracking.GetMacroDetails().Description)
			fmt.Printf("%v: %v\n", seedtags.GetMacroDetails().ID, seedtags.GetMacroDetails().Description)
			fmt.Println()*/
	},
}
