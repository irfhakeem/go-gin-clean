package mq

type NonRetryableError struct {
	error
}

func NonRetryable(err error) *NonRetryableError {
	if err == nil {
		return nil
	}

	// TODO: Check using switch case to make sure the error is not retryable, if it is retryable, return nil

	return &NonRetryableError{err}
}
