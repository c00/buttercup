package fileprovider

import (
	"io"
	"time"
)

// todo create tests for this
type FileProvider interface {
	//Retrieve a file from the local folder in the form of a io.reader
	RetrieveFile(path string) (io.ReadCloser, error)
	//Store a file in the local folder.
	//Also updates the index
	//fails if a lock is set
	StoreFile(fi FileInfo, stream io.Reader) error
	//Remove a file from the provider
	//Also updates the index,
	//fails if a lock is set
	RemoveFile(fi FileInfo) error
	//Update lastSynced date
	SetLastSynced(path string, date time.Time) error

	//Moves / renames a file
	MoveFile(oldPath, newPath string) error

	//Get file info for a single file
	GetFileInfo(path string) (FileInfo, error)
	//Get a page of FileInfos
	GetFileInfos(limit, offset int) ([]FileInfo, error)

	//Set an exclusive lock on the provider so other clients cannot write to it.
	Lock() error
	//Release the exclusive lock. Also persists the index to disk if needed.
	Unlock() error
}
