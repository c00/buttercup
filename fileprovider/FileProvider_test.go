package fileprovider

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func RunSuite(t *testing.T, pf func() FileProvider) {
	storeAndRetrieveFile(t, pf())
	checkLocks(t, pf())
	getFileInfo(t, pf())
	getInfosPaging(t, pf())
	renaming(t, pf())
	updateSyncDate(t, pf())
	storeDeleted(t, pf())
}

func storeDeleted(t *testing.T, p FileProvider) {
	path := "/foo.txt"
	assert.Nil(t, p.RemoveFile(FileInfo{Path: path, Deleted: true}))

	fi, err := p.GetFileInfo(path)
	assert.Nil(t, err)
	assert.True(t, fi.Deleted)

	//Now store a file over it
	assert.Nil(t, p.StoreFile(FileInfo{Path: path}, strings.NewReader("foo")))

	fi, err = p.GetFileInfo(path)
	assert.Nil(t, err)
	assert.False(t, fi.Deleted)

	//Now delete it again
	assert.Nil(t, p.RemoveFile(FileInfo{Path: path, Deleted: true}))

	fi, err = p.GetFileInfo(path)
	assert.Nil(t, err)
	assert.True(t, fi.Deleted)
}

func updateSyncDate(t *testing.T, p FileProvider) {
	path := "/foo.txt"
	assert.Nil(t, p.StoreFile(FileInfo{Path: path}, strings.NewReader("foo")))

	newDate := time.Date(2020, 2, 2, 2, 2, 2, 2, time.UTC)
	err := p.SetLastSynced(path, newDate)
	assert.Nil(t, err)

	fi, err := p.GetFileInfo(path)
	assert.Nil(t, err)
	assert.True(t, fi.LastSynced.Equal(newDate))
}

func renaming(t *testing.T, p FileProvider) {
	oldPath := "/foo.txt"
	newPath := "/somewhere/else/bar.txt"
	assert.Nil(t, p.StoreFile(FileInfo{Path: oldPath}, strings.NewReader("foo")))

	err := p.MoveFile(oldPath, newPath)
	assert.Nil(t, err)

	_, err = p.GetFileInfo(oldPath)
	assert.NotNil(t, err)

	_, err = p.GetFileInfo(newPath)
	assert.Nil(t, err)

}

func getInfosPaging(t *testing.T, p FileProvider) {
	assert.Nil(t, p.StoreFile(FileInfo{Path: "/foo1.txt"}, strings.NewReader("foo")))
	assert.Nil(t, p.StoreFile(FileInfo{Path: "/foo2.txt"}, strings.NewReader("foo")))
	assert.Nil(t, p.StoreFile(FileInfo{Path: "/foo3.txt"}, strings.NewReader("foo")))
	assert.Nil(t, p.StoreFile(FileInfo{Path: "/foo4.txt"}, strings.NewReader("foo")))

	type testCase struct {
		offset         int
		limit          int
		expectedLength int
	}
	cases := []testCase{
		{offset: 0, limit: 0, expectedLength: 4},
		{offset: 0, limit: 2, expectedLength: 2},
		{offset: 2, limit: 1, expectedLength: 1},
		{offset: 0, limit: 100, expectedLength: 4},
		{offset: 100, limit: 0, expectedLength: 0},
		{offset: 100, limit: 100, expectedLength: 0},
	}

	for _, tst := range cases {
		gotten, err := p.GetFileInfos(tst.limit, tst.offset)
		assert.Nil(t, err)
		assert.Len(t, gotten, tst.expectedLength)
	}
}

func getFileInfo(t *testing.T, p FileProvider) {
	err := p.StoreFile(FileInfo{Path: "/foo.txt"}, strings.NewReader("foo"))
	assert.Nil(t, err)
	_, err = p.GetFileInfo("/foo.txt")
	assert.Nil(t, err)
}

func checkLocks(t *testing.T, p FileProvider) {
	err := p.Lock()
	assert.Nil(t, err)

	err = p.Lock()
	assert.NotNil(t, err, "cannot lock twice")

	//Try writing when I am locking it
	err = p.StoreFile(FileInfo{Path: "/foo.txt"}, strings.NewReader("foo"))
	assert.Nil(t, err)

	//Try Deleting when I am locking it
	err = p.RemoveFile(FileInfo{Path: "/foo.txt"})
	assert.Nil(t, err)

	err = p.Unlock()
	assert.Nil(t, err)

	err = p.Unlock()
	assert.NotNil(t, err, "cannot unlock twice")

}

func storeAndRetrieveFile(t *testing.T, p FileProvider) {
	filename := "/foo.txt"
	content := "some content for automated tests"

	//Store a file
	storeRdr := strings.NewReader(content)
	err := p.StoreFile(FileInfo{Path: filename}, storeRdr)
	assert.Nil(t, err)

	retrieveRdr, err := p.RetrieveFile(filename)
	assert.Nil(t, err)
	defer retrieveRdr.Close()

	//read the stream
	bytes, err := io.ReadAll(retrieveRdr)
	assert.Nil(t, err)
	assert.Equal(t, string(bytes), content)

	//Delete the file
	err = p.RemoveFile(FileInfo{Path: filename})
	assert.Nil(t, err)

	//Delete non-existent file should not give an error
	err = p.RemoveFile(FileInfo{Path: filename})
	assert.Nil(t, err)

	//Get a non-existent file stream
	_, err = p.RetrieveFile("not-a-real-file.txt")
	assert.NotNil(t, err)
}
