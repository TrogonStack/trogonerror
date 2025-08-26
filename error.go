package trogonerror

import (
	"errors"
	"fmt"
	"maps"
	"runtime"
	"slices"
	"strings"
	"time"
)

// SpecVersion represents the version of the error specification
const SpecVersion = 1

// Code represents standardized error codes that map to HTTP status codes
type Code int

const (
	CodeCancelled Code = 1 + iota
	CodeUnknown
	CodeInvalidArgument
	CodeDeadlineExceeded
	CodeNotFound
	CodeAlreadyExists
	CodePermissionDenied
	CodeResourceExhausted
	CodeFailedPrecondition
	CodeAborted
	CodeOutOfRange
	CodeUnimplemented
	CodeInternal
	CodeUnavailable
	CodeDataLoss
	CodeUnauthenticated
)

// Visibility controls information disclosure across trust boundaries
type Visibility int

const (
	VisibilityInternal Visibility = 0
	VisibilityPrivate  Visibility = 1
	VisibilityPublic   Visibility = 2
)

// HelpLink provides documentation link
type HelpLink struct {
	description string
	url         string
}

// Help provides links to relevant documentation
type Help struct {
	links []HelpLink
}

// MetadataValue contains both the value and its visibility level
type MetadataValue struct {
	value      string
	visibility Visibility
}

// Metadata represents a map of metadata with visibility control
type Metadata = map[string]MetadataValue

// DebugInfo contains technical details for internal debugging
type DebugInfo struct {
	stackFrames []runtime.Frame
	detail      string
}

// LocalizedMessage provides translated error message
type LocalizedMessage struct {
	locale  string
	message string
}

// RetryInfo describes when a client can retry a failed request
// Following ADR requirements: servers MUST set either retry_offset OR retry_time, never both
type RetryInfo struct {
	retryOffset *time.Duration
	retryTime   *time.Time
}

// TrogonError represents the standardized error format following the ADR
type TrogonError struct {
	specVersion      int
	code             Code
	message          string
	domain           string
	reason           string
	metadata         Metadata
	causes           []*TrogonError
	visibility       Visibility
	subject          string
	id               string
	time             *time.Time
	help             *Help
	debugInfo        *DebugInfo
	localizedMessage *LocalizedMessage
	retryInfo        *RetryInfo
	sourceID         string
	wrappedErr       error
}

func (e TrogonError) Error() string {
	sb := &strings.Builder{}
	sb.WriteString(strings.TrimSpace(e.Message()))

	fmt.Fprintf(sb, "\n  visibility: %s", e.visibility.String())
	fmt.Fprintf(sb, "\n  domain: %s", e.domain)
	fmt.Fprintf(sb, "\n  reason: %s", e.reason)
	fmt.Fprintf(sb, "\n  code: %s", e.code.String())

	if e.id != "" {
		fmt.Fprintf(sb, "\n  id: %s", e.id)
	}

	if e.time != nil {
		fmt.Fprintf(sb, "\n  time: %s", e.time.Format(time.RFC3339))
	}

	if e.subject != "" {
		fmt.Fprintf(sb, "\n  subject: %s", e.subject)
	}

	if e.sourceID != "" {
		fmt.Fprintf(sb, "\n  sourceId: %s", e.sourceID)
	}

	if e.retryInfo != nil {
		var retryStr string
		if e.retryInfo.retryOffset != nil {
			retryStr = fmt.Sprintf("retryOffset=%s", e.retryInfo.retryOffset.String())
		} else if e.retryInfo.retryTime != nil {
			retryStr = fmt.Sprintf("retryTime=%s", e.retryInfo.retryTime.Format(time.RFC3339))
		}

		fmt.Fprintf(sb, "\n  retryInfo: %s", retryStr)
	}

	if len(e.metadata) > 0 {
		sb.WriteString("\n  metadata:")

		for _, k := range slices.Sorted(maps.Keys(e.metadata)) {
			v := e.metadata[k]
			fmt.Fprintf(sb, "\n    - %s: %s visibility=%s", k, v.value, v.visibility.String())
		}
	}

	if e.help != nil && len(e.help.links) > 0 {
		sb.WriteString("\n\n")
		for i, link := range e.help.links {
			if i > 0 {
				sb.WriteString("\n")
			}
			fmt.Fprintf(sb, "- %s: %s", link.description, link.url)
		}
	}

	if e.wrappedErr != nil {
		sb.WriteString("\n\nwrapped error: ")
		sb.WriteString(e.wrappedErr.Error())
	}

	if e.debugInfo != nil {
		sb.WriteString("\n")
		if e.debugInfo.detail != "" {
			sb.WriteString("\n")
			sb.WriteString(e.debugInfo.detail)
		}

		for _, entry := range e.debugInfo.StackEntries() {
			sb.WriteString("\n")
			sb.WriteString(entry)
		}
	}

	return sb.String()
}

func (e TrogonError) Is(target error) bool {
	switch t := target.(type) {
	case *TrogonError:
		return e.domain == t.domain && e.reason == t.reason
	case TrogonError:
		return e.domain == t.domain && e.reason == t.reason
	default:
		return errors.Is(e.wrappedErr, target)
	}
}

func (e TrogonError) Unwrap() error {
	return e.wrappedErr
}

func (c Code) Message() string {
	switch c {
	case CodeCancelled:
		return "the operation was cancelled"
	case CodeUnknown:
		return "unknown error"
	case CodeInvalidArgument:
		return "invalid argument provided"
	case CodeDeadlineExceeded:
		return "deadline exceeded"
	case CodeNotFound:
		return "resource not found"
	case CodeAlreadyExists:
		return "resource already exists"
	case CodePermissionDenied:
		return "permission denied"
	case CodeUnauthenticated:
		return "unauthenticated"
	case CodeResourceExhausted:
		return "resource exhausted"
	case CodeFailedPrecondition:
		return "failed precondition"
	case CodeAborted:
		return "operation aborted"
	case CodeOutOfRange:
		return "out of range"
	case CodeUnimplemented:
		return "not implemented"
	case CodeInternal:
		return "internal error"
	case CodeUnavailable:
		return "service unavailable"
	case CodeDataLoss:
		return "data loss or corruption"
	default:
		return "unknown error"
	}
}

func (c Code) HttpStatusCode() int {
	switch c {
	case CodeCancelled:
		return 499
	case CodeUnknown:
		return 500
	case CodeInvalidArgument:
		return 400
	case CodeDeadlineExceeded:
		return 504
	case CodeNotFound:
		return 404
	case CodeAlreadyExists:
		return 409
	case CodePermissionDenied:
		return 403
	case CodeResourceExhausted:
		return 429
	case CodeFailedPrecondition:
		return 400
	case CodeAborted:
		return 409
	case CodeOutOfRange:
		return 400
	case CodeUnimplemented:
		return 501
	case CodeInternal:
		return 500
	case CodeUnavailable:
		return 503
	case CodeDataLoss:
		return 500
	case CodeUnauthenticated:
		return 401
	default:
		return 500
	}
}

func (c Code) String() string {
	switch c {
	case CodeCancelled:
		return "CANCELLED"
	case CodeUnknown:
		return "UNKNOWN"
	case CodeInvalidArgument:
		return "INVALID_ARGUMENT"
	case CodeDeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case CodeNotFound:
		return "NOT_FOUND"
	case CodeAlreadyExists:
		return "ALREADY_EXISTS"
	case CodePermissionDenied:
		return "PERMISSION_DENIED"
	case CodeResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case CodeFailedPrecondition:
		return "FAILED_PRECONDITION"
	case CodeAborted:
		return "ABORTED"
	case CodeOutOfRange:
		return "OUT_OF_RANGE"
	case CodeUnimplemented:
		return "UNIMPLEMENTED"
	case CodeInternal:
		return "INTERNAL"
	case CodeUnavailable:
		return "UNAVAILABLE"
	case CodeDataLoss:
		return "DATA_LOSS"
	case CodeUnauthenticated:
		return "UNAUTHENTICATED"
	default:
		return "UNKNOWN"
	}
}

func (v Visibility) String() string {
	switch v {
	case VisibilityInternal:
		return "INTERNAL"
	case VisibilityPrivate:
		return "PRIVATE"
	case VisibilityPublic:
		return "PUBLIC"
	default:
		return "UNKNOWN"
	}
}

// ErrorOption represents options for error construction
type ErrorOption func(*TrogonError)

// NewError creates a new TrogonError following the ADR specification.
// Domain should be a simple identifier like "myapp.users" (not reversed-DNS).
// Reason should be an UPPERCASE identifier like "NOT_FOUND".
func NewError(domain, reason string, options ...ErrorOption) *TrogonError {
	err := &TrogonError{
		specVersion: SpecVersion,
		code:        CodeUnknown,
		message:     "", // empty string means use code's default message
		domain:      domain,
		reason:      reason,
		metadata:    make(Metadata),
		causes:      make([]*TrogonError, 0),
		visibility:  VisibilityInternal,
	}

	for _, option := range options {
		option(err)
	}

	return err
}

// WithCode sets the error code
func WithCode(code Code) ErrorOption {
	return func(e *TrogonError) {
		e.code = code
	}
}

// WithMessage sets the error message
func WithMessage(message string) ErrorOption {
	return func(e *TrogonError) {
		e.message = message
	}
}

// WithMetadata sets metadata with explicit visibility control
func WithMetadata(metadata map[string]MetadataValue) ErrorOption {
	return func(e *TrogonError) {
		maps.Copy(e.metadata, metadata)
	}
}

// WithMetadataValue sets a single metadata entry with specific visibility
func WithMetadataValue(visibility Visibility, key, value string) ErrorOption {
	return func(e *TrogonError) {
		addMetadataValue(e, visibility, key, value)
	}
}

// WithMetadataValuef sets a single metadata entry with printf-style formatting for the value
// Example: WithMetadataValuef(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/%s", orderID)
func WithMetadataValuef(visibility Visibility, key, valueFormat string, args ...any) ErrorOption {
	return func(e *TrogonError) {
		addMetadataValue(e, visibility, key, fmt.Sprintf(valueFormat, args...))
	}
}

// WithVisibility sets the error visibility
func WithVisibility(visibility Visibility) ErrorOption {
	return func(e *TrogonError) {
		e.visibility = visibility
	}
}

// WithSubject sets the error subject
func WithSubject(subject string) ErrorOption {
	return func(e *TrogonError) {
		e.subject = subject
	}
}

// WithID sets the error ID
func WithID(id string) ErrorOption {
	return func(e *TrogonError) {
		e.id = id
	}
}

// WithTime sets the error timestamp
func WithTime(timestamp time.Time) ErrorOption {
	return func(e *TrogonError) {
		e.time = &timestamp
	}
}

// WithSourceID sets the source ID
func WithSourceID(sourceID string) ErrorOption {
	return func(e *TrogonError) {
		e.sourceID = sourceID
	}
}

// WithHelp sets the help information
func WithHelp(help Help) ErrorOption {
	return func(e *TrogonError) {
		e.help = &help
	}
}

// WithHelpLink adds a help link with a static URL.
// Use WithHelpLinkf for URLs that need parameter interpolation.
func WithHelpLink(description, url string) ErrorOption {
	return func(e *TrogonError) {
		addHelpLink(e, description, url)
	}
}

// WithHelpLinkf adds a help link with printf-style formatting for the URL.
// Example: WithHelpLinkf("User Console", "https://console.myapp.com/users/%s", userID)
func WithHelpLinkf(description, urlFormat string, args ...any) ErrorOption {
	return func(e *TrogonError) {
		addHelpLink(e, description, fmt.Sprintf(urlFormat, args...))
	}
}

// WithDebugInfo sets debug information (for internal use only)
func WithDebugInfo(debugInfo DebugInfo) ErrorOption {
	return func(e *TrogonError) {
		e.debugInfo = &debugInfo
	}
}

// WithStackTrace annotates the error with a stack trace at the point WithStackTrace was called
// This captures the current call stack for debugging purposes (internal use only)
func WithStackTrace() ErrorOption {
	return WithStackTraceDepth(32) // Default depth
}

// WithDebugDetail sets debug detail message without capturing stack trace
func WithDebugDetail(detail string) ErrorOption {
	return func(e *TrogonError) {
		if e.debugInfo == nil {
			e.debugInfo = &DebugInfo{detail: detail}
		} else {
			e.debugInfo.detail = detail
		}
	}
}

// WithStackTraceDepth annotates the error with a stack trace up to the specified depth
func WithStackTraceDepth(maxDepth int) ErrorOption {
	return func(e *TrogonError) {
		stackFrames := captureStackTrace(2, maxDepth) // Skip WithStackTraceDepth and the calling ErrorOption wrapper
		if e.debugInfo == nil {
			e.debugInfo = &DebugInfo{
				stackFrames: stackFrames,
			}
		} else {
			e.debugInfo.stackFrames = stackFrames
		}
	}
}

// captureStackTrace captures the current call stack up to maxDepth frames
func captureStackTrace(skip, maxDepth int) []runtime.Frame {
	if maxDepth <= 0 {
		maxDepth = 32 // Reasonable default
	}

	var pcs = make([]uintptr, maxDepth)
	n := runtime.Callers(skip, pcs[:])

	frames := runtime.CallersFrames(pcs[:n])
	var stackFrames []runtime.Frame

	for {
		frame, more := frames.Next()
		stackFrames = append(stackFrames, frame)

		if !more {
			break
		}
	}

	return stackFrames
}

// WithLocalizedMessage sets localized message
func WithLocalizedMessage(locale, message string) ErrorOption {
	return func(e *TrogonError) {
		e.localizedMessage = &LocalizedMessage{
			locale:  locale,
			message: message,
		}
	}
}

// WithRetryInfoDuration sets retry information with a duration offset
// Following ADR: servers MUST set either retry_offset OR retry_time, never both
func WithRetryInfoDuration(retryOffset time.Duration) ErrorOption {
	return func(e *TrogonError) {
		e.retryInfo = &RetryInfo{
			retryOffset: &retryOffset,
			retryTime:   nil, // Explicitly ensure only one is set per ADR
		}
	}
}

// WithRetryTime sets retry information with an absolute time
// Following ADR: servers MUST set either retry_offset OR retry_time, never both
func WithRetryTime(retryTime time.Time) ErrorOption {
	return func(e *TrogonError) {
		e.retryInfo = &RetryInfo{
			retryOffset: nil, // Explicitly ensure only one is set per ADR
			retryTime:   &retryTime,
		}
	}
}

// WithCause adds one or more causes to the error
func WithCause(causes ...*TrogonError) ErrorOption {
	return func(e *TrogonError) {
		e.causes = append(e.causes, causes...)
	}
}

// WithErrorMessage sets the error message to the error's Error() string
func WithErrorMessage(err error) ErrorOption {
	return func(e *TrogonError) {
		e.message = err.Error()
	}
}

// WithWrap wraps an existing error
func WithWrap(err error) ErrorOption {
	return func(e *TrogonError) {
		e.wrappedErr = err
	}
}

func (e *TrogonError) copy() *TrogonError {
	clonedErr := &TrogonError{
		specVersion:      e.specVersion,
		code:             e.code,
		message:          e.message,
		domain:           e.domain,
		reason:           e.reason,
		visibility:       e.visibility,
		subject:          e.subject,
		id:               e.id,
		time:             e.time,
		sourceID:         e.sourceID,
		retryInfo:        e.retryInfo,
		localizedMessage: e.localizedMessage,
		wrappedErr:       e.wrappedErr,
	}

	if len(e.metadata) > 0 {
		clonedErr.metadata = make(Metadata, len(e.metadata))
		for k, v := range e.metadata {
			clonedErr.metadata[k] = v
		}
	}

	if len(e.causes) > 0 {
		clonedErr.causes = make([]*TrogonError, len(e.causes))
		copy(clonedErr.causes, e.causes)
	}

	if e.help != nil {
		helpCopy := e.help.copy()
		clonedErr.help = &helpCopy
	}
	if e.debugInfo != nil {
		debugInfoCopy := e.debugInfo.copy()
		clonedErr.debugInfo = &debugInfoCopy
	}

	return clonedErr
}

// ChangeOption represents a change to apply to a TrogonError
type ChangeOption func(*TrogonError)

// WithChanges applies multiple changes in a single copy operation for efficiency
func (e *TrogonError) WithChanges(changes ...ChangeOption) *TrogonError {
	clonedErr := e.copy()
	for _, change := range changes {
		change(clonedErr)
	}
	return clonedErr
}

// Change options for error mutation

// WithChangeMetadata sets metadata with explicit visibility control
func WithChangeMetadata(metadata map[string]MetadataValue) ChangeOption {
	return func(e *TrogonError) {
		e.metadata = make(Metadata)
		maps.Copy(e.metadata, metadata)
	}
}

// WithChangeMetadataValue sets a single metadata entry with specific visibility
func WithChangeMetadataValue(visibility Visibility, key, value string) ChangeOption {
	return func(e *TrogonError) {
		addMetadataValue(e, visibility, key, value)
	}
}

// WithChangeMetadataValuef sets a single metadata entry with printf-style formatting for the value
// Example: WithChangeMetadataValuef(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/%s", orderID)
func WithChangeMetadataValuef(visibility Visibility, key, valueFormat string, args ...any) ChangeOption {
	return func(e *TrogonError) {
		addMetadataValue(e, visibility, key, fmt.Sprintf(valueFormat, args...))
	}
}

// WithChangeID sets the error ID
func WithChangeID(id string) ChangeOption {
	return func(e *TrogonError) {
		e.id = id
	}
}

// WithChangeTime sets the timestamp
func WithChangeTime(timestamp time.Time) ChangeOption {
	return func(e *TrogonError) {
		e.time = &timestamp
	}
}

// WithChangeSourceID sets the source ID
func WithChangeSourceID(sourceID string) ChangeOption {
	return func(e *TrogonError) {
		e.sourceID = sourceID
	}
}

// WithChangeHelpLink adds a help link with a static URL (appends to existing help).
// Use WithChangeHelpLinkf for URLs that need parameter interpolation.
func WithChangeHelpLink(description, url string) ChangeOption {
	return func(e *TrogonError) {
		addHelpLink(e, description, url)
	}
}

// WithChangeHelpLinkf adds a help link with printf-style formatting for the URL (appends to existing help).
// Example: WithChangeHelpLinkf("Order Details", "https://console.myapp.com/orders/%s", orderID)
func WithChangeHelpLinkf(description, urlFormat string, args ...any) ChangeOption {
	return func(e *TrogonError) {
		addHelpLink(e, description, fmt.Sprintf(urlFormat, args...))
	}
}

// WithChangeRetryInfoDuration sets retry duration (replaces existing retry info)
func WithChangeRetryInfoDuration(retryOffset time.Duration) ChangeOption {
	return func(e *TrogonError) {
		e.retryInfo = &RetryInfo{
			retryOffset: &retryOffset,
		}
	}
}

// WithChangeRetryTime sets absolute retry time (replaces existing retry info)
func WithChangeRetryTime(retryTime time.Time) ChangeOption {
	return func(e *TrogonError) {
		e.retryInfo = &RetryInfo{
			retryTime: &retryTime,
		}
	}
}

// WithChangeLocalizedMessage sets localized message (replaces existing)
func WithChangeLocalizedMessage(locale, message string) ChangeOption {
	return func(e *TrogonError) {
		e.localizedMessage = &LocalizedMessage{
			locale:  locale,
			message: message,
		}
	}
}

func (e TrogonError) SpecVersion() int { return e.specVersion }
func (e TrogonError) Code() Code       { return e.code }
func (e TrogonError) Message() string {
	if e.message != "" {
		return e.message
	}
	return e.code.Message()
}
func (e TrogonError) Domain() string                      { return e.domain }
func (e TrogonError) Reason() string                      { return e.reason }
func (e TrogonError) Metadata() Metadata                  { return e.metadata }
func (e TrogonError) Causes() []*TrogonError              { return e.causes }
func (e TrogonError) Visibility() Visibility              { return e.visibility }
func (e TrogonError) Subject() string                     { return e.subject }
func (e TrogonError) ID() string                          { return e.id }
func (e TrogonError) Time() *time.Time                    { return e.time }
func (e TrogonError) Help() *Help                         { return e.help }
func (e TrogonError) DebugInfo() *DebugInfo               { return e.debugInfo }
func (e TrogonError) LocalizedMessage() *LocalizedMessage { return e.localizedMessage }
func (e TrogonError) RetryInfo() *RetryInfo               { return e.retryInfo }
func (e TrogonError) SourceID() string                    { return e.sourceID }

func (m MetadataValue) Value() string          { return m.value }
func (m MetadataValue) Visibility() Visibility { return m.visibility }

func (h HelpLink) Description() string { return h.description }
func (h HelpLink) URL() string         { return h.url }

func (h Help) copy() Help {
	if len(h.links) == 0 {
		return Help{}
	}
	// HelpLink is small immutable struct, use built-in copy for efficiency
	copiedLinks := make([]HelpLink, len(h.links))
	copy(copiedLinks, h.links)
	return Help{links: copiedLinks}
}

func (h Help) Links() []HelpLink { return h.links }

func (d DebugInfo) copy() DebugInfo {
	if len(d.stackFrames) == 0 {
		return DebugInfo{detail: d.detail}
	}
	copiedStackFrames := make([]runtime.Frame, len(d.stackFrames))
	copy(copiedStackFrames, d.stackFrames)
	return DebugInfo{
		stackFrames: copiedStackFrames,
		detail:      d.detail,
	}
}

// StackEntries converts the runtime.Frame objects to formatted strings
func (d DebugInfo) StackEntries() []string {
	if len(d.stackFrames) == 0 {
		return nil
	}

	entries := make([]string, len(d.stackFrames))
	for i, frame := range d.stackFrames {
		entries[i] = fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function)
	}
	return entries
}

// StackFrames returns the raw runtime.Frame objects for advanced use cases
func (d DebugInfo) StackFrames() []runtime.Frame {
	if len(d.stackFrames) == 0 {
		return nil
	}

	// Return a copy to prevent mutation
	frames := make([]runtime.Frame, len(d.stackFrames))
	copy(frames, d.stackFrames)
	return frames
}

func (d DebugInfo) Detail() string { return d.detail }

func (l LocalizedMessage) Locale() string  { return l.locale }
func (l LocalizedMessage) Message() string { return l.message }

func (r RetryInfo) RetryOffset() *time.Duration { return r.retryOffset }
func (r RetryInfo) RetryTime() *time.Time       { return r.retryTime }

// ErrorTemplate represents a reusable error definition
type ErrorTemplate struct {
	domain     string
	reason     string
	code       Code
	message    string // empty string means use code's default message
	visibility Visibility
	help       *Help
}

// TemplateOption represents options that can be applied to ErrorTemplate
type TemplateOption func(*ErrorTemplate)

// NewErrorTemplate creates a reusable error template for consistent error creation.
func NewErrorTemplate(domain, reason string, options ...TemplateOption) *ErrorTemplate {
	template := &ErrorTemplate{
		domain:     domain,
		reason:     reason,
		code:       CodeUnknown,
		message:    "", // empty string means use code's default message
		visibility: VisibilityInternal,
	}

	for _, option := range options {
		option(template)
	}

	return template
}

// Template option functions
func TemplateWithCode(code Code) TemplateOption {
	return func(t *ErrorTemplate) {
		t.code = code
	}
}

func TemplateWithMessage(message string) TemplateOption {
	return func(t *ErrorTemplate) {
		t.message = message
	}
}

func TemplateWithVisibility(visibility Visibility) TemplateOption {
	return func(t *ErrorTemplate) {
		t.visibility = visibility
	}
}

func TemplateWithHelp(help Help) TemplateOption {
	return func(t *ErrorTemplate) {
		t.help = &help
	}
}

func TemplateWithHelpLink(description, url string) TemplateOption {
	return func(t *ErrorTemplate) {
		if t.help == nil {
			t.help = &Help{}
		}
		t.help.links = append(t.help.links, HelpLink{
			description: description,
			url:         url,
		})
	}
}

// NewError creates a new error instance from the template
func (et *ErrorTemplate) NewError(options ...ErrorOption) *TrogonError {
	baseOptions := []ErrorOption{
		WithCode(et.code),
		WithVisibility(et.visibility)}

	if et.message != "" {
		baseOptions = append(baseOptions, WithMessage(et.message))
	}
	if et.help != nil {
		baseOptions = append(baseOptions, WithHelp(*et.help))
	}

	return NewError(et.domain, et.reason, append(baseOptions, options...)...)
}

// Is checks if the given error matches this template's domain and reason
// This allows checking if an error was created from this template without requiring
// the template to implement the error interface
func (et *ErrorTemplate) Is(target error) bool {
	switch t := target.(type) {
	case *TrogonError:
		return et.domain == t.domain && et.reason == t.reason
	case TrogonError:
		return et.domain == t.domain && et.reason == t.reason
	default:
		return false
	}
}

func addHelpLink(e *TrogonError, description, url string) {
	if e.help == nil {
		e.help = &Help{}
	}
	e.help.links = append(e.help.links, HelpLink{
		description: description,
		url:         url,
	})
}

func addMetadataValue(e *TrogonError, visibility Visibility, key, value string) {
	if len(e.metadata) == 0 {
		e.metadata = make(Metadata)
	}
	e.metadata[key] = MetadataValue{value: value, visibility: visibility}
}
