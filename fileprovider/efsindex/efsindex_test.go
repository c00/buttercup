package efsindex

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"

	"github.com/c00/buttercup/logger"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func createDb() (string, *EfsIndex) {
	nr, _ := rand.Int(rand.Reader, big.NewInt(100000))

	godotenv.Load("../../.env")
	dbPath := path.Join(os.Getenv("TEST_CONF_PATH"), fmt.Sprintf("buttercup-test-%v.db", nr))
	logger.Debug("db path: %v", dbPath)
	os.Remove(dbPath)
	db := New(dbPath, "foo")
	err := db.Load()
	if err != nil {
		panic("db cannot be openeed: " + err.Error())
	}

	return dbPath, db
}

func cleanupDb(path string, db *EfsIndex) {
	db.Close()
	os.Remove(path)
}

func TestBasics(t *testing.T) {
	dbPath, db := createDb()
	defer cleanupDb(dbPath, db)

	fi := EfsFileInfo{
		Path:       "/foo.txt",
		StoredPath: "some/encrypted/path",
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

func TestExisting(t *testing.T) {
	dbPath, db := createDb()
	defer cleanupDb(dbPath, db)

	fi := EfsFileInfo{
		Path:       "/foo.txt",
		StoredPath: "some/encrypted/path",
	}

	err := db.SetFileInfo(fi)
	assert.Nil(t, err)

	assert.Nil(t, db.Close())

	newDb := New(dbPath, "foo")
	defer newDb.Close()

	gotten, err := newDb.GetFileInfo("/foo.txt")
	assert.Nil(t, err)
	assert.Equal(t, gotten, fi)
}

func TestGetPage(t *testing.T) {
	dbPath, db := createDb()
	defer cleanupDb(dbPath, db)

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
	dbPath, db := createDb()
	defer cleanupDb(dbPath, db)

	assert.Nil(t, db.SetFileInfo(getFileInfo("/Bro.txt")))

	fi := EfsFileInfo{
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

func getFileInfo(path string) EfsFileInfo {
	return EfsFileInfo{
		Path:       path,
		StoredPath: "some/encrypted/path",
	}
}
