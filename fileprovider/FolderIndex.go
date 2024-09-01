package fileprovider

import (
	"slices"
	"strings"
)

func NewIndex() FolderIndex {
	return FolderIndex{
		Version: 1,
		Files:   []*FileInfo{},
	}
}

type FolderIndex struct {
	Version int
	Files   []*FileInfo
}

func (i *FolderIndex) SortFiles() {
	slices.SortStableFunc(i.Files, func(a *FileInfo, b *FileInfo) int {
		pathA := strings.ToLower(a.Path)
		pathB := strings.ToLower(b.Path)
		if pathA < pathB {
			return -1
		} else if pathA > pathB {
			return 1
		}
		return 0
	})
}
