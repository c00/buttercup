package fileprovider

import (
	"os"
	"testing"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/internal/fstests"
	"github.com/joho/godotenv"
)

func TestEfsProvider_RunProviderSuite(t *testing.T) {
	godotenv.Load("../.env")
	sourcePath := os.Getenv("TEST_SOURCE_PATH")

	RunSuite(t, func() FileProvider {
		fstests.SetupSourceFilesystem(sourcePath, false)
		return NewEfsProvider(appconfig.ProviderConfig{
			Type:      TypeFs,
			EfsConfig: &appconfig.EfsProviderConfig{Path: sourcePath, Passphrase: "foo"},
		})
	})
}
