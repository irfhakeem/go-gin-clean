package errors

import "errors"

type AppErrorType string
type ConsumerErrorType string

const (
	Unauthorized  AppErrorType = "UNAUTHORIZED"
	Forbidden     AppErrorType = "FORBIDDEN"
	BadRequest    AppErrorType = "BAD_REQUEST"
	NotFound      AppErrorType = "NOT_FOUND"
	Conflict      AppErrorType = "CONFLICT"
	Internal      AppErrorType = "INTERNAL"
	Unprocessable AppErrorType = "UNPROCESSABLE"

	Retryable    ConsumerErrorType = "RETRYABLE"
	NonRetryable ConsumerErrorType = "NON_RETRYABLE"
)

// ErrCacheMiss is a sentinel error that indicates a cached key was not found.
var ErrCacheMiss = errors.New("cache miss")
