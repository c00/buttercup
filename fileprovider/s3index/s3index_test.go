package s3index

import (
	"os"
	"testing"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/fileprovider/s3client"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func createDb() *S3Index {
	godotenv.Load("../../.env")

	s3client := s3client.New(appconfig.S3ProviderConfig{
		Passphrase:     "foobar",
		AccessKey:      os.Getenv("TEST_S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("TEST_S3_SECRET_KEY"),
		Endpoint:       os.Getenv("TEST_S3_ENDPOINT"),
		Region:         os.Getenv("TEST_S3_REGION"),
		Bucket:         os.Getenv("TEST_S3_BUCKET"),
		BasePath:       "automated",
		ForcePathStyle: true,
	})

	index := New(s3client, "foo")
	s3client.DeleteFolder("")

	return index
}

func cleanupDb(db *S3Index) {
	db.Close()
}

func TestBasics(t *testing.T) {
	db := createDb()
	defer cleanupDb(db)

	fi := S3FileInfo{
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
	db := createDb()
	defer cleanupDb(db)

	fi := S3FileInfo{
		Path:       "/foo.txt",
		StoredPath: "some/encrypted/path",
	}

	err := db.SetFileInfo(fi)
	assert.Nil(t, err)

	assert.Nil(t, db.Close())

	newDb := New(db.s3client, "foo")
	defer newDb.Close()

	gotten, err := newDb.GetFileInfo("/foo.txt")
	assert.Nil(t, err)
	assert.Equal(t, gotten, fi)
}

func TestGetPage(t *testing.T) {
	db := createDb()
	defer cleanupDb(db)

	//Cleanup
	db.DeleteFileInfo("/foo.txt")
	db.DeleteFileInfo("/dude.txt")

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
	db := createDb()
	defer cleanupDb(db)

	assert.Nil(t, db.SetFileInfo(getFileInfo("/Bro.txt")))

	fi := S3FileInfo{
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

func getFileInfo(path string) S3FileInfo {
	return S3FileInfo{
		Path:       path,
		StoredPath: "some/encrypted/path",
	}
}
