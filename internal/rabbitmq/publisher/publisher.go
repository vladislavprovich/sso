package publisher

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/streadway/amqp"
	"github.com/vladislavprovich/sso/internal/config"
)

// Publisher responsible for publishing messages to RabbitMQ.
type Publisher struct {
	Channel  *amqp.Channel // RabbitMQ channel for communication.
	Exchange string        // Exchange name to publish to.
	log      *slog.Logger
}

// NewPublisher creates a new publisher and configures delivery confirmation.
func NewPublisher(conn *amqp.Connection, cfg *config.Config, log *slog.Logger) (*Publisher, error) {
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
		cfg.Rabbit.ExchangeName, // Name Exchange.
		cfg.Rabbit.ExchangeType, // Exchange type (direct – sends messages to specific queues).
		cfg.Rabbit.Durable,      // Durable (remains after reboot).
		cfg.Rabbit.AutoDelet,    // Auto-deleted
		cfg.Rabbit.Internal,     // Internal
		cfg.Rabbit.NoWait,       // No-wait
		nil,                     // Arguments
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		Channel:  ch,
		Exchange: cfg.Rabbit.ExchangeName,
		log:      log,
	}, nil
}

// PublishMessage publishes a delivery confirmation message.
func (p *Publisher) PublishUser(cfg *config.Config, event MessagePublisher) error {
	body, err := json.Marshal(event)
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

	err = p.Channel.Publish(
		p.Exchange,            // Exchange.
		cfg.Rabbit.RoutingKey, // Routing key.
		cfg.Rabbit.Mandatory,  // Mandatory.
		cfg.Rabbit.Immediate,  // Immediate.
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.MessageID,
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
