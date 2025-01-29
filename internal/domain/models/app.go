package models

type App struct {
	ID     int64
	Name   string
	Secret string `envconfig:"SECRET_KEY"`
}
