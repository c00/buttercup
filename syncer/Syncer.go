package syncer

import (
	"fmt"

	"strings"

	"github.com/c00/buttercup/fileprovider"
	"github.com/c00/buttercup/logger"
	"github.com/c00/buttercup/source"
)

func New(local fileprovider.FileProvider, remote fileprovider.FileProvider) *Syncer {
	return &Syncer{
		source: source.NewSource(local, remote),
		local:  local,
		remote: remote,
	}
}

type Syncer struct {
	source source.PushPuller
	remote fileprovider.FileProvider
	local  fileprovider.FileProvider
}

func (s *Syncer) Pull() error {
	err := s.local.Lock()
	if err != nil {
		return fmt.Errorf("cannot lock local: %w", err)
	}
	defer s.local.Unlock()

	//Get remote index
	//todo introduce paging
	remoteFiles, err := s.remote.GetFileInfos(0, 0)
	if err != nil {
		return fmt.Errorf("could not get remote files: %w", err)
	}

	//Compare files, Pull new / updated ones.
	for _, rfile := range remoteFiles {
		lfile, err := s.local.GetFileInfo(rfile.Path)
		if err != nil {
			logger.Log("pulling new file: %v", rfile.Path)
			err := s.source.PullFile(rfile, rfile.Path)
			if err != nil {
				logger.Error("Error pulling file: %v", err)
			}
			continue
		}
		cmpResult, err := rfile.Compare(lfile, false)
		if err != nil {
			logger.Error("Skipping %v: %v", lfile.Path, err)
			continue
		}

		//Run Action
		switch cmpResult {
		case fileprovider.UpToDate:
			//don't log deleted files. It's confusing
			if !rfile.Deleted {
				logger.Info("%v: up-to-date already", rfile.Path)
			} else {
				logger.Debug("%v: up-to-date and deleted", rfile.Path)
			}
		case fileprovider.RemoteNewer:
			logger.Log("%v: pulling new version", rfile.Path)
			err := s.source.PullFile(rfile, lfile.Path)
			if err != nil {
				logger.Error("Error pulling file: %v", err)
			}
		case fileprovider.ConflictLocalNewer:
			logger.Log("%v: both files changed, local is more recent.\n", rfile.Path)
			//Keep local, but place the remote file with different name
			newPath := getConflictName(rfile.Path)
			err := s.source.PullFile(rfile, newPath)
			if err != nil {
				logger.Error("Error pulling file: %v", err)
			}
		case fileprovider.ConflictRemoteNewer:
			logger.Log("%v: both files changed, remote is more recent.\n", rfile.Path)
			//Rename local, and download remote file
			newPath := getConflictName(rfile.Path)
			err := s.local.MoveFile(rfile.Path, newPath)
			if err != nil {
				logger.Error("renaming file failed: %v", err)
				continue
			}

			//Download the new file
			err = s.source.PullFile(rfile, rfile.Path)
			if err != nil {
				logger.Error("Error fetching file: %v", err)
			}
		}
	}

	return nil
}

func (s *Syncer) canPush() (bool, error) {
	//We can pull if no file is newer on the remote than on the local.
	//Get remote index
	//todo introduce paging
	remoteFiles, err := s.remote.GetFileInfos(0, 0)
	if err != nil {
		return false, fmt.Errorf("could not get remote files: %w", err)
	}

	//Compare files, Pull new / updated ones.
	for _, rfile := range remoteFiles {
		lfile, err := s.local.GetFileInfo(rfile.Path)
		if err != nil {
			continue
		}
		cmpResult, err := rfile.Compare(lfile, false)
		if err != nil {
			return false, err
		}

		if cmpResult == fileprovider.LocalNewer || cmpResult == fileprovider.UpToDate {
			continue
		} else {
			return false, nil
		}
	}

	return true, nil
}

func (s *Syncer) Push() error {
	err := s.remote.Lock()
	if err != nil {
		return fmt.Errorf("cannot lock remote: %w", err)
	}
	defer s.remote.Unlock()

	//Also lock local because we want to update the lasSynced dates.
	err = s.local.Lock()
	if err != nil {
		return fmt.Errorf("cannot lock local: %w", err)
	}
	defer s.local.Unlock()

	canPull, err := s.canPush()
	if err != nil {
		return fmt.Errorf("cannot check if we can pull: %w", err)
	}
	if !canPull {
		return fmt.Errorf("cannot push, local is missing updates from remote. Pull first")
	}

	//Get localindex
	//todo introduce paging
	localFiles, err := s.local.GetFileInfos(0, 0)
	if err != nil {
		return fmt.Errorf("could not get local files: %w", err)
	}

	//Compare files, Pull new / updated ones.
	for _, localFi := range localFiles {
		remoteFi, err := s.remote.GetFileInfo(localFi.Path)
		if err != nil {
			logger.Log("pushing new file: %v", localFi.Path)
			err := s.source.PushFile(localFi)
			if err != nil {
				logger.Error("Error pushing file: %v", err)
			}
			continue
		}
		cmpResult, err := localFi.Compare(remoteFi, true)
		if err != nil {
			logger.Error("Skipping %v: %v", remoteFi.Path, err)
			continue
		}

		//Run Action
		switch cmpResult {
		case fileprovider.UpToDate:
			logger.Info("%v: up-to-date already", localFi.Path)
		case fileprovider.LocalNewer:
			logger.Log("%v: pushing updated file", localFi.Path)
			//Download the new file
			err = s.source.PushFile(localFi)
			if err != nil {
				logger.Error("Error pushing file: %v", err)
			}
		default:
			logger.Error("%v: unexpected compare result: %v\n", localFi.Path, cmpResult)
		}
	}

	return nil
}

func (s *Syncer) Sync() error {
	err := s.Pull()
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return s.Push()
}

func getConflictName(path string) string {
	if path == "" {
		return path
	}

	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		return path + ".conflict"
	}

	parts = append(parts[:len(parts)-1], "conflict", parts[len(parts)-1])
	return strings.Join(parts, ".")
}
