package port

import "go-gin-clean/internal/application/port/outbound"

type UserFilter = outbound.UserFilter
type FindAllUsersParams = outbound.FindAllUsersParams
type UserRepository = outbound.UserRepository
type RefreshTokenRepository = outbound.RefreshTokenRepository
type OutboxRepository = outbound.OutboxRepository
type TransactionManager = outbound.TransactionManager
