package fileprovider

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/fileprovider/efsindex"
	"github.com/c00/buttercup/modifiers"
)

// Create a new Encrypted File System Provider
func NewEfsProvider(conf appconfig.ProviderConfig) *EfsProvider {
	if conf.EfsConfig == nil {
		panic("efs config is not defined")
	}
	provider := &EfsProvider{
		Path:       conf.EfsConfig.Path,
		index:      efsindex.New(path.Join(conf.EfsConfig.Path, sqliteIndexName), conf.EfsConfig.Passphrase),
		name:       conf.ClientName,
		passphrase: conf.EfsConfig.Passphrase,
	}

	os.Mkdir(provider.Path, 0700)

	err := provider.index.Load()
	if err != nil {
		panic(fmt.Errorf("cannot load database: %w", err))
	}

	return provider
}

type EfsProvider struct {
	Path       string
	name       string
	passphrase string
	index      *efsindex.EfsIndex
}

func (p *EfsProvider) SetLastSynced(filePath string, date time.Time) error {
	fi, err := p.index.GetFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	fi.LastSynced = date
	p.index.SetFileInfo(fi)
	return nil
}

func (r *EfsProvider) MoveFile(oldPath, newPath string) error {
	err := r.index.UpdatePath(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("could not update path in db: %w", err)
	}

	return nil
}

func (p *EfsProvider) RetrieveFile(filePath string) (io.ReadCloser, error) {
	fi, err := p.index.GetFileInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	fullPath := path.Join(p.Path, fi.StoredPath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file for retrieval: %w", err)
	}

	reader, err := modifiers.DecryptAndDecompress(file, p.passphrase)
	if err != nil {
		return nil, fmt.Errorf("could not compress and encrypt: %w", err)
	}
	return reader, nil
}

func (p *EfsProvider) StoreFile(otherFi FileInfo, stream io.Reader) error {
	fi, err := p.index.GetFileInfo(otherFi.Path)
	if err != nil {
		storedPath, err := CreateRandomPath()
		if err != nil {
			return fmt.Errorf("cannot create store path: %w", err)
		}
		fi = efsindex.EfsFileInfo{
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

func (p *EfsProvider) store(fi efsindex.EfsFileInfo, stream io.Reader) error {
	fullPath := path.Join(p.Path, fi.StoredPath)
	os.MkdirAll(path.Dir(fullPath), 0755)

	writer, err := os.OpenFile(fullPath, os.O_CREATE+os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("could not open or create local path: %w", err)
	}
	defer writer.Close()

	err = modifiers.CompressAndEncrypt(stream, writer, p.passphrase)
	if err != nil {
		return fmt.Errorf("could not compress and encrypt: %w", err)
	}

	err = os.Chtimes(fullPath, fi.Updated, fi.Updated)
	if err != nil {
		return fmt.Errorf("could not set new updated times: %w", err)
	}

	//todo update permissions

	return nil
}

func (p *EfsProvider) RemoveFile(otherFi FileInfo) error {
	fi, err := p.index.GetFileInfo(otherFi.Path)
	if err != nil {
		storedPath, err := CreateRandomPath()
		if err != nil {
			return fmt.Errorf("cannot generate random path: %w", err)
		}
		p.index.SetFileInfo(efsindex.EfsFileInfo{
			Path:       otherFi.Path,
			Updated:    otherFi.Updated,
			Deleted:    true,
			StoredPath: storedPath,
		})

		return nil
	}

	fullPath := path.Join(p.Path, fi.StoredPath)

	err = os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("could not remove file: %w", err)
	}

	fi.Deleted = true
	fi.Updated = otherFi.Updated
	err = p.index.SetFileInfo(fi)
	if err != nil {
		return fmt.Errorf("could not delete in index: %w", err)
	}

	return nil
}

func (p *EfsProvider) Lock() error {
	fullPath := path.Join(p.Path, lockfileName)
	_, err := os.Stat(fullPath)
	if err == nil {
		return errors.New("cannot set lock, already locked")
	}

	data := []byte(p.name)
	err = os.WriteFile(fullPath, data, 0600)
	if err != nil {
		return fmt.Errorf("error setting lock: %w", err)
	}

	return nil
}

func (p *EfsProvider) Unlock() error {
	fullPath := path.Join(p.Path, lockfileName)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("already unlocked")
	}

	if string(data) != p.name {
		return fmt.Errorf("store locked by another client: %v", string(data))
	}

	err = os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("error removing lock file: %w", err)
	}

	err = p.index.Close()
	if err != nil {
		return fmt.Errorf("error closing db: %w", err)
	}
	return nil
}

func (p *EfsProvider) GetFileInfo(path string) (FileInfo, error) {
	fi, err := p.index.GetFileInfo(path)
	if err != nil {
		return FileInfo{}, err
	}
	return efsFileInfoToFileInfo(fi), nil
}

func (p *EfsProvider) GetFileInfos(limit, offset int) ([]FileInfo, error) {
	files, err := p.index.GetPage(offset, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get page: %w", err)
	}

	result := make([]FileInfo, 0, len(files))
	for _, fi := range files {
		result = append(result, efsFileInfoToFileInfo(fi))
	}

	return result, nil
}
