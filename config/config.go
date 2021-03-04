package config

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
	stream "github.com/adevinta/vulcan-stream"
)

// Config defines required configuration for VulcanStream
type Config struct {
	Logger  stream.LoggerConfig `toml:"Logger"`
	Sender  stream.SenderConfig `toml:"Sender"`
	API     stream.APIConfig    `toml:"API"`
	Storage stream.RedisConfig  `toml:"Storage"`
}

// MustReadConfig reads TOML file with Vulcan Stream configuration
func MustReadConfig(path string) Config {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read configuration file (%v)", err)
	}

	var config Config
	if _, err := toml.Decode(string(configData), &config); err != nil {
		log.Fatalf("Cannot decode configuration file (%v)", err)
	}

	return config
}
