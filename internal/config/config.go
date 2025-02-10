package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string     `yaml:"env" default:"local"`
	GRPC           GRPCConfig `yaml:"grpc"`
	Database       Database   `yaml:"database"`
	MigrationsPath string
	TokenTTL       time.Duration `yaml:"token_ttl" default:"1h"`
	Otel           Otel          `yaml:"otel"`
	Logging        Logging       `yaml:"logging"`
	Tracing        Tracing       `yaml:"tracing"`
}

type Database struct {
	Driver             string        `yaml:"driver"`
	Host               string        `yaml:"host"`
	Port               int           `yaml:"port"`
	User               string        `yaml:"user"`
	Password           string        `yaml:"password"`
	DBName             string        `yaml:"dbname"`
	SSLMode            string        `yaml:"sslmode"`
	MaxConnections     int           `yaml:"max_connections"`
	MaxIdleConnections int           `yaml:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `yaml:"conn_max_lifetime"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type Otel struct {
	Endpoint    string `yaml:"endpoint"`
	MetricsPort int    `yaml:"metrics_port"`
	Adr         string `yaml:"adr"`
}

type Logging struct {
	Level   string `yaml:"level"`
	Format  string `yaml:"format"`
	LokiURL string `yaml:"loki_url"`
}

type Tracing struct {
	TempoURL string `yaml:"tempo_url"`
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
