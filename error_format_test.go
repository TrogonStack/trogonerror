package trogonerror_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/TrogonStack/trogonerror"
	"github.com/stretchr/testify/assert"
)

func TestTrogonError_ExactFormat_MinimalError(t *testing.T) {
	err := trogonerror.NewError("shopify.core", "SYSTEM_ERROR")

	expected := `unknown error
  visibility: INTERNAL
  domain: shopify.core
  reason: SYSTEM_ERROR
  code: UNKNOWN`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_WithRetryDuration(t *testing.T) {
	err := trogonerror.NewError("shopify.api", "RATE_LIMIT_EXCEEDED",
		trogonerror.WithCode(trogonerror.CodeResourceExhausted),
		trogonerror.WithMessage("API rate limit exceeded"),
		trogonerror.WithRetryInfoDuration(60*time.Second),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "limit", "1000"))

	expected := `API rate limit exceeded
  visibility: INTERNAL
  domain: shopify.api
  reason: RATE_LIMIT_EXCEEDED
  code: RESOURCE_EXHAUSTED
  retryInfo: retryOffset=1m0s
  metadata:
    - limit: 1000 visibility=PUBLIC`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_MetadataOrdering(t *testing.T) {
	err := trogonerror.NewError("shopify.core", "METADATA_SORT_TEST",
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "productType", "digital"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/1234567890"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/5432109876"))

	expected := `unknown error
  visibility: INTERNAL
  domain: shopify.core
  reason: METADATA_SORT_TEST
  code: UNKNOWN
  metadata:
    - customerId: gid://shopify/Customer/1234567890 visibility=PUBLIC
    - orderId: gid://shopify/Order/5432109876 visibility=PUBLIC
    - productType: digital visibility=PUBLIC`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_EmptyOptionalFields(t *testing.T) {
	err := trogonerror.NewError("shopify.core", "RESOURCE_MISSING",
		trogonerror.WithCode(trogonerror.CodeNotFound),
		trogonerror.WithSubject(""),
		trogonerror.WithID(""),
		trogonerror.WithSourceID(""))

	expected := `resource not found
  visibility: INTERNAL
  domain: shopify.core
  reason: RESOURCE_MISSING
  code: NOT_FOUND`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_MultipleHelpLinks(t *testing.T) {
	err := trogonerror.NewError("shopify.support", "HELP_SYSTEM_DOWN",
		trogonerror.WithHelpLink("Contact Support", "https://admin.shopify.com/support/new"),
		trogonerror.WithHelpLink("Check Status", "https://status.shopify.com"),
		trogonerror.WithHelpLink("Retry Request", "https://admin.shopify.com/support/retry"))

	expected := `unknown error
  visibility: INTERNAL
  domain: shopify.support
  reason: HELP_SYSTEM_DOWN
  code: UNKNOWN

- Contact Support: https://admin.shopify.com/support/new
- Check Status: https://status.shopify.com
- Retry Request: https://admin.shopify.com/support/retry`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_DefaultMessage(t *testing.T) {
	err := trogonerror.NewError("shopify.core", "PRODUCT_NOT_FOUND",
		trogonerror.WithCode(trogonerror.CodeNotFound))

	expected := `resource not found
  visibility: INTERNAL
  domain: shopify.core
  reason: PRODUCT_NOT_FOUND
  code: NOT_FOUND`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_CustomMessage(t *testing.T) {
	err := trogonerror.NewError("shopify.core", "ORDER_NOT_FOUND",
		trogonerror.WithCode(trogonerror.CodeNotFound),
		trogonerror.WithMessage("Custom error message"))

	expected := `Custom error message
  visibility: INTERNAL
  domain: shopify.core
  reason: ORDER_NOT_FOUND
  code: NOT_FOUND`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_WithAllOptionalFields(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	retryTime := time.Date(2024, 1, 15, 14, 35, 45, 0, time.UTC)

	err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
		trogonerror.WithCode(trogonerror.CodeInternal),
		trogonerror.WithMessage("Payment processing failed"),
		trogonerror.WithVisibility(trogonerror.VisibilityPrivate),
		trogonerror.WithSubject("/payment/amount"),
		trogonerror.WithID("err_2024_01_15_payment_abc123def"),
		trogonerror.WithTime(timestamp),
		trogonerror.WithSourceID("payment-gateway-prod-cluster-01"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "amount", "299.99"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "currency", "USD"),
		trogonerror.WithRetryTime(retryTime),
		trogonerror.WithHelpLink("Retry Payment", "https://admin.shopify.com/orders/pay_2024_01_15_def456ghi789/retry"))

	expected := `Payment processing failed
  visibility: PRIVATE
  domain: shopify.payments
  reason: PAYMENT_DECLINED
  code: INTERNAL
  id: err_2024_01_15_payment_abc123def
  time: 2024-01-15T14:30:45Z
  subject: /payment/amount
  sourceId: payment-gateway-prod-cluster-01
  retryInfo: retryTime=2024-01-15T14:35:45Z
  metadata:
    - amount: 299.99 visibility=PRIVATE
    - currency: USD visibility=PUBLIC

- Retry Payment: https://admin.shopify.com/orders/pay_2024_01_15_def456ghi789/retry`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_WithStackTrace(t *testing.T) {
	err := trogonerror.NewError("shopify.debugging", "STACK_TRACE_ERROR",
		trogonerror.WithStackTrace(),
		trogonerror.WithDebugDetail("Custom debug detail"))

	errorOutput := err.Error()

	expectedPrefix := `unknown error
  visibility: INTERNAL
  domain: shopify.debugging
  reason: STACK_TRACE_ERROR
  code: UNKNOWN

Custom debug detail`
	assert.True(t, strings.HasPrefix(errorOutput, expectedPrefix))
	assert.Contains(t, errorOutput, ".go:")
	assert.Contains(t, errorOutput, "github.com/TrogonStack/trogonerror")
}

func TestTrogonError_ExactFormat_CompleteErrorWithStackTrace(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	retryTime := time.Date(2024, 1, 15, 14, 35, 45, 0, time.UTC)

	err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
		trogonerror.WithCode(trogonerror.CodeInternal),
		trogonerror.WithMessage("Payment processing failed due to upstream service error"),
		trogonerror.WithVisibility(trogonerror.VisibilityPrivate),
		trogonerror.WithSubject("/payment/amount"),
		trogonerror.WithID("err_2024_01_15_payment_abc123def"),
		trogonerror.WithTime(timestamp),
		trogonerror.WithSourceID("payment-gateway-prod-cluster-01"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "paymentId", "pay_2024_01_15_def456ghi789"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "amount", "299.99"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "currency", "USD"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "merchantId", "gid://shopify/Shop/7890123456"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "transactionType", "purchase"),
		trogonerror.WithRetryTime(retryTime),
		trogonerror.WithHelpLink("Retry Payment", "https://admin.shopify.com/orders/pay_2024_01_15_def456ghi789/retry"),
		trogonerror.WithHelpLink("Contact Support", "https://admin.shopify.com/support/new?payment_id=pay_2024_01_15_def456ghi789"),
		trogonerror.WithStackTrace(),
		trogonerror.WithDebugDetail("Payment gateway integration failure: upstream timeout"))

	errorOutput := err.Error()

	expectedPrefix := `Payment processing failed due to upstream service error
  visibility: PRIVATE
  domain: shopify.payments
  reason: PAYMENT_DECLINED
  code: INTERNAL
  id: err_2024_01_15_payment_abc123def
  time: 2024-01-15T14:30:45Z
  subject: /payment/amount
  sourceId: payment-gateway-prod-cluster-01
  retryInfo: retryTime=2024-01-15T14:35:45Z
  metadata:
    - amount: 299.99 visibility=PRIVATE
    - currency: USD visibility=PUBLIC
    - merchantId: gid://shopify/Shop/7890123456 visibility=INTERNAL
    - paymentId: pay_2024_01_15_def456ghi789 visibility=PRIVATE
    - transactionType: purchase visibility=PUBLIC

- Retry Payment: https://admin.shopify.com/orders/pay_2024_01_15_def456ghi789/retry
- Contact Support: https://admin.shopify.com/support/new?payment_id=pay_2024_01_15_def456ghi789

Payment gateway integration failure: upstream timeout`

	assert.True(t, strings.HasPrefix(errorOutput, expectedPrefix))
	assert.Contains(t, errorOutput, ".go:")
	assert.Contains(t, errorOutput, "github.com/TrogonStack/trogonerror")
}

func TestTrogonError_ExactFormat_WithWrappedStandardError(t *testing.T) {
	originalErr := errors.New("connection timeout after 30s")
	err := trogonerror.NewError("myapp.database", "CONNECTION_FAILED",
		trogonerror.WithCode(trogonerror.CodeUnavailable),
		trogonerror.WithMessage("Failed to connect to database"),
		trogonerror.WithWrap(originalErr))

	expected := `Failed to connect to database
  visibility: INTERNAL
  domain: myapp.database
  reason: CONNECTION_FAILED
  code: UNAVAILABLE

wrapped error: connection timeout after 30s`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_WithWrappedTrogonError(t *testing.T) {
	innerErr := trogonerror.NewError("myapp.network", "DNS_RESOLUTION_FAILED",
		trogonerror.WithCode(trogonerror.CodeInternal),
		trogonerror.WithMessage("DNS lookup failed for db.example.com"))

	outerErr := trogonerror.NewError("myapp.database", "CONNECTION_FAILED",
		trogonerror.WithCode(trogonerror.CodeUnavailable),
		trogonerror.WithMessage("Database connection establishment failed"),
		trogonerror.WithWrap(innerErr))

	expected := `Database connection establishment failed
  visibility: INTERNAL
  domain: myapp.database
  reason: CONNECTION_FAILED
  code: UNAVAILABLE

wrapped error: DNS lookup failed for db.example.com
  visibility: INTERNAL
  domain: myapp.network
  reason: DNS_RESOLUTION_FAILED
  code: INTERNAL`

	assert.Equal(t, expected, outerErr.Error())
}

func TestTrogonError_ExactFormat_NoWrappedError(t *testing.T) {
	err := trogonerror.NewError("myapp.validation", "INVALID_INPUT",
		trogonerror.WithCode(trogonerror.CodeInvalidArgument),
		trogonerror.WithMessage("Email format is invalid"))

	expected := `Email format is invalid
  visibility: INTERNAL
  domain: myapp.validation
  reason: INVALID_INPUT
  code: INVALID_ARGUMENT`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_WrappedErrorWithMetadata(t *testing.T) {
	originalErr := errors.New("invalid JSON at position 42")
	err := trogonerror.NewError("myapp.parser", "PARSE_ERROR",
		trogonerror.WithCode(trogonerror.CodeInvalidArgument),
		trogonerror.WithMessage("Failed to parse request body"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "contentType", "application/json"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "position", "42"),
		trogonerror.WithWrap(originalErr))

	expected := `Failed to parse request body
  visibility: INTERNAL
  domain: myapp.parser
  reason: PARSE_ERROR
  code: INVALID_ARGUMENT
  metadata:
    - contentType: application/json visibility=PUBLIC
    - position: 42 visibility=PRIVATE

wrapped error: invalid JSON at position 42`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_WrappedErrorWithAllFields(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	originalErr := errors.New("network connection reset")

	err := trogonerror.NewError("myapp.service", "SERVICE_ERROR",
		trogonerror.WithCode(trogonerror.CodeInternal),
		trogonerror.WithMessage("Service request failed"),
		trogonerror.WithVisibility(trogonerror.VisibilityPrivate),
		trogonerror.WithID("err_2024_01_15_svc_123"),
		trogonerror.WithTime(timestamp),
		trogonerror.WithSubject("/api/endpoint"),
		trogonerror.WithSourceID("service-node-01"),
		trogonerror.WithHelpLink("Retry Request", "https://docs.example.com/retry"),
		trogonerror.WithWrap(originalErr))

	expected := `Service request failed
  visibility: PRIVATE
  domain: myapp.service
  reason: SERVICE_ERROR
  code: INTERNAL
  id: err_2024_01_15_svc_123
  time: 2024-01-15T14:30:45Z
  subject: /api/endpoint
  sourceId: service-node-01

- Retry Request: https://docs.example.com/retry

wrapped error: network connection reset`

	assert.Equal(t, expected, err.Error())
}

func TestTrogonError_ExactFormat_WrappedErrorBeforeStackTrace(t *testing.T) {
	originalErr := errors.New("database connection timeout")
	err := trogonerror.NewError("myapp.service", "OPERATION_FAILED",
		trogonerror.WithCode(trogonerror.CodeInternal),
		trogonerror.WithMessage("Service operation failed"),
		trogonerror.WithWrap(originalErr),
		trogonerror.WithStackTrace(),
		trogonerror.WithDebugDetail("Debug: Connection pool exhausted"))

	output := err.Error()

	assert.Contains(t, output, "wrapped error: database connection timeout")
	assert.True(t, strings.Index(output, "wrapped error:") < strings.Index(output, "Debug: Connection pool exhausted"))
}
