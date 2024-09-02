package s3client

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/c00/buttercup/appconfig"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestS3Client_UploadFile(t *testing.T) {
	godotenv.Load("../../.env")

	client := New(appconfig.S3ProviderConfig{
		Passphrase:     "foobar",
		AccessKey:      os.Getenv("TEST_S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("TEST_S3_SECRET_KEY"),
		Endpoint:       os.Getenv("TEST_S3_ENDPOINT"),
		Region:         os.Getenv("TEST_S3_REGION"),
		Bucket:         os.Getenv("TEST_S3_BUCKET"),
		BasePath:       "automated",
		ForcePathStyle: true,
	})

	err := client.UploadFile("/foo.txt", strings.NewReader("some content"))
	assert.Nil(t, err)

	reader, err := client.DownloadFile("/foo.txt")
	assert.Nil(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	assert.Nil(t, err)
	assert.Equal(t, string(data), "some content")

	err = client.DeleteFile("/foo.txt")
	assert.Nil(t, err)

	_, err = client.DownloadFile("/foo.txt")
	assert.NotNil(t, err)
}

func TestS3Client_DeleteFolder(t *testing.T) {
	godotenv.Load("../../.env")

	client := New(appconfig.S3ProviderConfig{
		Passphrase:     "foobar",
		AccessKey:      os.Getenv("TEST_S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("TEST_S3_SECRET_KEY"),
		Endpoint:       os.Getenv("TEST_S3_ENDPOINT"),
		Region:         os.Getenv("TEST_S3_REGION"),
		Bucket:         os.Getenv("TEST_S3_BUCKET"),
		BasePath:       "automated",
		ForcePathStyle: true,
	})

	assert.Nil(t, client.UploadFile("/delfolder/foo.txt", strings.NewReader("some content")))
	assert.Nil(t, client.UploadFile("/delfolder/bar.txt", strings.NewReader("some more content")))

	err := client.DeleteFolder("/delfolder")
	assert.Nil(t, err)

	exists, err := client.HasFile("/delfolder/foo.txt")
	assert.Nil(t, err)
	assert.False(t, exists)

	exists, err = client.HasFile("/delfolder/bar.txt")
	assert.Nil(t, err)
	assert.False(t, exists)

}

func TestS3Client_HasFile(t *testing.T) {
	godotenv.Load("../../.env")

	client := New(appconfig.S3ProviderConfig{
		Passphrase:     "foobar",
		AccessKey:      os.Getenv("TEST_S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("TEST_S3_SECRET_KEY"),
		Endpoint:       os.Getenv("TEST_S3_ENDPOINT"),
		Region:         os.Getenv("TEST_S3_REGION"),
		Bucket:         os.Getenv("TEST_S3_BUCKET"),
		BasePath:       "automated",
		ForcePathStyle: true,
	})

	filepath := "/Foo.txt"
	assert.Nil(t, client.UploadFile(filepath, strings.NewReader("some content")))

	has, err := client.HasFile(filepath)
	assert.Nil(t, err)
	assert.True(t, has)

	has, err = client.HasFile("/wacky.foo")
	assert.Nil(t, err)
	assert.False(t, has)
}
