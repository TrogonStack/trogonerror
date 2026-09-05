// Package errconnect adapts trogonerror to connectrpc.com/connect, translating
// error codes so trogonerror-based services can serve Connect RPC handlers.
package errconnect

import (
	"connectrpc.com/connect"

	"github.com/TrogonStack/trogonerror"
)

// Code maps the two vocabularies. Both are the canonical gRPC set and the
// numbers line up today, but a conversion written as a cast is a silent
// mistranslation the day either side adds a code, and an unknown code has to
// fail towards Internal rather than towards whatever it collides with.
func Code(c trogonerror.Code) connect.Code {
	switch c {
	case trogonerror.CodeCancelled:
		return connect.CodeCanceled
	case trogonerror.CodeUnknown:
		return connect.CodeUnknown
	case trogonerror.CodeInvalidArgument:
		return connect.CodeInvalidArgument
	case trogonerror.CodeDeadlineExceeded:
		return connect.CodeDeadlineExceeded
	case trogonerror.CodeNotFound:
		return connect.CodeNotFound
	case trogonerror.CodeAlreadyExists:
		return connect.CodeAlreadyExists
	case trogonerror.CodePermissionDenied:
		return connect.CodePermissionDenied
	case trogonerror.CodeResourceExhausted:
		return connect.CodeResourceExhausted
	case trogonerror.CodeFailedPrecondition:
		return connect.CodeFailedPrecondition
	case trogonerror.CodeAborted:
		return connect.CodeAborted
	case trogonerror.CodeOutOfRange:
		return connect.CodeOutOfRange
	case trogonerror.CodeUnimplemented:
		return connect.CodeUnimplemented
	case trogonerror.CodeInternal:
		return connect.CodeInternal
	case trogonerror.CodeUnavailable:
		return connect.CodeUnavailable
	case trogonerror.CodeDataLoss:
		return connect.CodeDataLoss
	case trogonerror.CodeUnauthenticated:
		return connect.CodeUnauthenticated
	default:
		return connect.CodeInternal
	}
}
