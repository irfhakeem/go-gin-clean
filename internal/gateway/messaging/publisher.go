package messaging

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher[T any] struct {
	ch       *amqp091.Channel
	exchange string
}

func (p *Publisher[T]) Publish(routingKey string, message T) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.ch.Publish(
		p.exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
