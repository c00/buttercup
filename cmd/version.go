package main

import (
	"github.com/c00/buttercup/logger"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Log("buttercup %v\n", version)
	},
}
