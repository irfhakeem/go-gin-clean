package errors

import (
	"errors"
	"fmt"

	"go-gin-clean/pkg/message"
)

type AppError struct {
	Type AppErrorType
	Key  message.MessageKey
	Err  error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Key, e.Err)
	}

	return fmt.Sprintf("[%s] %s", e.Type, e.Key)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(errType AppErrorType, key message.MessageKey) *AppError {
	return &AppError{
		Type: errType,
		Key:  key,
	}
}

func WrapAppError(errType AppErrorType, key message.MessageKey, err error) *AppError {
	return &AppError{
		Type: errType,
		Key:  key,
		Err:  err,
	}
}

// IsAppError reports whether err is (or wraps) an *AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError returns err itself if it is already (or wraps) an *AppError,
// otherwise it wraps err into a new *AppError with the given type and key.
func AsAppError(errType AppErrorType, key message.MessageKey, err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return &AppError{
		Type: errType,
		Key:  key,
		Err:  err,
	}
}
