package fileprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFolderIndex_SortFiles(t *testing.T) {
	i := NewIndex()
	i.Files = []*FileInfo{
		{Path: "/Bear.txt"},
		{Path: "/Apple.txt"},
		{Path: "/banana.txt"},
		{Path: "/anus.txt"},
	}

	i.SortFiles()

	assert.Equal(t, i.Files[0].Path, "/anus.txt")
	assert.Equal(t, i.Files[1].Path, "/Apple.txt")
	assert.Equal(t, i.Files[2].Path, "/banana.txt")
	assert.Equal(t, i.Files[3].Path, "/Bear.txt")
}
