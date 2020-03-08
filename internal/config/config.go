package config

import (
	"github.com/spf13/viper"
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

//Load loads configured environment
func Load(env string) (*Configuration, error) {
	viper.SetConfigName(env)
	viper.SetConfigType("yml")
	viper.AddConfigPath("./internal/config")
	viper.AutomaticEnv()
	var cfg = new(Configuration)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
