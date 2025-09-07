package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Server       ServerConfig       `mapstructure:"server"`
	Auth         AuthConfig         `mapstructure:"auth"`
	Storage      StorageConfig      `mapstructure:"storage"`
	Cache        CacheConfig        `mapstructure:"cache"`
	Search       SearchConfig       `mapstructure:"search"`
	Email        EmailConfig        `mapstructure:"email"`
	Sheduler     ShedulerConfig     `mapstructure:"sheduler"`
	CloudStorage CloudStorageConfig `mapstructure:"cloudstorage"`
	OAuth        OAuthConfig        `mapstructure:"oauth"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
}

type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           string        `mapstructure:"port"`
	MetricPort     string        `mapstructure:"metricPort"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	IdleTimeout    time.Duration `mapstructure:"idleTimeout"`
	RequestTimeout time.Duration `mapstructure:"requestTimeout"`
	CloseTimeout   time.Duration `mapstructure:"closeTimeout"`
	RateLimit      int           `mapstructure:"rateLimit"`
	Burst          int           `mapstructure:"burst"`
}

type AuthConfig struct {
	Secret       string        `mapstructure:"secret"`
	CodeTTL      time.Duration `mapstructure:"codeTTL"`
	CodeAttempts int           `mapstructure:"codeAttempts"`
	TokenTTL     time.Duration `mapstructure:"tokenTTL"`
}

type OAuthConfig struct {
	GoogleClientId     string        `mapstructure:"googleClientId"`
	GoogleClientSecret string        `mapstructure:"googleClientSecret"`
	GoogleRedirectURL  string        `mapstructure:"googleRedirectURL"`
	GithubClientId     string        `mapstructure:"githubClientId"`
	GithubClientSecret string        `mapstructure:"githubClientSecret"`
	GithubRedirectURL  string        `mapstructure:"githubRedirectURL"`
	StateTTL           time.Duration `mapstructure:"stateTTL"`
}

type StorageConfig struct {
	User           string        `mapstructure:"user"`
	Password       string        `mapstructure:"password"`
	Database       string        `mapstructure:"database"`
	Timezone       string        `mapstructure:"timezone"`
	Host           string        `mapstructure:"host"`
	Port           string        `mapstructure:"port"`
	SSLMode        string        `mapstructure:"sslMode"`
	ConnectTimeout time.Duration `mapstructure:"connectTimeout"`
	PingTimeout    time.Duration `mapstructure:"pingTimeout"`
	AmountOfConns  int32         `mapstructure:"amountOfConns"`
}

type CacheConfig struct {
	Host        string        `mapstructure:"host"`
	Port        string        `mapstructure:"port"`
	Password    string        `mapstructure:"password"`
	PingTimeout time.Duration `mapstructure:"pingTimeout"`
}

type SearchConfig struct {
	Host        string        `mapstructure:"host"`
	Port        string        `mapstructure:"port"`
	User        string        `mapstructure:"user"`
	Password    string        `mapstructure:"password"`
	PingTimeout time.Duration `mapstructure:"pingTimeout"`
}

type EmailConfig struct {
	Name              string `mapstructure:"name"`
	Password          string `mapstructure:"password"`
	Address           string `mapstructure:"address"`
	SmtpAddress       string `mapstructure:"smtpAddress"`
	SmtpServerAddress string `mapstructure:"smtpServerAddress"`
}

type ShedulerConfig struct {
	WidgetBulkTime      time.Duration `mapstructure:"widgetBulkTime"`
	WidgetBulkTimeout   time.Duration `mapstructure:"widgetBulkTimeout"`
	TemplateBulkTime    time.Duration `mapstructure:"templateBulkTime"`
	TemplateBulkTimeout time.Duration `mapstructure:"templateBulkTimeout"`
	CleanCodesTime      time.Duration `mapstructure:"cleanCodesTime"`
	CleanCodesTimeout   time.Duration `mapstructure:"cleanCodesTimeout"`
}

type CloudStorageConfig struct {
	CloudURL string        `mapstructure:"cloudURL"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

func MustLoadConfig(path string) *Config {
	if path == "" {
		panic("config path is empty")
	}
	filename := filepath.Join(path, "config-local.yaml")
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}
	data = []byte(os.ExpandEnv(string(data)))
	v := viper.New()
	v.SetConfigName("config-local")
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.AddConfigPath(path)
	if err := v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %w", err))
	}
	return cfg
}
