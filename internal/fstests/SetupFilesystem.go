package fstests

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func SetupSourceFilesystem(path string, createFiles bool) {
	if path == "" {
		panic("path should not me empty")
	} else if !strings.HasSuffix(path, "source") {
		panic("Source path should end in /source to avoid accidentally deleting the wrong files")
	}

	cleanup(path)
	if createFiles {
		createStandardFiles(path)
	}
}

// Create the standard subset of files in the source folder for testing.
func SetupDestFilesystem(path string, createFiles bool) {
	if path == "" {
		panic("path should not me empty")
	} else if !strings.HasSuffix(path, "dest") {
		panic("Dest path should end in /dest to avoid accidentally deleting the wrong files")
	}

	cleanup(path)

	if createFiles {
		createStandardFiles(path)
	}
}

func cleanup(path string) {
	// Delete files and folders.
	err := os.RemoveAll(path)
	if err != nil {
		panic(fmt.Errorf("cannot setup file system, removing failed: %w", err))
	}

	err = os.Mkdir(path, 0755)
	if err != nil {
		panic(fmt.Errorf("cannot setup file system, create folder failed: %w", err))
	}
}

func createStandardFiles(path string) {
	CreateTestFile(path, "/foo.txt")
	CreateTestFile(path, "/file01.md")
	CreateTestFile(path, "/file02.log")
	CreateTestFile(path, "/sub/another.txt")
	CreateTestFile(path, "/sub/andonemore.txt")
}

func CreateTestFile(destPath, filePath string) {
	fullPath := path.Join(destPath, filePath)

	os.MkdirAll(path.Dir(fullPath), 0755)

	err := os.WriteFile(fullPath, []byte(fmt.Sprintf("content for file: %v", filePath)), 0644)
	if err != nil {
		panic(fmt.Errorf("could not write file: %w", err))
	}
}
