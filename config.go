package gack

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

var (
	ErrConfigKeyNotFound = errors.New("Config key not found")
)

type Config struct {
	PackageName      string                 `yaml:"package_name"`
	Version          string                 `yaml:"version"`
	Maintainer       string                 `yaml:"maintainer"`
	MaintainerEmail  string                 `yaml:"maintainer_email"`
	Homepage         string                 `yaml:"homepage"`
	ShortDescription string                 `yaml:"short_description"`
	Description      string                 `yaml:"description"`
	Targets          map[string]interface{} `yaml:",inline"`
}

func NewConfig() *Config {
	return &Config{
		Targets: make(map[string]interface{}),
	}
}

func decodeConfig(input, output interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  output,
		TagName: "config_name",
	})

	if err == nil {
		err = decoder.Decode(input)
	}
	return err
}

func (c *Config) Get(name string, output interface{}) error {
	input, found := c.Targets[name]
	if found {
		return decodeConfig(input, output)
	}
	return ErrConfigKeyNotFound
}

func ReadConfigFile(filename string) (*Config, error) {
	config := NewConfig()
	configFile, err := os.Open(filename)
	if err == nil {
		err = ReadConfig(config, configFile)
	}
	return config, err
}

func ReadConfig(config *Config, reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
	if err == nil {
		err = yaml.Unmarshal(data, config)
	}

	return err
}

func DefaultConfig(mux Mux) *Config {
	defaultConfig := NewConfig()

	defaultConfig.PackageName = "package"
	defaultConfig.Version = "version"
	defaultConfig.Maintainer = "Maintainer"
	defaultConfig.MaintainerEmail = "maintainer@email.com"
	defaultConfig.Homepage = "http://package.com"
	defaultConfig.ShortDescription = "short description"
	defaultConfig.Description = "description"

	for _, target := range mux.targets {
		if configurable, ok := target.Executable.(DefaultConfigurable); ok {
			name, config := configurable.DefaultConfig()
			defaultConfig.Targets[name] = config
		}
	}
	return defaultConfig
}

func WriteConfig(writer io.Writer, config *Config) error {
	data, err := yaml.Marshal(config)
	if err == nil {
		_, err = writer.Write(data)
	}
	return err
}
