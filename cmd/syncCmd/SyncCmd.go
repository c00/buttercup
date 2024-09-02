package synccmd

import (
	"fmt"
	"os"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/fileprovider"
	"github.com/c00/buttercup/logger"
	"github.com/c00/buttercup/syncer"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync files between local and remote",
	Args:  cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := appconfig.LoadFromUser()
		if err != nil {
			panic(fmt.Errorf("cannot load config: %w", err))
		}

		folderName := conf.DefaultFolder
		if len(args) == 1 {
			folderName = args[0]
		}

		folder := conf.GetFolder(folderName)
		logger.Log("Syncing folder: %v...", folder.Local.GetFolderPath())

		local := fileprovider.GetProvider(folder.Local)
		remote := fileprovider.GetProvider(folder.Remote)

		syncer := syncer.New(local, remote)

		logger.Log("Pulling changes from the remote...")
		err = syncer.Pull()
		if err != nil {
			logger.Error2(err)
			os.Exit(1)
		}

		logger.Log("Pushing local changes to the remote...")
		err = syncer.Push()
		if err != nil {
			logger.Error2(err)
			os.Exit(1)
		}
	},
}
