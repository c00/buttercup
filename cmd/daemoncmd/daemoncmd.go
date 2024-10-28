// deamoncmd is the command for running as a service
package daemoncmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"sync"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/fileprovider"
	"github.com/c00/buttercup/internal/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var log = logger.New("Daemon")

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the service that watches your filesystem and updates automatically.",
	Args:  cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := appconfig.LoadFromUser()
		if err != nil {
			panic(fmt.Errorf("cannot load config: %w", err))
		}

		ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

		wg := &sync.WaitGroup{}
		//Start watcher for each folder
		for _, f := range conf.Folders {
			if f.Local.Type != fileprovider.TypeFs {
				log.Info("not watching folder of type %v", f.Local.Type)
				continue
			}

			wg.Add(1)
			//Start new wachter
			go startWatcher(ctx, wg, f.Local)
		}

		wg.Wait()

		log.Debug("Daemon stopped gracefully")
	},
}

func startWatcher(ctx context.Context, wg *sync.WaitGroup, f appconfig.ProviderConfig) {
	defer wg.Done()

	//Create filewatcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("could not create new file watcher: %v", err)
		return
	}
	defer watcher.Close()

	//todo have a max depth or a max number of dirs
	err = BreadthFirstDirWalk(f.FsConfig.Path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("cannot get dirinfo: %w", err)
		}

		err = watcher.Add(path)
		if err != nil {
			return fmt.Errorf("could not add folder to watcher: %w", err)
		}
		log.Debug("Monitoring folder: %v", path)

		return nil
	})

	if err != nil {
		log.Error("could not watch folders: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Debug("Stopping file watcher...")
			return
		case event, ok := <-watcher.Events:
			if !ok {
				log.Error("watcher event was not ok for some reason")
				continue
			}
			log.Log("TODO respond to event: %+v", event)
		}

	}
}
