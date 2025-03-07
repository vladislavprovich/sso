package publisher

import (
	"time"
)

type MessagePublisher struct {
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
