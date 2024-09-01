package fsindex

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestBasics(t *testing.T) {
	db := New(":memory:")

	fi := FsFileInfo{
		Path: "/foo.txt",
	}

	err := db.SetFileInfo(fi)

	assert.Nil(t, err)

	gotten, err := db.GetFileInfo("/foo.txt")
	assert.Nil(t, err)
	assert.Equal(t, gotten, fi)

	_, err = db.GetFileInfo("/noper.txt")
	assert.NotNil(t, err)

	err = db.UpdatePath("/foo.txt", "/something/else.txt")
	assert.Nil(t, err)

	_, err = db.GetFileInfo("/something/else.txt")
	assert.Nil(t, err)

	_, err = db.GetFileInfo("/foo.txt")
	assert.NotNil(t, err)

	assert.Nil(t, db.DeleteFileInfo("/something/else.txt"))
	_, err = db.GetFileInfo("/something/else.txt")
	assert.NotNil(t, err)
}

func TestGetPage(t *testing.T) {
	db := New(":memory:")

	assert.Nil(t, db.SetFileInfo(getFileInfo("/Bro.txt")))
	assert.Nil(t, db.SetFileInfo(getFileInfo("/Ankle.txt")))
	assert.Nil(t, db.SetFileInfo(getFileInfo("/Clarinet.txt")))

	infos, err := db.GetPage(0, 0)
	assert.Nil(t, err)
	assert.Len(t, infos, 3)

	infos, err = db.GetPage(1, 1)
	assert.Nil(t, err)
	assert.Len(t, infos, 1)

	infos, err = db.GetPage(100, 1)
	assert.Nil(t, err)
	assert.Len(t, infos, 0)
}

func TestMarkDeleted(t *testing.T) {
	db := New(":memory:")

	assert.Nil(t, db.SetFileInfo(getFileInfo("/Bro.txt")))

	fi := FsFileInfo{
		Path:          "/dude.txt",
		TrackingValue: 12345,
	}
	assert.Nil(t, db.SetFileInfo(fi))

	err := db.MarkDeleted(12345)
	assert.Nil(t, err)

	fi, err = db.GetFileInfo("/dude.txt")
	assert.Nil(t, err)
	assert.False(t, fi.Deleted)

	fi, err = db.GetFileInfo("/Bro.txt")
	assert.Nil(t, err)
	assert.True(t, fi.Deleted)
}

func getFileInfo(path string) FsFileInfo {
	return FsFileInfo{
		Path: path,
	}
}
