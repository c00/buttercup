package syncer

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/c00/buttercup/fileprovider"
	"github.com/stretchr/testify/assert"
)

func TestPullNewFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	syncer := New(local, remote)

	//setup starting situation
	path := "/foo.txt"
	assert.Nil(t, setFileContent(remote, fileprovider.FileInfo{Path: path, Updated: getDate(1)}, "remote"))

	err := syncer.Pull()
	assert.Nil(t, err)

	//Check that the new state is achieved.
	assert.True(t, fileHasContent(local, path, "remote"))
	assert.True(t, fileHasContent(remote, path, "remote"))
}

func TestPullUpdatedFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	syncer := New(local, remote)

	path := "/foo.txt"

	assert.Nil(t, setFileContent(local, fileprovider.FileInfo{Path: path, Updated: getDate(0), LastSynced: getDate(0)}, "source"))
	assert.Nil(t, setFileContent(remote, fileprovider.FileInfo{Path: path, Updated: getDate(1)}, "remote"))

	err := syncer.Pull()
	assert.Nil(t, err)

	newFi, err := local.GetFileInfo(path)
	assert.Nil(t, err)
	assert.True(t, newFi.Updated.Equal(getDate(1)))

	assert.True(t, fileHasContent(local, path, "remote"))
}

func TestPullConflict1(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	syncer := New(local, remote)

	path := "/foo.txt"
	conflictPath := "/foo.conflict.txt"

	assert.Nil(t, setFileContent(local, fileprovider.FileInfo{Path: path, Updated: getDate(1), LastSynced: getDate(0)}, "source"))
	assert.Nil(t, setFileContent(remote, fileprovider.FileInfo{Path: path, Updated: getDate(2)}, "remote"))

	err := syncer.Pull()
	assert.Nil(t, err)

	newFi, err := local.GetFileInfo(path)
	assert.Nil(t, err)
	assert.True(t, newFi.Updated.Equal(getDate(2)))
	assert.True(t, fileHasContent(local, path, "remote"))

	//Check conflict file
	newFi, err = local.GetFileInfo(conflictPath)
	assert.Nil(t, err)
	assert.True(t, newFi.Updated.Equal(getDate(1)))
	assert.True(t, fileHasContent(local, conflictPath, "source"))
}

func TestPullConflict2(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	syncer := New(local, remote)

	path := "/foo.txt"
	conflictPath := "/foo.conflict.txt"

	assert.Nil(t, setFileContent(local, fileprovider.FileInfo{Path: path, Updated: getDate(2), LastSynced: getDate(0)}, "source"))
	assert.Nil(t, setFileContent(remote, fileprovider.FileInfo{Path: path, Updated: getDate(1)}, "remote"))

	err := syncer.Pull()
	assert.Nil(t, err)

	newFi, err := local.GetFileInfo(path)
	assert.Nil(t, err)
	assert.True(t, newFi.Updated.Equal(getDate(2)))
	assert.True(t, fileHasContent(local, path, "source"))

	//Check conflict file
	newFi, err = local.GetFileInfo(conflictPath)
	assert.Nil(t, err)
	assert.True(t, newFi.Updated.Equal(getDate(1)))
	assert.True(t, fileHasContent(local, conflictPath, "remote"))
}

func TestDontPullUpToDateFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	syncer := New(local, remote)

	path := "/foo.txt"

	assert.Nil(t, setFileContent(local, fileprovider.FileInfo{Path: path, Updated: getDate(1), LastSynced: getDate(0)}, "source"))
	assert.Nil(t, setFileContent(remote, fileprovider.FileInfo{Path: path, Updated: getDate(0)}, "remote"))

	err := syncer.Pull()
	assert.Nil(t, err)

	newFi, err := local.GetFileInfo(path)
	assert.Nil(t, err)
	assert.True(t, newFi.Updated.Equal(getDate(1)))

	assert.True(t, fileHasContent(local, path, "source"))
}

func fileHasContent(fp fileprovider.FileProvider, path string, content string) bool {
	reader, err := fp.RetrieveFile(path)
	if err != nil {
		return false
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return false
	}
	return string(data) == content
}

func TestPushNewFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	syncer := New(local, remote)

	//setup starting situation
	path := "/foo.txt"
	assert.Nil(t, setFileContent(local, fileprovider.FileInfo{Path: path, Updated: getDate(1)}, "local"))

	err := syncer.Push()
	assert.Nil(t, err)

	//Check that the new state is achieved.
	assert.True(t, fileHasContent(local, path, "local"))
	assert.True(t, fileHasContent(remote, path, "local"))
}

func TestPushUpdatedFile(t *testing.T) {
	local := fileprovider.NewInMemoryProvider("client")
	remote := fileprovider.NewInMemoryProvider("client")
	syncer := New(local, remote)

	//setup starting situation
	path := "/foo.txt"
	assert.Nil(t, setFileContent(local, fileprovider.FileInfo{Path: path, Updated: getDate(1), LastSynced: getDate(0)}, "local"))
	assert.Nil(t, setFileContent(remote, fileprovider.FileInfo{Path: path, Updated: getDate(0)}, "remote"))

	err := syncer.Push()
	assert.Nil(t, err)

	//Check that the new state is achieved.
	assert.True(t, fileHasContent(remote, path, "local"))
}

func setFileContent(fp fileprovider.FileProvider, fi fileprovider.FileInfo, content string) error {
	err := fp.StoreFile(fi, strings.NewReader(content))
	if err != nil {
		return err
	}

	if !fi.LastSynced.Equal(time.Time{}) {
		return fp.SetLastSynced(fi.Path, fi.LastSynced)
	}
	return nil
}

func getDate(hourOffset int) time.Time {
	return time.Date(2020, 6, 10, 12, 0, 0, 0, time.UTC).Add(time.Hour * time.Duration(hourOffset))
}
