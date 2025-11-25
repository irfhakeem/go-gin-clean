package messaging

import (
	"go-gin-clean/internal/model"

	"github.com/rabbitmq/amqp091-go"
)

type UserPublisher struct {
	registerPublisher      Publisher[model.RegisterEvent]
	resetPasswordPublisher Publisher[model.ResetPasswordEvent]
}

func NewUserPublisher(ch *amqp091.Channel) *UserPublisher {
	return &UserPublisher{
		registerPublisher: Publisher[model.RegisterEvent]{
			ch:       ch,
			exchange: "pc_main_event_bus",
		},
		resetPasswordPublisher: Publisher[model.ResetPasswordEvent]{
			ch:       ch,
			exchange: "pc_main_event_bus",
		},
	}
}

func (p *UserPublisher) RegisterEventPublish(event model.RegisterEvent) error {
	return p.registerPublisher.Publish("user.register", event)
}

func (p *UserPublisher) ResetPasswordEventPublish(event model.ResetPasswordEvent) error {
	return p.resetPasswordPublisher.Publish("user.reset_password", event)
}
