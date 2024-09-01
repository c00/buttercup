package fileprovider

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryProvider_RunProviderSuite(t *testing.T) {
	RunSuite(t, func() FileProvider {
		return NewInMemoryProvider("client")
	})
}

func TestWriteWithOtherLocks(t *testing.T) {
	p := NewInMemoryProvider("client")
	err := p.Lock()
	assert.Nil(t, err)
	p.name = "some-other-client"

	err = p.StoreFile(FileInfo{Path: "/foo.txt"}, strings.NewReader("foo"))
	assert.NotNil(t, err, "should be locked by someone else")

	err = p.RemoveFile(FileInfo{Path: "/foo.txt"})
	assert.NotNil(t, err, "should be locked by someone else")

	err = p.Unlock()
	assert.NotNil(t, err, "should be locked by someone else")

	p.name = "client"

	err = p.StoreFile(FileInfo{Path: "/foo.txt"}, strings.NewReader("foo"))
	assert.Nil(t, err)

	err = p.RemoveFile(FileInfo{Path: "/foo.txt"})
	assert.Nil(t, err)

	err = p.Unlock()
	assert.Nil(t, err)

}
