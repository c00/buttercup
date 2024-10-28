package daemoncmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/c00/buttercup/internal/fstests"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestBreadthFirstWalk(t *testing.T) {
	godotenv.Load("../../.env")
	sourcePath := os.Getenv("TEST_SOURCE_PATH")

	fstests.SetupSourceFilesystem(sourcePath, false)
	fstests.CreateTestFile(sourcePath, "bar/file1.txt")
	fstests.CreateTestFile(sourcePath, "file1.txt")
	fstests.CreateTestFile(sourcePath, "bar/baz/file1.txt")
	fstests.CreateTestFile(sourcePath, "bar/file2.txt")
	fstests.CreateTestFile(sourcePath, "file2.txt")
	fstests.CreateTestFile(sourcePath, "bar/baz/file2.txt")

	index := 0
	expected := []string{
		sourcePath,
		filepath.Join(sourcePath, "bar"),
		filepath.Join(sourcePath, "file1.txt"),
		filepath.Join(sourcePath, "file2.txt"),
		filepath.Join(sourcePath, "bar/baz"),
		filepath.Join(sourcePath, "bar/file1.txt"),
		filepath.Join(sourcePath, "bar/file2.txt"),
		filepath.Join(sourcePath, "bar/baz/file1.txt"),
		filepath.Join(sourcePath, "bar/baz/file2.txt"),
	}

	got := []string{}
	err := BreadthFirstWalk(sourcePath, func(path string, entry fs.DirEntry, err error) error {
		got = append(got, path)
		index++
		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestBreadthFirstDirWalk(t *testing.T) {
	godotenv.Load("../../.env")
	sourcePath := os.Getenv("TEST_SOURCE_PATH")

	fstests.SetupSourceFilesystem(sourcePath, false)
	fstests.CreateTestFile(sourcePath, "bar/file1.txt")
	fstests.CreateTestFile(sourcePath, "file1.txt")
	fstests.CreateTestFile(sourcePath, "bar/baz/file1.txt")
	fstests.CreateTestFile(sourcePath, "bar/file2.txt")
	fstests.CreateTestFile(sourcePath, "file2.txt")
	fstests.CreateTestFile(sourcePath, "bar/baz/file2.txt")

	index := 0
	expected := []string{
		sourcePath,
		filepath.Join(sourcePath, "bar"),
		filepath.Join(sourcePath, "bar/baz"),
	}

	got := []string{}
	err := BreadthFirstDirWalk(sourcePath, func(path string, entry fs.DirEntry, err error) error {
		got = append(got, path)
		index++
		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}
