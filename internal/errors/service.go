package errors

import "fmt"

type ServiceError interface {
	Error() string
	GetStatus() int
}

type serviceError struct {
	err    error
	status int
}

func (e *serviceError) Unwrap() error {
	return e.err
}

func (e *serviceError) GetStatus() int {
	return e.status
}

func (e *serviceError) Error() string {
	return e.err.Error()
}

func (e *serviceError) String() string {
	return e.err.Error()
}

func NewServiceError(status int, format string, a ...any) ServiceError {
	return &serviceError{fmt.Errorf(format, a...), status}
}
