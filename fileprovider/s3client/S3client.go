package s3client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/c00/buttercup/appconfig"
)

// todo do the splitting up into chunkeronis
func New(conf appconfig.S3ProviderConfig) *S3Client {
	return &S3Client{
		config: conf,
	}
}

type S3Client struct {
	config appconfig.S3ProviderConfig
	client *s3.Client
}

func (c *S3Client) getClient() (*s3.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithRegion(c.config.Region),
	)

	if err != nil {
		return nil, err
	}

	c.client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = &c.config.Endpoint
		o.Credentials = credentials.NewStaticCredentialsProvider(c.config.AccessKey, c.config.SecretKey, "")
	})

	return c.client, nil
}

func (c *S3Client) UploadFile(filepath string, content io.Reader) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	key := path.Join(c.config.BasePath, filepath)

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Body:   content,
		Bucket: &c.config.Bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("upload to s3 failed: %w", err)
	}

	return nil
}

func (c *S3Client) DownloadFile(filepath string) (io.ReadCloser, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	key := path.Join(c.config.BasePath, filepath)

	result, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &c.config.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

func (c *S3Client) DeleteFile(filepath string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	key := path.Join(c.config.BasePath, filepath)

	_, err = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &c.config.Bucket,
		Key:    &key,
	})

	//Note, no errors are thrown if the key already doesn't exist.
	if err != nil {
		return err
	}

	return nil
}

func (c *S3Client) DeleteFolder(prefix string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	list, err := client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(c.config.Bucket),
		Prefix: aws.String(path.Join(c.config.BasePath, prefix)),
	})
	if err != nil {
		return fmt.Errorf("could not list items in folder: %w", err)
	}

	objects := []types.ObjectIdentifier{}

	for _, c := range list.Contents {
		objects = append(objects, types.ObjectIdentifier{Key: c.Key})
	}

	_, err = client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
		Bucket: aws.String(c.config.Bucket),
		Delete: &types.Delete{Objects: objects},
	})

	if err != nil {
		return fmt.Errorf("could not delete folder: %w", err)
	}

	return nil
}

func (c *S3Client) HasFile(filepath string) (bool, error) {
	client, err := c.getClient()
	if err != nil {
		return false, err
	}

	_, err = client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(path.Join(c.config.BasePath, filepath)),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, fmt.Errorf("could not head item: %w", err)
	}
	return true, nil
}
