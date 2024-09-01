package source

import (
	"github.com/c00/buttercup/fileprovider"
)

type PushPuller interface {
	PullFile(fi fileprovider.FileInfo, newPath string) error
	PushFile(fi fileprovider.FileInfo) error
}
