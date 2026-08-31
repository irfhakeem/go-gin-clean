package errors

type ConsumerError struct {
	Type ConsumerErrorType
	Err  error
}

func (e *ConsumerError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return string(e.Type)
}

func (e *ConsumerError) Unwrap() error {
	return e.Err
}

func NewConsumerError(errType ConsumerErrorType, err error) *ConsumerError {
	return &ConsumerError{
		Type: errType,
		Err:  err,
	}
}
