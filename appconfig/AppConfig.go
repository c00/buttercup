package appconfig

import (
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFile = `config.yaml`

type AppConfig struct {
	DefaultFolder string         `yaml:"defaultFolder"`
	ClientName    string         `yaml:"clientName"`
	Folders       []FolderConfig `yaml:"folders"`
}

type FolderConfig struct {
	Name   string         `yaml:"name"`
	Local  ProviderConfig `yaml:"local"`
	Remote ProviderConfig `yaml:"remote"`
}

type ProviderConfig struct {
	Type       string             `yaml:"type"`
	FsConfig   *FsProviderConfig  `yaml:"fsConfig,omitempty"`
	EfsConfig  *EfsProviderConfig `yaml:"efsConfig,omitempty"`
	ClientName string             `yaml:"-"`
}

// File System provider
// Local storage, used as 'local'
type FsProviderConfig struct {
	Path string `yaml:"path"`
}

// Encrypted File System Provider
// Used as 'remote' but accessible through some file interface
type EfsProviderConfig struct {
	Path       string `yaml:"path"`
	Passphrase string `yaml:"passphrase,omitempty"`
}

func (c *AppConfig) GetDefault() FolderConfig {
	if c.DefaultFolder == "" {
		panic("No default folder set")
	}

	for _, folder := range c.Folders {
		if folder.Name == c.DefaultFolder {
			folder.Local.ClientName = c.ClientName
			folder.Remote.ClientName = c.ClientName
			return folder
		}
	}

	panic("Specified default folder does not exist")
}

func (c *AppConfig) GetFolder(name string) FolderConfig {
	for _, folder := range c.Folders {
		if folder.Name == name {
			folder.Local.ClientName = c.ClientName
			folder.Remote.ClientName = c.ClientName
			return folder
		}
	}

	panic("No configuration for folder: " + name)
}

func LoadFromUser() (AppConfig, error) {
	u, err := user.Current()
	if err != nil {
		return AppConfig{}, err
	}

	configPath := filepath.Join(u.HomeDir, ".buttercup", configFile)
	return Load(configPath)
}

func Load(path string) (AppConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}

	config := AppConfig{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return AppConfig{}, err
	}

	return config, nil
}
