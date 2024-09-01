package source

import (
	"strings"
	"testing"
	"time"

	"github.com/c00/buttercup/fileprovider"
	"github.com/stretchr/testify/assert"
)

func TestSource_PullFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	source := NewSource(local, remote)

	filePath := "/foo.txt"
	fileContent := "some test content for foo"
	fileDate := time.Date(2024, 01, 01, 12, 00, 00, 00, time.UTC)

	remoteFi := fileprovider.FileInfo{Path: filePath, Updated: fileDate}

	remote.StoreFile(remoteFi, strings.NewReader(fileContent))

	//Pull it to the local
	err := source.PullFile(remoteFi, filePath)
	assert.Nil(t, err)

	//confirm it exists locally
	fi, err := local.GetFileInfo(filePath)
	assert.Nil(t, err)
	assert.Equal(t, fi.Path, filePath)
	assert.Equal(t, fi.Updated, fileDate)

	//Write as new file
	localFilePath := "/foo.conflict.txt"
	err = source.PullFile(remoteFi, localFilePath)
	assert.Nil(t, err)

	fi2, err := local.GetFileInfo(localFilePath)
	assert.Nil(t, err)
	assert.Equal(t, fi2.Path, localFilePath)
	assert.Equal(t, fi2.Updated, fileDate)
}

func TestSource_PushFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	source := NewSource(local, remote)

	filePath := "/foo.txt"
	fileContent := "some test content for foo"
	fileDate := time.Date(2024, 01, 01, 12, 00, 00, 00, time.UTC)

	localFi := fileprovider.FileInfo{Path: filePath, Updated: fileDate}

	local.StoreFile(localFi, strings.NewReader(fileContent))

	//Pull it to the local
	err := source.PushFile(localFi)
	assert.Nil(t, err)

	//confirm it exists Remotelu
	fi, err := remote.GetFileInfo(filePath)
	assert.Nil(t, err)
	assert.Equal(t, fi.Path, filePath)
	assert.Equal(t, fi.Updated, fileDate)
}

func TestSource_PullNewDeletedFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	source := NewSource(local, remote)

	filePath := "/foo.txt"
	fileDate := time.Date(2024, 01, 01, 12, 00, 00, 00, time.UTC)

	remoteFi := fileprovider.FileInfo{Path: filePath, Updated: fileDate, Deleted: true}

	//Sync deleted state to local before it was ever created
	err := source.PullFile(remoteFi, filePath)
	assert.Nil(t, err)

	//confirm it exists locally and is deleted
	fi, err := local.GetFileInfo(filePath)
	assert.Nil(t, err)
	assert.Equal(t, fi.Path, filePath)
	assert.Equal(t, fi.Updated, fileDate)
	assert.True(t, fi.Deleted)
}
