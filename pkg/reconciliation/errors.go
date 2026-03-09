package reconciliation

import (
	"fmt"
)

type Reason string

const (
	SubResourceGenerateFailed  Reason = "SubResourceGenerateFailed"
	InternalError              Reason = "InternalError"
	ResourceDependencyNotFound Reason = "ResourceDependencyNotFound"
	UnsupportedTypeResource    Reason = "UnsupportedTypeResource"
	ContainerImageNotFound     Reason = "ContainerImageNotFound"
)

type SubResourceError struct {
	Message   string
	WrapErr   error
	Reason    Reason
	Retryable bool
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

func (e *SubResourceError) GetReason() string {
	return string(e.Reason)
}

func (e *SubResourceError) GetWrapErr() error {
	if e.WrapErr == nil {
		return fmt.Errorf("%s", e.Message)
	}
	return e.WrapErr
}

// IsRetryable returns true if the error is retryable from a reconciler perspective.
func (e *SubResourceError) IsRetryable() bool {
	switch e.Reason {
	// Dependency might become available later (depending on other controllers)
	case ResourceDependencyNotFound:
		return true
	// Will not automatically fix itself
	case InternalError, UnsupportedTypeResource, ContainerImageNotFound:
		return false
	// New/unknown statuses must signal whether they are retriable
	default:
		return e.Retryable
	}
}
