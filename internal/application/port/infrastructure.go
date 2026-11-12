package port

import "go-gin-clean/internal/application/port/outbound"

type Publisher = outbound.Publisher
type Consumer = outbound.Consumer
type Mailer = outbound.Mailer
type Storage = outbound.Storage
type Cache = outbound.Cache
