package fileprovider

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
)

func CreateRandomPath() (string, error) {
	bytes := make([]byte, 12)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("cannot create random path: %w", err)
	}

	randStr := hex.EncodeToString(bytes)
	return path.Join(randStr[:8], randStr[8:16], randStr[16:]), nil
}
