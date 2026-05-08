package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	queue      string
	routingKey string
}

func NewRabbitMQ(url, exchange, queue, routingKey string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	if err := channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq declare exchange: %w", err)
	}

	if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq declare queue: %w", err)
	}

	if err := channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq bind queue: %w", err)
	}

	return &RabbitMQ{
		conn:       conn,
		channel:    channel,
		exchange:   exchange,
		queue:      queue,
		routingKey: routingKey,
	}, nil
}

func (r *RabbitMQ) PublishWeeklyReportJob(ctx context.Context, job ports.WeeklyReportJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal weekly report job: %w", err)
	}

	err = r.channel.PublishWithContext(ctx, r.exchange, r.routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("publish weekly report job: %w", err)
	}

	return nil
}

func (r *RabbitMQ) ConsumeWeeklyReportJobs(ctx context.Context, processor func(context.Context, ports.WeeklyReportJob) error) error {
	msgs, err := r.channel.Consume(r.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume weekly report jobs: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var job ports.WeeklyReportJob
				if err := json.Unmarshal(msg.Body, &job); err != nil {
					log.Printf("invalid report job payload: %v", err)
					_ = msg.Nack(false, false)
					continue
				}

				if err := processor(ctx, job); err != nil {
					log.Printf("report job failed request_id=%s: %v", job.RequestID, err)
					_ = msg.Nack(false, true)
					continue
				}

				_ = msg.Ack(false)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) Close() error {
	var closeErr error
	if r.channel != nil {
		closeErr = r.channel.Close()
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}
