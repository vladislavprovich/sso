package publisher

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/streadway/amqp"
	"github.com/vladislavprovich/sso/internal/config"
)

type UserPublisher interface {
	PublishUser(msg *RegisteredUser) error
	PublisherClose() error
}

// Publisher responsible for publishing messages to RabbitMQ.
type Publisher struct {
	Channel  *amqp.Channel // RabbitMQ channel for communication.
	Exchange string        // Exchange name to publish to.
	log      *slog.Logger
	cfg      *config.RabbitMQConfig
}

// NewPublisher creates a new publisher and configures delivery confirmation.
func NewPublisher(conn *amqp.Connection, cfg *config.RabbitMQConfig, log *slog.Logger) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Enable confirm mode to confirm message delivery.
	err = ch.Confirm(false)
	if err != nil {
		return nil, err
	}

	log.Info("RabbitMQ publisher started")

	// Declare Exchange if it doesn't exist yet.
	err = ch.ExchangeDeclare(
		cfg.ExchangeName, // Name Exchange.
		cfg.ExchangeType, // Exchange type (direct – sends messages to specific queues).
		cfg.Durable,      // Durable (remains after reboot).
		cfg.AutoDelet,    // Auto-deleted
		cfg.Internal,     // Internal
		cfg.NoWait,       // No-wait
		nil,              // Arguments
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		Channel:  ch,
		Exchange: cfg.ExchangeName,
		log:      log,
		cfg:      cfg,
	}, nil
}

// PublishMessage publishes a delivery confirmation message.
func (p *Publisher) PublishUser(msg *RegisteredUser) error {
	body, err := json.Marshal(msg)
	if err != nil {
		p.log.Error("Error marshalling event:", slog.Any("error", err))
		return err
	}

	confirm := p.Channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	go func() {
		for confirmed := range confirm {
			if !confirmed.Ack {
				p.log.Warn("Failed to confirm message delivery")
			}
		}
	}()

	// Generate unique id for message.
	messageID := uuid.New().String()
	err = p.Channel.Publish(
		p.Exchange,       // Exchange.
		p.cfg.RoutingKey, // Routing key.
		p.cfg.Mandatory,  // Mandatory.
		p.cfg.Immediate,  // Immediate.
		amqp.Publishing{
			ContentType:  p.cfg.ContentType,
			DeliveryMode: amqp.Persistent,
			MessageId:    messageID,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		p.log.Error("Error publishing message:", slog.Any("error", err))
		return err
	}

	p.log.Info("User created event published", slog.String("body", string(body)))
	return nil
}

func (p *Publisher) PublisherClose() error {
	if err := p.Channel.Close(); err != nil {
		p.log.Error("Failed to close RabbitMQ channel", slog.Any("error", err))
		return err
	}
	p.log.Info("RabbitMQ channel closed successfully")
	return nil
}
