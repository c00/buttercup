package fileprovider

import (
	"os"

	"github.com/c00/buttercup/appconfig"
	"github.com/c00/buttercup/logger"
)

const TypeFs = "filesystem"
const TypeEfs = "encrypted-filesystem"
const TypeInMemory = "in-memory"
const TypeS3 = "s3"

func GetProvider(conf appconfig.ProviderConfig) FileProvider {
	switch conf.Type {
	case TypeFs:
		return NewFsProvider(conf)
	case TypeEfs:
		return NewEfsProvider(conf)
	case TypeS3:
		return NewS3Provider(conf)
	case TypeInMemory:
		return NewInMemoryProvider("client")
	}

	logger.Error("unknown provider type: %v\n", conf.Type)
	os.Exit(1)
	return nil
}
