package initcmd

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"os/user"
	"path/filepath"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/fileprovider"
	"github.com/c00/buttercup/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the initial configuration file",
	Args:  cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		u, err := user.Current()
		if err != nil {
			panic("cannot get current user for some reason")
		}

		configPath := filepath.Join(u.HomeDir, ".buttercup", appconfig.ConfigFile)

		_, err = os.Stat(configPath)
		if err == nil {
			logger.Log("Configuration file already exists at %v", configPath)
			return
		}

		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown-device"
		}

		bytes := make([]byte, 32)
		_, err = rand.Read(bytes)
		if err != nil {
			logger.Error("cannot generate random passphrase: %v", err)
			return
		}
		passphrase := base64.RawStdEncoding.EncodeToString(bytes)

		config := appconfig.AppConfig{
			DefaultFolder: "default",
			ClientName:    hostname,
			Folders: []appconfig.FolderConfig{
				{
					Name: "default",
					Local: appconfig.ProviderConfig{
						Type: fileprovider.TypeFs,
						FsConfig: &appconfig.FsProviderConfig{
							Path: filepath.Join(u.HomeDir, "Buttercup"),
						},
					},
					Remote: appconfig.ProviderConfig{
						Type: fileprovider.TypeS3,
						S3Config: &appconfig.S3ProviderConfig{
							Passphrase: passphrase,
							AccessKey:  "youraccesskey",
							SecretKey:  "yoursecretkey",
							BasePath:   "Buttercup-files",
							Bucket:     "youtbucket",
							Endpoint:   "yourendpoint",
							Region:     "yourregion",
						},
					},
				},
			},
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			logger.Error2(err)
			return
		}

		err = os.WriteFile(configPath, data, 0600)
		if err != nil {
			logger.Error2(err)
			return
		}

		logger.Log("Configuration file written at: %v", configPath)
		logger.Log("To get started, you have to configure some things. See this guide for more details: https://github.com/c00/buttercup/blob/main/guides/installation.md")
	},
}
