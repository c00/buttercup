package fileprovider

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/internal/fstests"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestFsProvider_RunProviderSuite(t *testing.T) {
	godotenv.Load("../.env")
	sourcePath := os.Getenv("TEST_SOURCE_PATH")

	RunSuite(t, func() FileProvider {
		fstests.SetupSourceFilesystem(sourcePath, false)
		return NewFsProvider(appconfig.ProviderConfig{
			Type:     TypeFs,
			FsConfig: &appconfig.FsProviderConfig{Path: sourcePath},
		})
	})
}

func TestIndexFileInIndex(t *testing.T) {
	godotenv.Load("../.env")
	sourcePath := os.Getenv("TEST_SOURCE_PATH")

	fstests.SetupSourceFilesystem(sourcePath, false)

	NewFsProvider(appconfig.ProviderConfig{
		Type:     TypeFs,
		FsConfig: &appconfig.FsProviderConfig{Path: sourcePath},
	})

	p := NewFsProvider(appconfig.ProviderConfig{
		Type:     TypeFs,
		FsConfig: &appconfig.FsProviderConfig{Path: sourcePath},
	})

	_, err := p.getFileInfo(sqliteIndexName)
	assert.NotNil(t, err)
}

func TestFsProvider_RetrieveFile(t *testing.T) {
	godotenv.Load("../.env")
	sourcePath := os.Getenv("TEST_SOURCE_PATH")

	fstests.SetupSourceFilesystem(sourcePath, true)

	p := NewFsProvider(appconfig.ProviderConfig{
		Type:     TypeFs,
		FsConfig: &appconfig.FsProviderConfig{Path: sourcePath},
	})

	filename := "foo.txt"

	reader, err := p.RetrieveFile(filename)
	assert.Nil(t, err)
	defer reader.Close()

	//read the stream
	bytes, err := io.ReadAll(reader)
	assert.Nil(t, err)
	assert.Equal(t, string(bytes), "content for file: /foo.txt")

	//Get a non-existent file stream
	_, err = p.RetrieveFile("not-a-real-file.txt")
	assert.NotNil(t, err)
}

func TestFsProvider_StoreFile(t *testing.T) {
	godotenv.Load("../.env")
	sourcePath := os.Getenv("TEST_SOURCE_PATH")

	fstests.SetupSourceFilesystem(sourcePath, false)

	p := NewFsProvider(appconfig.ProviderConfig{
		Type:     TypeFs,
		FsConfig: &appconfig.FsProviderConfig{Path: sourcePath},
	})

	filename := "foo.txt"

	//Create a stream
	reader := strings.NewReader("some content")

	err := p.StoreFile(FileInfo{Path: filename}, reader)
	assert.Nil(t, err)

	//read the stream
	data, err := os.ReadFile(path.Join(sourcePath, filename))
	assert.Nil(t, err)
	assert.Equal(t, string(data), "some content")
}
