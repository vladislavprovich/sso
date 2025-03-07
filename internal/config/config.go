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
	Postgres       config.PostgresConfig `yaml:"database"`
	MigrationsPath string
	TokenTTL       time.Duration  `yaml:"token_ttl" default:"1h"`
	Otel           OtelConfig     `yaml:"otel"`
	Logging        LoggingConfig  `yaml:"logging"`
	Tracing        TracingConfig  `yaml:"tracing"`
	Rabbit         RabbitMQConfig `yaml:"rabbit"`
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

type RabbitMQConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	User           string        `yaml:"user"`
	Password       string        `yaml:"password"`
	ExchangeName   string        `yaml:"exchange_name"`
	MaxRetries     uint64        `yaml:"max_retries"`
	MaxElapsedTime time.Duration `yaml:"max_elapsed_time"`

	ExchangeType string `yaml:"exchange_type"`
	Durable      bool   `yaml:"durable" default:"true"`
	AutoDelet    bool   `yaml:"auto_delet" default:"false"`
	Internal     bool   `yaml:"internal" default:"false"`
	NoWait       bool   `yaml:"no_wait" default:"false"`

	RoutingKey string `yaml:"routing_key" default:"sso.integration"`
	Mandatory  bool   `yaml:"mandatory" default:"false"`
	Immediate  bool   `yaml:"immediate" default:"false"`
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
