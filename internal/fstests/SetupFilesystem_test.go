package fstests

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupSourceFilesystem(t *testing.T) {
	assert.Panics(t, func() { SetupSourceFilesystem("", true) }, "foo")
	assert.Panics(t, func() { SetupSourceFilesystem("/not-ending-in-source-folder", true) }, "foo")

	//Create parent folder
	os.Mkdir(path.Join(os.TempDir(), "buttercup-test"), 0755)
	testPath := path.Join(os.TempDir(), "buttercup-test/source")

	err := os.RemoveAll(testPath)
	assert.Nil(t, err)
	SetupSourceFilesystem(testPath, true)

	info, err := os.Stat(testPath)
	assert.Nil(t, err)
	assert.True(t, info.IsDir())
}

func TestSetupDestFilesystem(t *testing.T) {
	assert.Panics(t, func() { SetupDestFilesystem("", true) }, "foo")
	assert.Panics(t, func() { SetupDestFilesystem("/not-ending-in-dest-folder", true) }, "foo")

	//Create parent folder
	os.Mkdir(path.Join(os.TempDir(), "buttercup-test"), 0755)
	testPath := path.Join(os.TempDir(), "buttercup-test/dest")

	err := os.RemoveAll(testPath)
	assert.Nil(t, err)
	SetupDestFilesystem(testPath, true)

	info, err := os.Stat(testPath)
	assert.Nil(t, err)
	assert.True(t, info.IsDir())
}
