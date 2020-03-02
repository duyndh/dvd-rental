package config

import (
	"os"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)
//Services represents services config.
type Service struct {
	Name     string    `yaml:"name,omitempty"`
	Database *Database `yaml:"database,omitempty"`
	Cache    *Cache    `yaml:"cache,omitempty"`
}

//Database represents the database config.
type Database struct {
	DBName  string `yaml:"dbName,omitempty"`
	Timeout int    `yaml:"timeout,omitempty"`
	PSN     string `yaml:"psn,omitempty"`
}

//Cache represents the cache config.
type Cache struct {
	Addr     string `yaml:"addr,omitempty"`
	Password string `yaml:"password,omitempty"`
	CacheKey string `yaml:"cacheKey,omitempty"`
}
//Configuration represent app config
type Configuration struct {
	Services []Service `yaml:"services,omitempty"`
}

//loadConfig loads config from yaml file.
func loadConfig(path string) (*Configuration, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading config file, %s", err)
	}
	var cfg = new(Configuration)
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}
	return cfg, nil
}

//Load loads configured environment
func Load(env string) (*Configuration, error) {
	path := fmt.Sprintf("./configs/%s.yaml", env)
	return loadConfig(path)
}