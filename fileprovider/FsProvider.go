package fileprovider

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/fileprovider/efsindex"
	"github.com/c00/buttercup/fileprovider/fsindex"
)

func NewFsProvider(conf appconfig.ProviderConfig) *FsProvider {
	if conf.FsConfig == nil {
		panic("fs config is not defined")
	}
	provider := &FsProvider{
		Path:  conf.FsConfig.Path,
		index: fsindex.New(path.Join(conf.FsConfig.Path, sqliteIndexName)),
		name:  conf.ClientName,
	}

	os.Mkdir(provider.Path, 0700)

	err := provider.index.Load()
	if err != nil {
		panic(fmt.Errorf("cannot load database: %w", err))
	}

	err = provider.refreshDates()
	if err != nil {
		panic(fmt.Errorf("cannot scan filesystem: %w", err))
	}

	return provider
}

//todo write and read from StoredPath instead of Path
//Or remote storedPath in this thing if we choose not to support it.

type FsProvider struct {
	Path string
	name string
	// useEncryption bool
	// passphrase    string
	index *fsindex.FsIndex
}

func (p *FsProvider) SetLastSynced(filePath string, date time.Time) error {
	fi, err := p.getFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	fi.LastSynced = date
	p.index.SetFileInfo(fi)
	return nil
}

func (r *FsProvider) MoveFile(oldPath, newPath string) error {
	fullNewPath := path.Join(r.Path, newPath)
	err := os.MkdirAll(path.Dir(fullNewPath), 0755)
	if err != nil {
		return fmt.Errorf("could not create dirs: %w", err)
	}

	err = os.Rename(path.Join(r.Path, oldPath), fullNewPath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	err = r.index.UpdatePath(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("could not update path in db: %w", err)
	}

	return nil
}

func (p *FsProvider) RetrieveFile(filePath string) (io.ReadCloser, error) {
	fullPath := path.Join(p.Path, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file for retrieval: %w", err)
	}

	return file, nil
}

func (p *FsProvider) StoreFile(otherFi FileInfo, stream io.Reader) error {
	//Get index
	fi, err := p.getFileInfo(otherFi.Path)
	if err != nil {
		fi = fsindex.FsFileInfo{
			Path:    otherFi.Path,
			Updated: otherFi.Updated,
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

func (p *FsProvider) store(fi fsindex.FsFileInfo, stream io.Reader) error {
	fullPath := path.Join(p.Path, fi.Path)
	os.MkdirAll(path.Dir(fullPath), 0755)

	writer, err := os.OpenFile(fullPath, os.O_CREATE+os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("could not open or create local path: %w", err)
	}

	_, err = io.Copy(writer, stream)
	if err != nil {
		writer.Close()
		return fmt.Errorf("could not copy file: %w", err)
	}

	writer.Close()
	err = os.Chtimes(fullPath, fi.Updated, fi.Updated)
	if err != nil {
		return fmt.Errorf("could not set new updated times: %w", err)
	}

	return nil
}

func (p *FsProvider) RemoveFile(otherFi FileInfo) error {
	fi, err := p.getFileInfo(otherFi.Path)
	if err != nil {
		p.index.SetFileInfo(fsindex.FsFileInfo{
			Path:    otherFi.Path,
			Updated: otherFi.Updated,
			Deleted: true,
		})
		return nil
	}

	fullPath := path.Join(p.Path, otherFi.Path)

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

func (p *FsProvider) Lock() error {
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

func (p *FsProvider) Unlock() error {
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

// todo rename
func (p *FsProvider) getFileInfo(filePath string) (fsindex.FsFileInfo, error) {
	fi, err := p.index.GetFileInfo(filePath)
	if err != nil {
		return fsindex.FsFileInfo{}, errors.New("file not found in index")
	}
	return fi, nil
}

func (p *FsProvider) GetFileInfo(path string) (FileInfo, error) {
	fi, err := p.getFileInfo(path)
	if err != nil {
		return FileInfo{}, err
	}
	return fsFileInfoToFileInfo(fi), nil
}

func (p *FsProvider) GetFileInfos(limit, offset int) ([]FileInfo, error) {
	files, err := p.index.GetPage(offset, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get page: %w", err)
	}

	result := make([]FileInfo, 0, len(files))
	for _, fi := range files {
		result = append(result, fsFileInfoToFileInfo(fi))
	}

	return result, nil
}

// Update files with Updated date
func (p *FsProvider) refreshDates() error {
	trackingValue := time.Now().Unix()

	//Walk the folder
	//Remove final slash
	p.Path = strings.TrimSuffix(p.Path, "/")

	//Read files from IO.
	err := filepath.Walk(p.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle error if there is any
		}

		// Check if the file is not a directory
		if info.IsDir() {
			return nil
		}

		relativePath := path[len(p.Path):]

		if relativePath == "/"+sqliteIndexName || relativePath == "/"+lockfileName {
			return nil
		}

		//Check if file exists in index
		fi, err := p.index.GetFileInfo(relativePath)
		if err != nil {
			//Newly created file
			err = p.index.SetFileInfo(fsindex.FsFileInfo{
				Path:          relativePath,
				Updated:       info.ModTime(),
				Deleted:       false,
				TrackingValue: trackingValue,
			})
			if err != nil {
				return fmt.Errorf("could not store file info: %w", err)
			}
		} else {
			//Set magic value
			fi.TrackingValue = trackingValue
			fi.Updated = info.ModTime()
			err = p.index.SetFileInfo(fi)
			if err != nil {
				return fmt.Errorf("could not update file info: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the path: %w", err)
	}

	err = p.markDeleted(trackingValue)
	if err != nil {
		return fmt.Errorf("error marking deleted: %w", err)
	}

	return nil
}

// Mark files with no Updated value as deleted todo
func (p *FsProvider) markDeleted(trackingValue int64) error {

	//Everything that does NOT have this tracking value, should be marked as deleted.
	return p.index.MarkDeleted(trackingValue)
}

func fsFileInfoToFileInfo(fi fsindex.FsFileInfo) FileInfo {
	return FileInfo{
		Path:       fi.Path,
		LastSynced: fi.LastSynced,
		Updated:    fi.Updated,
		Deleted:    fi.Deleted,
	}
}

func efsFileInfoToFileInfo(fi efsindex.EfsFileInfo) FileInfo {
	return FileInfo{
		Path:       fi.Path,
		LastSynced: fi.LastSynced,
		Updated:    fi.Updated,
		Deleted:    fi.Deleted,
	}
}
