package fileprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRandomPath(t *testing.T) {
	str, err := CreateRandomPath()
	assert.Nil(t, err)
	assert.Len(t, str, 26)
}
