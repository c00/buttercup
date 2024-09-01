package fileprovider

import (
	"fmt"
	"time"
)

const UpToDate = 1
const RemoteNewer = 2
const LocalNewer = 3
const ConflictRemoteNewer = 4
const ConflictLocalNewer = 5

type FileInfo struct {
	// Path to the file
	Path string
	// Date the file was last synced with the remote
	// This date will not be used on remotes (as they could be synced with multiple clients)
	LastSynced time.Time
	// File System updated date
	Updated time.Time
	// Whether the file is deleted or not
	Deleted bool
}

// Compare a remote file with a local file
// The goal of this function is to decide whether or not to download the file.
// asLocal indicates whether this called from a fileInfo that comes from a local vs a remote.
func (fi FileInfo) Compare(other FileInfo, asLocal bool) (int, error) {
	var changedLocally, changedRemotely bool
	var local, remote FileInfo

	if asLocal {
		changedLocally = fi.LastSynced.Before(fi.Updated)
		changedRemotely = fi.LastSynced.Before(other.Updated)

		local = fi
		remote = other
	} else {
		changedLocally = other.LastSynced.Before(other.Updated)
		changedRemotely = other.LastSynced.Before(fi.Updated)

		local = other
		remote = fi
	}

	// up-to-date - No action needed. Files are the same.
	if !changedLocally && !changedRemotely {
		return UpToDate, nil
	}

	//Remote out of date.
	if changedLocally && !changedRemotely {
		return LocalNewer, nil
	}

	// remote-newer - Remote should replace local
	if !changedLocally && changedRemotely {
		return RemoteNewer, nil
	}

	//Only conflict states are left now.

	// conflict-remote-newer - Both have changed, remote is more recent.
	if remote.Updated.Before(local.Updated) {
		return ConflictLocalNewer, nil
	}

	// conflict-local-newer - Both have changes, local is more recent.
	if remote.Updated.After(local.Updated) {
		return ConflictRemoteNewer, nil
	}

	//Technically unreachable.
	return 0, fmt.Errorf("unknown state. Path: %s, local.LastSynced: %s local.Updated: %s, remote.Updated: %s", fi.Path, local.LastSynced.Format(time.RFC822Z), local.Updated.Format(time.RFC822Z), other.Updated.Format(time.RFC822Z))
}
