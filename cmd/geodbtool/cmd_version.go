package main

import (
	"github.com/anexia-it/geodbtools"
	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print version and license information",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("%s v%s\n", cmd.Root().Use, geodbtools.VersionString())
		cmd.Println("Copyright (C) 2019 Anexia Internetdienstleistungs GmbH")
		cmd.Println("License: MIT")
	},
}

func init() {
	cmdRoot.AddCommand(cmdVersion)
}
