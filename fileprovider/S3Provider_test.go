package fileprovider

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/fileprovider/s3client"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestS3Provider_RunProviderSuite(t *testing.T) {
	godotenv.Load("../.env")

	RunSuite(t, func() FileProvider {

		s3conf := &appconfig.S3ProviderConfig{
			Passphrase:     "foo",
			AccessKey:      os.Getenv("TEST_S3_ACCESS_KEY"),
			SecretKey:      os.Getenv("TEST_S3_SECRET_KEY"),
			Endpoint:       os.Getenv("TEST_S3_ENDPOINT"),
			Region:         os.Getenv("TEST_S3_REGION"),
			Bucket:         os.Getenv("TEST_S3_BUCKET"),
			BasePath:       "TestS3Provider_RunProviderSuite",
			ForcePathStyle: true,
		}
		client := s3client.New(*s3conf)
		client.DeleteFolder("")

		return NewS3Provider(appconfig.ProviderConfig{
			Type:     TypeFs,
			S3Config: s3conf,
		})
	})
}

func TestFigureOutHowToReuseTempFiles(t *testing.T) {
	file, err := os.CreateTemp("", "something")
	assert.Nil(t, err)
	written, err := io.Copy(file, strings.NewReader("I live in a giant bucket"))
	assert.Nil(t, err)
	assert.True(t, written > 0)

	_, err = file.Seek(0, 0)
	assert.Nil(t, err)

	data, err := io.ReadAll(file)
	assert.Nil(t, err)
	assert.Equal(t, string(data), "I live in a giant bucket")
}
