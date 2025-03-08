package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/vladislavprovich/sso/internal/rabbitmq/publisher"
)

type ConvectorToPublisher struct {
}

func NewConvectorToPublisher() *ConvectorToPublisher {
	return &ConvectorToPublisher{}
}

func (c *ConvectorToPublisher) ConvectorToPublisher(userID int64, email, pass string) *publisher.RegisteredUser {
	return &publisher.RegisteredUser{
		MessageID: uuid.New().String(),
		UserID:    userID,
		Email:     email,
		Password:  pass,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
