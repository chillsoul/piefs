package config

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
)

type Config struct {
	Master
	Storage
	Cache
	Redis
}

type Master struct {
	Host    string
	Port    int
	Replica int
}

type Storage struct {
	Host string
	Port int
	Dir  string
}

type Cache struct {
	NumCounters int64
	MaxCost     int64
	BufferItems int64
}

type Redis struct {
	Host     string
	Port     int
	Password string
	Database int
}

func LoadConfig(path string) (config *Config, err error) {
	var (
		file *os.File
		blob []byte
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	if blob, err = ioutil.ReadAll(file); err != nil {
		return
	}
	config = &Config{}
	err = toml.Unmarshal(blob, config)
	if err != nil {
		return
	}
	return config, nil
}
