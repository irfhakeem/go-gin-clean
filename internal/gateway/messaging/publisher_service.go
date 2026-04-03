package messaging

import "github.com/rabbitmq/amqp091-go"

const (
	UserRegisteredEvent    = "user.register"
	UserResetPasswordEvent = "user.reset_password"
)

type PublisherInterface interface {
	Publish(routingKey string, body []byte) error
}

type PublisherService struct {
	ch       *amqp091.Channel
	exchange string
}

func NewPublisherService(ch *amqp091.Channel, exchange string) *PublisherService {
	return &PublisherService{
		ch:       ch,
		exchange: exchange,
	}
}

func (p *PublisherService) Publish(routingKey string, body []byte) error {
	return p.ch.Publish(
		p.exchange,
		routingKey,
		true,  // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
}
