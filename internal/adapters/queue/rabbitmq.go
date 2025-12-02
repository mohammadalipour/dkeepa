package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// ScrapeTask represents a task to scrape a product.
type ScrapeTask struct {
	DkpID     string `json:"dkp_id"`
	VariantID string `json:"variant_id,omitempty"`
}

// NewRabbitMQ creates a new RabbitMQ connection.
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
	}, nil
}

// DeclareQueue declares a queue.
func (r *RabbitMQ) DeclareQueue(queueName string) error {
	_, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	return err
}

// Consume starts consuming messages from a queue.
func (r *RabbitMQ) Consume(ctx context.Context, queueName string, handler func(ScrapeTask) error) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Waiting for messages on queue: %s", queueName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			var task ScrapeTask
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Nack(false, false) // Reject and don't requeue
				continue
			}

			log.Printf("Processing task: %+v", task)

			if err := handler(task); err != nil {
				log.Printf("Failed to process task: %v", err)
				msg.Nack(false, true) // Reject and requeue
			} else {
				msg.Ack(false) // Acknowledge
			}
		}
	}
}

// Publish publishes a message to a queue.
func (r *RabbitMQ) Publish(queueName string, task ScrapeTask) error {
	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = r.channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection.
func (r *RabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}
