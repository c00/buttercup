package fileprovider

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/c00/buttercup/appconfig"
	s3client "github.com/c00/buttercup/fileprovider/s3client"
	"github.com/c00/buttercup/fileprovider/s3index"
	"github.com/c00/buttercup/modifiers"
)

// Create a new S3 Provider (Encrypted)
func NewS3Provider(conf appconfig.ProviderConfig) *S3Provider {
	if conf.S3Config == nil {
		panic("efs config is not defined")
	}

	s3 := s3client.New(*conf.S3Config)

	provider := &S3Provider{
		index:    s3index.New(s3, conf.S3Config.Passphrase),
		name:     conf.ClientName,
		config:   *conf.S3Config,
		s3client: s3,
	}

	err := provider.index.Load()
	if err != nil {
		panic(fmt.Errorf("cannot load database: %w", err))
	}

	return provider
}

type S3Provider struct {
	name     string
	config   appconfig.S3ProviderConfig
	index    *s3index.S3Index
	s3client *s3client.S3Client
}

func (p *S3Provider) SetLastSynced(filePath string, date time.Time) error {
	fi, err := p.index.GetFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	fi.LastSynced = date
	p.index.SetFileInfo(fi)
	return nil
}

func (r *S3Provider) MoveFile(oldPath, newPath string) error {
	err := r.index.UpdatePath(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("could not update path in db: %w", err)
	}

	return nil
}

func (p *S3Provider) RetrieveFile(filePath string) (io.ReadCloser, error) {
	fi, err := p.index.GetFileInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	file, err := p.s3client.DownloadFile(fi.StoredPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file for retrieval: %w", err)
	}

	reader, err := modifiers.DecryptAndDecompress(file, p.config.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("could not compress and encrypt: %w", err)
	}
	return reader, nil
}

func (p *S3Provider) StoreFile(otherFi FileInfo, stream io.Reader) error {
	fi, err := p.index.GetFileInfo(otherFi.Path)
	if err != nil {
		storedPath, err := CreateRandomPath()
		if err != nil {
			return fmt.Errorf("cannot create store path: %w", err)
		}
		fi = s3index.S3FileInfo{
			Path:       otherFi.Path,
			Updated:    otherFi.Updated,
			StoredPath: storedPath,
		}
		//File not found, create the fileInfo instead
		err = p.index.SetFileInfo(fi)
		if err != nil {
			return fmt.Errorf("could not store new file info: %w", err)
		}
	}

	err = p.store(fi, stream)
	if err != nil {
		return fmt.Errorf("could not store file: %w", err)
	}

	fi.Updated = otherFi.Updated
	fi.Deleted = false

	err = p.index.SetFileInfo(fi)
	if err != nil {
		return fmt.Errorf("could not update index db: %w", err)
	}

	return nil
}

func (p *S3Provider) store(fi s3index.S3FileInfo, stream io.Reader) error {
	tmpFile, err := os.CreateTemp("", "staged-")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	err = modifiers.CompressAndEncrypt(stream, tmpFile, p.config.Passphrase)
	if err != nil {
		return fmt.Errorf("could not compress and encrypt: %w", err)
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("could not reset head to 0: %w", err)
	}

	err = p.s3client.UploadFile(fi.StoredPath, tmpFile)
	if err != nil {
		return fmt.Errorf("could not upload to s3: %w", err)
	}

	return nil
}

func (p *S3Provider) RemoveFile(otherFi FileInfo) error {
	fi, err := p.index.GetFileInfo(otherFi.Path)
	if err != nil {
		storedPath, err := CreateRandomPath()
		if err != nil {
			return fmt.Errorf("cannot generate random path: %w", err)
		}
		p.index.SetFileInfo(s3index.S3FileInfo{
			Path:       otherFi.Path,
			Updated:    otherFi.Updated,
			Deleted:    true,
			StoredPath: storedPath,
		})

		return nil
	}

	err = p.s3client.DeleteFile(fi.StoredPath)
	if err != nil {
		return fmt.Errorf("could not delete file from s3: %w", err)
	}

	fi.Deleted = true
	fi.Updated = otherFi.Updated
	err = p.index.SetFileInfo(fi)
	if err != nil {
		return fmt.Errorf("could not delete in index: %w", err)
	}

	return nil
}

// todo am I checking for locks?
func (p *S3Provider) Lock() error {
	_, err := p.s3client.DownloadFile(lockfileName)
	if err == nil {
		return errors.New("cannot set lock, already locked")
	}

	err = p.s3client.UploadFile(lockfileName, strings.NewReader(p.name))
	if err != nil {
		return fmt.Errorf("error setting lock: %w", err)
	}

	return nil
}

func (p *S3Provider) Unlock() error {
	lockfileData, err := p.s3client.DownloadFile(lockfileName)
	if err != nil {
		return errors.New("cannot unlock, already unlocked")
	}

	data, err := io.ReadAll(lockfileData)
	if err != nil {
		return fmt.Errorf("cannot read lockfile stream: %w", err)
	}

	if string(data) != p.name {
		return fmt.Errorf("store locked by another client: %v", string(data))
	}

	err = p.s3client.DeleteFile(lockfileName)
	if err != nil {
		return fmt.Errorf("error removing lock file: %w", err)
	}

	err = p.index.Close()
	if err != nil {
		return fmt.Errorf("error closing db: %w", err)
	}
	return nil
}

func (p *S3Provider) GetFileInfo(path string) (FileInfo, error) {
	fi, err := p.index.GetFileInfo(path)
	if err != nil {
		return FileInfo{}, err
	}
	return s3FileInfoToFileInfo(fi), nil
}

func (p *S3Provider) GetFileInfos(limit, offset int) ([]FileInfo, error) {
	files, err := p.index.GetPage(offset, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get page: %w", err)
	}

	result := make([]FileInfo, 0, len(files))
	for _, fi := range files {
		result = append(result, s3FileInfoToFileInfo(fi))
	}

	return result, nil
}
