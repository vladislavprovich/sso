package config

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PostgresConfig struct {
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

func (c PostgresConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Driver, validation.Required),
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.User, validation.Required),
		validation.Field(&c.Password, validation.Required),
		validation.Field(&c.DBName, validation.Required),
		validation.Field(&c.SSLMode, validation.Required),
		validation.Field(&c.MaxConnections, validation.Required),
		validation.Field(&c.MaxIdleConnections, validation.Required),
		validation.Field(&c.ConnMaxLifetime, validation.Required),
	)
}
