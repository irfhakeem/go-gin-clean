package messaging

import (
	"go-gin-clean/internal/model"

	"github.com/rabbitmq/amqp091-go"
)

type UserPublisher struct {
	exchange               string
	registerPublisher      Publisher[model.RegisterEvent]
	resetPasswordPublisher Publisher[model.ResetPasswordEvent]
}

func NewUserPublisher(ch *amqp091.Channel, ex string) *UserPublisher {
	return &UserPublisher{
		registerPublisher: Publisher[model.RegisterEvent]{
			ch:       ch,
			exchange: ex,
		},
		resetPasswordPublisher: Publisher[model.ResetPasswordEvent]{
			ch:       ch,
			exchange: ex,
		},
	}
}

func (p *UserPublisher) RegisterEventPublish(event model.RegisterEvent) error {
	return p.registerPublisher.Publish("user.register", event)
}

func (p *UserPublisher) ResetPasswordEventPublish(event model.ResetPasswordEvent) error {
	return p.resetPasswordPublisher.Publish("user.reset_password", event)
}
