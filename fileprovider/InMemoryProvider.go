package fileprovider

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/c00/buttercup/simplekeyvaluestore"
)

type storedFile struct {
	data    []byte
	updated time.Time
}

func NewInMemoryProvider(clientName string) *InMemoryProvider {
	return &InMemoryProvider{
		store: simplekeyvaluestore.NewSimpleKeyValueStore[storedFile](),
		index: NewIndex(),
		name:  clientName,
	}
}

type InMemoryProvider struct {
	//Name used in Locks
	name  string
	store simplekeyvaluestore.SimpleKeyValueStore[storedFile]
	index FolderIndex
}

func (r *InMemoryProvider) MoveFile(oldPath, newPath string) error {
	f, err := r.store.Get(oldPath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	r.store.Set(newPath, f)
	r.store.Delete(oldPath)

	//Update index
	for _, file := range r.index.Files {
		if file.Path == oldPath {
			file.Path = newPath
			break
		}
	}

	return nil
}

func (p *InMemoryProvider) SetLastSynced(filePath string, date time.Time) error {
	fi, err := p.getFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	fi.LastSynced = date
	return nil
}

func (p *InMemoryProvider) RetrieveFile(filePath string) (io.ReadCloser, error) {
	file, err := p.store.Get(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file for retrieval: %w", err)
	}

	return &memReadCloser{reader: bytes.NewReader(file.data)}, nil
}

func (p *InMemoryProvider) StoreFile(otherFi FileInfo, stream io.Reader) error {
	if !p.canWrite() {
		return fmt.Errorf("resource locked by other client")
	}

	fi, err := p.getFileInfo(otherFi.Path)
	if err != nil {
		//File not found, create the fileInfo instead
		fi = &FileInfo{
			Path:    otherFi.Path,
			Updated: otherFi.Updated,
		}
		p.index.Files = append(p.index.Files, fi)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("could not read stream: %w", err)
	}
	p.store.Set(otherFi.Path, storedFile{data: data, updated: otherFi.Updated})

	fi.Updated = otherFi.Updated
	fi.Deleted = false

	return nil
}

func (p *InMemoryProvider) RemoveFile(otherFi FileInfo) error {
	if !p.canWrite() {
		return fmt.Errorf("resource locked by other client")
	}

	fi, err := p.getFileInfo(otherFi.Path)
	//Create it if it doesn't exist
	if err != nil {
		otherFi.Deleted = true
		p.index.Files = append(p.index.Files, &otherFi)
		return nil
	}

	p.store.Delete(otherFi.Path)

	fi.Updated = otherFi.Updated
	fi.Deleted = true

	return nil
}

func (p *InMemoryProvider) canWrite() bool {
	file, err := p.store.Get(lockfileName)
	if err != nil {
		return true
	}

	if string(file.data) != p.name {
		return false
	}

	return true
}

func (p *InMemoryProvider) Lock() error {
	if p.store.Has(lockfileName) {
		return errors.New("cannot set lock, already locked")
	}

	data := []byte(p.name)
	p.store.Set(lockfileName, storedFile{data: data, updated: time.Now()})
	return nil
}

func (p *InMemoryProvider) Unlock() error {
	file, err := p.store.Get(lockfileName)
	if err != nil {
		return fmt.Errorf("already unlocked")
	}

	if string(file.data) != p.name {
		return fmt.Errorf("store locked by another client: %v", string(file.data))
	}

	p.store.Delete(lockfileName)
	return nil
}

func (p *InMemoryProvider) getFileInfo(path string) (*FileInfo, error) {
	//todo make this cache?
	for _, v := range p.index.Files {
		if v.Path == path {
			return v, nil
		}
	}

	return nil, errors.New("not found")
}

func (p *InMemoryProvider) GetFileInfo(path string) (FileInfo, error) {
	fi, err := p.getFileInfo(path)
	if err != nil {
		return FileInfo{}, err
	}
	return *fi, nil
}

func (p *InMemoryProvider) GetFileInfos(limit, offset int) ([]FileInfo, error) {
	if limit == 0 {
		limit = len(p.index.Files)
	}

	endIndex := limit + offset
	if endIndex > len(p.index.Files) {
		endIndex = len(p.index.Files)
	}

	if offset > len(p.index.Files) {
		return []FileInfo{}, nil
	}

	slice := p.index.Files[offset:endIndex]
	result := make([]FileInfo, 0, len(slice))
	for _, file := range slice {
		result = append(result, *file)
	}

	return result, nil
}

type memReadCloser struct {
	reader *bytes.Reader
}

func (rc *memReadCloser) Read(b []byte) (int, error) {
	return rc.reader.Read(b)
}

func (rc *memReadCloser) Close() error {
	return nil
}
