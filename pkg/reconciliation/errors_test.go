package reconciliation

import (
	"errors"
	"testing"
)

func TestSubResourceError_Error(t *testing.T) {
	wrapped := errors.New("wrapped")

	tests := []struct {
		name string
		err  *SubResourceError
		want string
	}{
		{
			name: "without wrapped error",
			err: &SubResourceError{
				Message: "generation failed",
			},
			want: "generation failed",
		},
		{
			name: "with wrapped error",
			err: &SubResourceError{
				Message: "generation failed",
				WrapErr: wrapped,
			},
			want: "generation failed: wrapped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubResourceError_Unwrap(t *testing.T) {
	wrapped := errors.New("wrapped")
	err := &SubResourceError{WrapErr: wrapped}

	if got := err.Unwrap(); !errors.Is(got, wrapped) {
		t.Fatalf("Unwrap() did not return wrapped error")
	}
}

func TestSubResourceError_GetReason(t *testing.T) {
	err := &SubResourceError{Reason: InternalError}

	if got := err.GetReason(); got != "InternalError" {
		t.Fatalf("GetReason() = %q, want %q", got, "InternalError")
	}
}

func TestSubResourceError_GetWrapErr(t *testing.T) {
	wrapped := errors.New("wrapped")

	tests := []struct {
		name  string
		err   *SubResourceError
		check func(t *testing.T, got error)
	}{
		{
			name: "returns wrapped error when present",
			err: &SubResourceError{
				Message: "generation failed",
				WrapErr: wrapped,
			},
			check: func(t *testing.T, got error) {
				t.Helper()
				if !errors.Is(got, wrapped) {
					t.Fatalf("GetWrapErr() did not return wrapped error")
				}
			},
		},
		{
			name: "returns message error when wrap is nil",
			err: &SubResourceError{
				Message: "generation failed",
			},
			check: func(t *testing.T, got error) {
				t.Helper()
				if got == nil {
					t.Fatalf("GetWrapErr() returned nil")
				}
				if got.Error() != "generation failed" {
					t.Fatalf("GetWrapErr().Error() = %q, want %q", got.Error(), "generation failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.err.GetWrapErr())
		})
	}
}

func TestSubResourceError_IsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  *SubResourceError
		want bool
	}{
		{
			name: "resource dependency not found is always retryable",
			err:  &SubResourceError{Reason: ResourceDependencyNotFound, Retryable: false},
			want: true,
		},
		{
			name: "internal error is never retryable",
			err:  &SubResourceError{Reason: InternalError, Retryable: true},
			want: false,
		},
		{
			name: "unsupported type resource is never retryable",
			err:  &SubResourceError{Reason: UnsupportedTypeResource, Retryable: true},
			want: false,
		},
		{
			name: "container image not found is never retryable",
			err:  &SubResourceError{Reason: ContainerImageNotFound, Retryable: true},
			want: false,
		},
		{
			name: "default branch uses retryable field when true",
			err:  &SubResourceError{Reason: SubResourceGenerateFailed, Retryable: true},
			want: true,
		},
		{
			name: "default branch uses retryable field when false",
			err:  &SubResourceError{Reason: SubResourceGenerateFailed, Retryable: false},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsRetryable(); got != tt.want {
				t.Fatalf("IsRetryable() = %t, want %t", got, tt.want)
			}
		})
	}
}
