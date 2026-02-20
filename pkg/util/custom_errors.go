package util

import "fmt"

type Reason string

const (
	SubResourceGenerateFailed  Reason = "SubResourceGenerateFailed"
	InternalError              Reason = "InternalError"
	ResourceDependencyNotFound Reason = "ResourceDependencyNotFound"
	UnsupportedTypeResource    Reason = "UnsupportedTypeResource"
	ContainerImageNotFound     Reason = "ContainerImageNotFound"
)

type SubResourceError struct {
	Message string
	WrapErr error
	Reason  Reason
}

func (e *SubResourceError) Error() string {
	if e.WrapErr == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.WrapErr)
}

func (e *SubResourceError) Unwrap() error {
	return e.WrapErr
}

func (e *SubResourceError) ReasonString() string {
	return string(e.Reason)
}
