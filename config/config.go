package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	lc := litConfig.New(litConfig.DEFAULT_NETWORK)

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

func Load() *Config {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(filepath.Join(wd, ".getlit", "config.yml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config file not found. Please call `init`")
		return nil
	}

	key, err := os.ReadFile(filepath.Join(wd, ".getlit", "keyfile"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Keyfile not found. Please call `init`")
		return nil
	}

	c := &Config{}
	if err := yaml.Unmarshal(data, c); err != nil {
		panic(err)
	}

	c.LitConfig = litConfig.New(litConfig.DEFAULT_NETWORK)
	c.WorkingDir = wd
	c.PrivateKey = strings.TrimSpace(string(key))

	return c
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
