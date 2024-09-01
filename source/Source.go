package source

import (
	"fmt"

	"github.com/c00/buttercup/fileprovider"
)

func NewSource(local fileprovider.FileProvider, remote fileprovider.FileProvider) *Source {
	return &Source{
		remote: remote,
		local:  local,
	}
}

type Source struct {
	remote fileprovider.FileProvider
	local  fileprovider.FileProvider
}

// Pull a file from the remote into the local
func (s *Source) PullFile(fi fileprovider.FileInfo, newPath string) error {
	if fi.Deleted {
		err := s.local.RemoveFile(fi)
		if err != nil {
			return fmt.Errorf("could not remove file locally: %w", err)
		}
	} else {

		reader, err := s.remote.RetrieveFile(fi.Path)
		if err != nil {
			return fmt.Errorf("could not retrieve file from remote: %w", err)
		}
		defer reader.Close()

		fi.Path = newPath
		err = s.local.StoreFile(fi, reader)
		if err != nil {
			return fmt.Errorf("could not store file locally: %w", err)
		}
	}

	err := s.local.SetLastSynced(fi.Path, fi.Updated)
	if err != nil {
		return fmt.Errorf("could not set lastSynced date: %w", err)
	}

	return nil
}

func (s *Source) PushFile(fi fileprovider.FileInfo) error {
	if fi.Deleted {
		err := s.remote.RemoveFile(fi)
		if err != nil {
			return fmt.Errorf("could not remove file remotely: %w", err)
		}
	} else {
		reader, err := s.local.RetrieveFile(fi.Path)
		if err != nil {
			return fmt.Errorf("could not read local file: %w", err)
		}
		defer reader.Close()

		err = s.remote.StoreFile(fi, reader)
		if err != nil {
			return fmt.Errorf("could not store file remotely: %w", err)
		}
	}

	err := s.local.SetLastSynced(fi.Path, fi.Updated)
	if err != nil {
		return fmt.Errorf("could not set lastSynced date: %w", err)
	}

	return nil
}
