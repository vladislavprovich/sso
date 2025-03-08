package publisher

import (
	"time"
)

type RegisteredUser struct {
	MessageID string    `json:"message_id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
