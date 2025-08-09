package config

import "github.com/spf13/viper"

type Config struct {
	App     AppConfig
	Server  ServerConfig
	Auth    AuthConfig
	Storage StorageConfig
	Cache   CacheConfig
}

type AppConfig struct {
	Name    string
	Version string
}

type ServerConfig struct {
	Host           string
	Port           int
	WriteTimeout   int
	ReadTimeout    int
	IdleTimeout    int
	RequestTimeout int
}

type AuthConfig struct {
	Secret   string
	TokenTTL int
}

type StorageConfig struct {
	User     string
	Password string
	Database string
	Timezone string
	Host     string
	Port     int
	SSLMode  string
}

type CacheConfig struct {
	User     string
	Host     string
	Port     string
	Password string
}

type SearchConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Timeout  int
}

func LoadConfig(path string) *Config {
	if path == "" {
		panic("config path is empty")
	}
	v := viper.New()
	v.SetConfigName("config-local.yaml")
	v.AddConfigPath(path)
	if err := v.ReadInConfig(); err != nil {
		panic("failed to read config: " + err.Error())
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}
	return cfg
}
