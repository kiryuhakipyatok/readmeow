package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Search   SearchConfig   `mapstructure:"search"`
	Email    EmailConfig    `mapstructure:"email"`
	Sheduler ShedulerConfig `mapstructure:"sheduler"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	WriteTimeout   int    `mapstructure:"writeTimeout"`
	ReadTimeout    int    `mapstructure:"readTimeout"`
	IdleTimeout    int    `mapstructure:"idleTimeout"`
	RequestTimeout int    `mapstructure:"requestTimeout"`
	CloseTimeout   int    `mapstructure:"closeTimeout"`
}

type AuthConfig struct {
	Secret       string `mapstructure:"secret"`
	CodeTTL      int    `mapstructure:"codeTTL"`
	CodeAttempts int    `mapstructure:"codeAttempts"`
	TokenTTL     int    `mapstructure:"tokenTTL"`
}

type StorageConfig struct {
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Database       string `mapstructure:"database"`
	Timezone       string `mapstructure:"timezone"`
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	SSLMode        string `mapstructure:"sslMode"`
	ConnectTimeout int    `mapstructure:"connectTimeout"`
}

type CacheConfig struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	Password       string `mapstructure:"password"`
	ConnectTimeout int    `mapstructure:"connectTimeout"`
}

type SearchConfig struct {
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Timeout     int    `mapstructure:"timeout"`
	PingTimeout int    `mapstructure:"pingTimeout"`
}

type EmailConfig struct {
	Name              string `mapstructure:"name"`
	Password          string `mapstructure:"password"`
	Address           string `mapstructure:"address"`
	SmtpAddress       string `mapstructure:"smtpAddress"`
	SmtpServerAddress string `mapstructure:"smtpServerAddress"`
}

type ShedulerConfig struct {
	WidgetBulkTime      int `mapstructure:"widgetBulkTime"`
	WidgetBulkTimeout   int `mapstructure:"widgetBulkTimeout"`
	TemplateBulkTime    int `mapstructure:"templateBulkTime"`
	TemplateBulkTimeout int `mapstructure:"templateBulkTimeout"`
	CleanCodesTime      int `mapstructure:"cleanCodesTime"`
	CleanCodesTimeout   int `mapstructure:"cleanCodesTimeout"`
}

func LoadConfig(path string) *Config {
	if path == "" {
		panic("config path is empty")
	}
	filename := filepath.Join(path, "config-local.yaml")
	data, err := os.ReadFile(filename)
	if err != nil {
		panic("failed to read config file: " + err.Error())
	}
	data = []byte(os.ExpandEnv(string(data)))
	v := viper.New()
	v.SetConfigName("config-local")
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.AddConfigPath(path)
	if err := v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		panic("failed to read config: " + err.Error())
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}
	return cfg
}
