package errconnect_test

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"

	"github.com/TrogonStack/trogonerror"
	"github.com/TrogonStack/trogonerror/errconnect"
)

func TestCode(t *testing.T) {
	tests := []struct {
		name string
		code trogonerror.Code
		want connect.Code
	}{
		{"CANCELLED", trogonerror.CodeCancelled, connect.CodeCanceled},
		{"UNKNOWN", trogonerror.CodeUnknown, connect.CodeUnknown},
		{"INVALID_ARGUMENT", trogonerror.CodeInvalidArgument, connect.CodeInvalidArgument},
		{"DEADLINE_EXCEEDED", trogonerror.CodeDeadlineExceeded, connect.CodeDeadlineExceeded},
		{"NOT_FOUND", trogonerror.CodeNotFound, connect.CodeNotFound},
		{"ALREADY_EXISTS", trogonerror.CodeAlreadyExists, connect.CodeAlreadyExists},
		{"PERMISSION_DENIED", trogonerror.CodePermissionDenied, connect.CodePermissionDenied},
		{"RESOURCE_EXHAUSTED", trogonerror.CodeResourceExhausted, connect.CodeResourceExhausted},
		{"FAILED_PRECONDITION", trogonerror.CodeFailedPrecondition, connect.CodeFailedPrecondition},
		{"ABORTED", trogonerror.CodeAborted, connect.CodeAborted},
		{"OUT_OF_RANGE", trogonerror.CodeOutOfRange, connect.CodeOutOfRange},
		{"UNIMPLEMENTED", trogonerror.CodeUnimplemented, connect.CodeUnimplemented},
		{"INTERNAL", trogonerror.CodeInternal, connect.CodeInternal},
		{"UNAVAILABLE", trogonerror.CodeUnavailable, connect.CodeUnavailable},
		{"DATA_LOSS", trogonerror.CodeDataLoss, connect.CodeDataLoss},
		{"UNAUTHENTICATED", trogonerror.CodeUnauthenticated, connect.CodeUnauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, errconnect.Code(tt.code))
		})
	}

	t.Run("unknown code falls back to Internal", func(t *testing.T) {
		var unknownCode trogonerror.Code = 999
		assert.Equal(t, connect.CodeInternal, errconnect.Code(unknownCode))
	})
}
