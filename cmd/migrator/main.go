package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dbURL, migrationsPath string

	flag.StringVar(&dbURL, "db-url", "", "PostgreSQL connection URL")
	flag.StringVar(&migrationsPath, "migrations-path", "", "Path to migration files")
	flag.Parse()

	if dbURL == "" {
		panic("db-url is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New("file://"+migrationsPath, dbURL)
	if err != nil {
		panic(err)
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No new migrations to apply")
			return
		}
		panic(err)
	}

	log.Println("Migrations applied successfully")
}

// Log represents the logger.
type Log struct {
	verbose bool
	log     *slog.Logger
}

// Printf prints out formatted string into a log.
func (l *Log) Printf(format string, v ...interface{}) {
	l.log.Info("Log message", "message", fmt.Sprintf(format, v...))
}

// Verbose shows if verbose print enabled.
func (l *Log) Verbose() bool {
	if l.verbose {
		l.log.Info("Verbose mode enabled")
	}
	return false
}
