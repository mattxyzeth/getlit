package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	litConfig "github.com/crcls/lit-go-sdk/config"
)

type Config struct {
	ChainId    string
	LitConfig  *litConfig.Config `yaml:"-"`
	Network    string
	PrivateKey string `yaml:"-"`
	WorkingDir string `yaml:"-"`
}

var idNameMap = map[string]string{
	"ethereum": "1",
	"polygon":  "137",
	"mumbai":   "80001",
}

func ChainIdForName(name string) string {
	return idNameMap[name]
}

func New(network, privateKey string) *Config {
	lc := litConfig.New(network)

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return &Config{
		ChainId:    ChainIdForName(network),
		LitConfig:  lc,
		Network:    network,
		PrivateKey: privateKey,
		WorkingDir: wd,
	}
}

func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Join(c.WorkingDir, ".getlit"), 0750); err != nil {
		return err
	}

	confFile, err := os.Create(filepath.Join(c.WorkingDir, ".getlit", "config.yml"))
	if err != nil {
		return err
	}
	defer confFile.Close()

	yml, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if _, err := confFile.Write(yml); err != nil {
		return err
	}

	return nil
}
