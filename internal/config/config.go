package config

//Services represents services config.
type Services struct {
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
