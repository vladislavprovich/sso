package config

import (
	"flag"
	"os"
	"time"

	"github.com/vladislavprovich/sso/internal/storage/postgres/config"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string                `yaml:"env" default:"local"`
	GRPC           GRPCConfig            `yaml:"grpc"`
	Postgres       config.ConfigPostgres `yaml:"database"`
	MigrationsPath string
	TokenTTL       time.Duration `yaml:"token_ttl" default:"1h"`
	Otel           OtelConfig    `yaml:"otel"`
	Logging        LoggingConfig `yaml:"logging"`
	Tracing        TracingConfig `yaml:"tracing"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type OtelConfig struct {
	Endpoint          string        `yaml:"endpoint"`
	MetricsPort       int           `yaml:"metrics_port"`
	Adr               string        `yaml:"adr"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
}

type LoggingConfig struct {
	Level   string `yaml:"level"`
	Format  string `yaml:"format"`
	LokiURL string `yaml:"loki_url"`
	LogDir  string `yaml:"log_dir"`
}

type TracingConfig struct {
	TempoURL  string `yaml:"tempo_url"`
	NameSpase string `yaml:"namespase"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
