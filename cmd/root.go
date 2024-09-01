package main

import (
	"fmt"
	"os"

	"github.com/c00/buttercup/cmd/pullcmd"
	pushcmd "github.com/c00/buttercup/cmd/pushCmd"
	synccmd "github.com/c00/buttercup/cmd/syncCmd"
	"github.com/c00/buttercup/logger"
	"github.com/spf13/cobra"
)

var verbosity int

func init() {
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase verbosity")

	rootCmd.AddCommand(
		versionCmd,
		pullcmd.PullCmd,
		pushcmd.PushCmd,
		synccmd.SyncCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:   binary,
	Short: fmt.Sprintf("%v is a tool for syncing folders securely over the internet.", binary),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verboseFlags, err := cmd.Flags().GetCount("verbose")
		if err != nil {
			return
		}

		if verboseFlags == 0 {
			return
		}

		if verboseFlags > 5 {
			verboseFlags = 5
		}

		logger.IncreaseLevel(verboseFlags)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
