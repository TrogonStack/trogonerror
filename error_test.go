package trogonerror_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TrogonStack/trogonerror"
	"github.com/stretchr/testify/assert"
)

func TestTrogonErrorCreation(t *testing.T) {
	t.Run("Basic error creation with default values", func(t *testing.T) {
		err := trogonerror.NewError("shopify.users", "NOT_FOUND",
			trogonerror.WithCode(trogonerror.CodeNotFound))

		assert.Equal(t, trogonerror.SpecVersion, err.SpecVersion())
		assert.Equal(t, trogonerror.CodeNotFound, err.Code())
		assert.Equal(t, "resource not found", err.Message())
		assert.Equal(t, "shopify.users", err.Domain())
		assert.Equal(t, "NOT_FOUND", err.Reason())
		assert.Equal(t, trogonerror.VisibilityInternal, err.Visibility())
	})
}

func TestTrogonErrorOptions(t *testing.T) {
	t.Run("WithMetadata adds metadata values", func(t *testing.T) {
		err := trogonerror.NewError("shopify.orders", "INVALID_ORDER_DATA",
			trogonerror.WithCode(trogonerror.CodeInvalidArgument),
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/5432109876"),
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "requestId", "req_2024_01_15_abc123def456"))

		assert.Equal(t, "gid://shopify/Order/5432109876", err.Metadata()["orderId"].Value())
		assert.Equal(t, "req_2024_01_15_abc123def456", err.Metadata()["requestId"].Value())
	})

	t.Run("WithMetadata function adds multiple metadata at once", func(t *testing.T) {
		metadata := make(map[string]trogonerror.MetadataValue)

		baseErr := trogonerror.NewError("shopify.inventory", "NO_STOCK")
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "productId", "gid://shopify/Product/7890123456")(baseErr)
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "inventoryItemId", "gid://shopify/InventoryItem/1234567890")(baseErr)

		for k, v := range baseErr.Metadata() {
			metadata[k] = v
		}

		err := trogonerror.NewError("shopify.cache", "CACHE_MISS",
			trogonerror.WithMetadata(metadata))

		assert.Equal(t, "gid://shopify/Product/7890123456", err.Metadata()["productId"].Value())
		assert.Equal(t, trogonerror.VisibilityPublic, err.Metadata()["productId"].Visibility())
		assert.Equal(t, "gid://shopify/InventoryItem/1234567890", err.Metadata()["inventoryItemId"].Value())
		assert.Equal(t, trogonerror.VisibilityPrivate, err.Metadata()["inventoryItemId"].Visibility())
	})

	t.Run("WithOptionalFields sets various optional fields", func(t *testing.T) {
		timestamp := time.Now()
		err := trogonerror.NewError("shopify.validation", "INVALID_EMAIL",
			trogonerror.WithSubject("/email"),
			trogonerror.WithID("err_2024_01_15_validation_abc123"),
			trogonerror.WithTime(timestamp),
			trogonerror.WithSourceID("validation-service-prod-01"))

		assert.NotEmpty(t, err.Subject())
		assert.Equal(t, "/email", err.Subject())
		assert.NotEmpty(t, err.ID())
		assert.Equal(t, "err_2024_01_15_validation_abc123", err.ID())
		assert.NotNil(t, err.Time())
		assert.True(t, err.Time().Equal(timestamp))
		assert.NotEmpty(t, err.SourceID())
		assert.Equal(t, "validation-service-prod-01", err.SourceID())
	})
}

func TestTrogonErrorHelp(t *testing.T) {
	t.Run("WithHelpLink adds help resolution links", func(t *testing.T) {
		err := trogonerror.NewError("shopify.orders", "ORDER_FAILED",
			trogonerror.WithHelpLink("Retry Order", "https://admin.shopify.com/orders/5432109876/retry"),
			trogonerror.WithHelpLink("Contact Support", "https://admin.shopify.com/support/new?order_id=5432109876"))

		assert.NotNil(t, err.Help())
		assert.Len(t, err.Help().Links(), 2)
		assert.Equal(t, "Retry Order", err.Help().Links()[0].Description())
		assert.Equal(t, "https://admin.shopify.com/orders/5432109876/retry", err.Help().Links()[0].URL())
		assert.Equal(t, "Contact Support", err.Help().Links()[1].Description())
		assert.Equal(t, "https://admin.shopify.com/support/new?order_id=5432109876", err.Help().Links()[1].URL())
		assert.NotEmpty(t, err.Help().Links())
	})

	t.Run("WithHelpLinkf adds formatted help resolution links", func(t *testing.T) {
		userID := "1234567890"
		orderID := "5432109876"
		err := trogonerror.NewError("shopify.payments", "PAYMENT_FAILED",
			trogonerror.WithHelpLinkf("Fix Payment", "https://admin.shopify.com/customers/%s/payment-methods", userID),
			trogonerror.WithHelpLinkf("Retry Order", "https://admin.shopify.com/orders/%s/retry", orderID))

		assert.NotNil(t, err.Help())
		assert.Len(t, err.Help().Links(), 2)
		assert.Equal(t, "Fix Payment", err.Help().Links()[0].Description())
		assert.Equal(t, "https://admin.shopify.com/customers/1234567890/payment-methods", err.Help().Links()[0].URL())
		assert.Equal(t, "Retry Order", err.Help().Links()[1].Description())
		assert.Equal(t, "https://admin.shopify.com/orders/5432109876/retry", err.Help().Links()[1].URL())
	})
}

func TestTrogonErrorMetadataValuef(t *testing.T) {
	t.Run("WithMetadataValuef formats metadata values", func(t *testing.T) {
		userID := "1234567890"
		orderID := "5432109876"
		productID := "9876543210"
		err := trogonerror.NewError("shopify.orders", "ORDER_CREATED",
			trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/%s", userID),
			trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/%s", orderID),
			trogonerror.WithMetadataValuef(trogonerror.VisibilityInternal, "productId", "gid://shopify/Product/%s", productID))

		assert.Equal(t, "gid://shopify/Customer/1234567890", err.Metadata()["customerId"].Value())
		assert.Equal(t, trogonerror.VisibilityPublic, err.Metadata()["customerId"].Visibility())
		assert.Equal(t, "gid://shopify/Order/5432109876", err.Metadata()["orderId"].Value())
		assert.Equal(t, trogonerror.VisibilityPublic, err.Metadata()["orderId"].Visibility())
		assert.Equal(t, "gid://shopify/Product/9876543210", err.Metadata()["productId"].Value())
		assert.Equal(t, trogonerror.VisibilityInternal, err.Metadata()["productId"].Visibility())
	})

	t.Run("WithChangeMetadataValuef formats metadata values during change", func(t *testing.T) {
		original := trogonerror.NewError("shopify.inventory", "STOCK_CHECK",
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "productId", "gid://shopify/Product/1111111111"))

		userID := "2222222222"
		warehouseID := "nyc-01"
		modified := original.WithChanges(
			trogonerror.WithChangeMetadataValuef(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/%s", userID),
			trogonerror.WithChangeMetadataValuef(trogonerror.VisibilityInternal, "warehouseId", "warehouse-%s", warehouseID))

		// Original should be unchanged
		assert.Equal(t, "gid://shopify/Product/1111111111", original.Metadata()["productId"].Value())
		assert.NotContains(t, original.Metadata(), "customerId")
		assert.NotContains(t, original.Metadata(), "warehouseId")

		// Modified should have new formatted values
		assert.Equal(t, "gid://shopify/Product/1111111111", modified.Metadata()["productId"].Value())
		assert.Equal(t, "gid://shopify/Customer/2222222222", modified.Metadata()["customerId"].Value())
		assert.Equal(t, trogonerror.VisibilityPublic, modified.Metadata()["customerId"].Visibility())
		assert.Equal(t, "warehouse-nyc-01", modified.Metadata()["warehouseId"].Value())
		assert.Equal(t, trogonerror.VisibilityInternal, modified.Metadata()["warehouseId"].Visibility())
	})

	t.Run("WithChangeMetadataValuef on empty metadata", func(t *testing.T) {
		original := trogonerror.NewError("shopify.cache", "CACHE_MISS")

		customerID := "3333333333"
		modified := original.WithChanges(
			trogonerror.WithChangeMetadataValuef(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/%s", customerID))

		// Original should have no metadata
		assert.Empty(t, original.Metadata())

		// Modified should have formatted metadata value
		assert.Equal(t, "gid://shopify/Customer/3333333333", modified.Metadata()["customerId"].Value())
		assert.Equal(t, trogonerror.VisibilityPublic, modified.Metadata()["customerId"].Visibility())
	})
}

func TestTrogonErrorDebugInfo(t *testing.T) {
	t.Run("WithDebugDetail sets debug detail without stack trace", func(t *testing.T) {
		err := trogonerror.NewError("shopify.database", "CONNECTION_FAILED",
			trogonerror.WithDebugDetail("Connection pool exhausted - max connections: 100, active: 100"))

		assert.NotNil(t, err.DebugInfo())
		assert.Equal(t, "Connection pool exhausted - max connections: 100, active: 100", err.DebugInfo().Detail())
		assert.Empty(t, err.DebugInfo().StackEntries())
	})

	t.Run("WithStackTrace and WithDebugDetail can be combined", func(t *testing.T) {
		err := trogonerror.NewError("shopify.parser", "INVALID_JSON",
			trogonerror.WithStackTrace(),
			trogonerror.WithDebugDetail("Connection pool exhausted - max connections: 100, active: 100"))

		assert.NotNil(t, err.DebugInfo())
		assert.Equal(t, "Connection pool exhausted - max connections: 100, active: 100", err.DebugInfo().Detail())
		assert.NotEmpty(t, err.DebugInfo().StackEntries())
		foundTestFunc := false
		for _, entry := range err.DebugInfo().StackEntries() {
			if strings.Contains(entry, "TestTrogonErrorDebugInfo") {
				foundTestFunc = true
				break
			}
		}
		assert.True(t, foundTestFunc, "Stack trace should contain test function name")
	})

	t.Run("StackFrames returns raw runtime.Frame objects", func(t *testing.T) {
		err := trogonerror.NewError("shopify.renderer", "TEMPLATE_ERROR",
			trogonerror.WithStackTrace())

		assert.NotNil(t, err.DebugInfo())

		frames := err.DebugInfo().StackFrames()
		assert.NotEmpty(t, frames)

		for _, frame := range frames {
			assert.NotEmpty(t, frame.File)
			assert.NotEmpty(t, frame.Function)
			assert.Greater(t, frame.Line, 0)
		}

		entries := err.DebugInfo().StackEntries()
		assert.Equal(t, len(frames), len(entries))
	})

	t.Run("WithDebugInfo sets complete debug info", func(t *testing.T) {
		tempErr := trogonerror.NewError("shopify.analytics", "METRIC_CALCULATION_FAILED",
			trogonerror.WithStackTrace(),
			trogonerror.WithDebugDetail("Analytics calculation failed: division by zero in revenue computation"))

		err := trogonerror.NewError("shopify.profiler", "PROFILE_GENERATION_FAILED",
			trogonerror.WithDebugInfo(*tempErr.DebugInfo()))

		assert.NotNil(t, err.DebugInfo())
		assert.Equal(t, "Analytics calculation failed: division by zero in revenue computation", err.DebugInfo().Detail())
		assert.NotEmpty(t, err.DebugInfo().StackFrames())
	})
	t.Run("WithStackTrace captures stack trace without setting detail", func(t *testing.T) {
		err := trogonerror.NewError("shopify.parser", "SYNTAX_ERROR",
			trogonerror.WithCode(trogonerror.CodeInternal),
			trogonerror.WithStackTrace())

		assert.NotNil(t, err.DebugInfo())
		assert.Empty(t, err.DebugInfo().Detail())

		stackEntries := err.DebugInfo().StackEntries()
		assert.NotEmpty(t, stackEntries)

		found := false
		for _, entry := range stackEntries {
			if strings.Contains(entry, "TestTrogonErrorDebugInfo") {
				found = true
				break
			}
		}
		assert.True(t, found, "Stack trace should contain the calling function")
	})

	t.Run("WithStackTraceDepth limits stack trace depth without setting detail", func(t *testing.T) {
		err := trogonerror.NewError("shopify.parser", "MAX_DEPTH_EXCEEDED",
			trogonerror.WithStackTraceDepth(5)) // Only capture 5 frames

		assert.NotNil(t, err.DebugInfo())
		assert.Empty(t, err.DebugInfo().Detail())
		stackEntries := err.DebugInfo().StackEntries()
		assert.NotEmpty(t, stackEntries)
		assert.LessOrEqual(t, len(stackEntries), 5, "Stack should be limited to 5 frames")
	})

}

func TestTrogonErrorMutation(t *testing.T) {
	t.Run("WithChanges applies multiple changes", func(t *testing.T) {
		original := trogonerror.NewError("shopify.settings", "CONFIG_NOT_FOUND",
			trogonerror.WithCode(trogonerror.CodeUnknown),
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "configKey", "theme_customization_enabled"))

		modified := original.WithChanges(
			trogonerror.WithChangeID("new-id"),
			trogonerror.WithChangeSourceID("new-source"),
			trogonerror.WithChangeMetadataValue(trogonerror.VisibilityPrivate, "userId", "gid://shopify/Customer/9876543210"))

		assert.Empty(t, original.ID())
		assert.Empty(t, original.SourceID())
		assert.NotContains(t, original.Metadata(), "userId")

		assert.Equal(t, "new-id", modified.ID())
		assert.Equal(t, "new-source", modified.SourceID())
		assert.Equal(t, "gid://shopify/Customer/9876543210", modified.Metadata()["userId"].Value())
		assert.Equal(t, "theme_customization_enabled", modified.Metadata()["configKey"].Value()) // Original metadata preserved
	})

	t.Run("WithChangeMetadata replaces metadata", func(t *testing.T) {
		original := trogonerror.NewError("shopify.cache", "CACHE_MISS",
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "legacyCustomerId", "cust_legacy_123456"))

		newMetadata := make(map[string]trogonerror.MetadataValue)
		baseErr := trogonerror.NewError("shopify.settings", "CONFIG_UPDATED")
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "customerId", "gid://shopify/Customer/1234567890")(baseErr)
		for k, v := range baseErr.Metadata() {
			newMetadata[k] = v
		}

		modified := original.WithChanges(trogonerror.WithChangeMetadata(newMetadata))

		assert.Contains(t, original.Metadata(), "legacyCustomerId")
		assert.Contains(t, modified.Metadata(), "customerId")
		assert.NotContains(t, modified.Metadata(), "legacyCustomerId")
	})

	t.Run("WithChangeTime", func(t *testing.T) {
		original := trogonerror.NewError("shopify.scheduler", "SCHEDULE_CONFLICT")

		specificTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		withSpecificTime := original.WithChanges(trogonerror.WithChangeTime(specificTime))

		assert.Nil(t, original.Time())
		assert.True(t, withSpecificTime.Time().Equal(specificTime))
	})

	t.Run("WithChangeHelpLink adds help link", func(t *testing.T) {
		original := trogonerror.NewError("shopify.docs", "API_DOCS_UNAVAILABLE")

		modified := original.WithChanges(
			trogonerror.WithChangeHelpLink("API Status", "https://status.shopify.com"))

		assert.Nil(t, original.Help())
		assert.NotNil(t, modified.Help())
		assert.Len(t, modified.Help().Links(), 1)
		assert.Equal(t, "API Status", modified.Help().Links()[0].Description())
		assert.Equal(t, "https://status.shopify.com", modified.Help().Links()[0].URL())
	})

	t.Run("WithChangeHelpLinkf adds formatted help link", func(t *testing.T) {
		original := trogonerror.NewError("shopify.docs", "API_DOCS_UNAVAILABLE")
		userID := "gid://shopify/Customer/4567890123"

		modified := original.WithChanges(
			trogonerror.WithChangeHelpLinkf("Customer Console", "https://admin.shopify.com/customers/%s/help", userID))

		assert.Nil(t, original.Help())
		assert.NotNil(t, modified.Help())
		assert.Len(t, modified.Help().Links(), 1)
		assert.Equal(t, "Customer Console", modified.Help().Links()[0].Description())
		assert.Equal(t, "https://admin.shopify.com/customers/gid://shopify/Customer/4567890123/help", modified.Help().Links()[0].URL())
	})

	t.Run("WithChangeRetryInfoDuration and WithChangeRetryTime", func(t *testing.T) {
		original := trogonerror.NewError("shopify.queue", "QUEUE_FULL")

		withDuration := original.WithChanges(trogonerror.WithChangeRetryInfoDuration(30 * time.Second))
		retryTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		withTime := original.WithChanges(trogonerror.WithChangeRetryTime(retryTime))

		assert.Nil(t, original.RetryInfo())

		assert.NotNil(t, withDuration.RetryInfo())
		assert.Equal(t, 30*time.Second, *withDuration.RetryInfo().RetryOffset())
		assert.Nil(t, withDuration.RetryInfo().RetryTime())

		assert.NotNil(t, withTime.RetryInfo())
		assert.True(t, withTime.RetryInfo().RetryTime().Equal(retryTime))
		assert.Nil(t, withTime.RetryInfo().RetryOffset())
	})

	t.Run("WithChangeLocalizedMessage sets localized message", func(t *testing.T) {
		original := trogonerror.NewError("shopify.i18n", "TRANSLATION_MISSING")

		modified := original.WithChanges(
			trogonerror.WithChangeLocalizedMessage("es-ES", "Traducción no encontrada para esta región"))

		assert.Nil(t, original.LocalizedMessage())
		assert.NotNil(t, modified.LocalizedMessage())
		assert.Equal(t, "es-ES", modified.LocalizedMessage().Locale())
		assert.Equal(t, "Traducción no encontrada para esta región", modified.LocalizedMessage().Message())
	})

	t.Run("copy method creates independent copy", func(t *testing.T) {
		original := trogonerror.NewError("shopify.backup", "BACKUP_FAILED",
			trogonerror.WithCode(trogonerror.CodeUnknown),
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "configKey", "theme_customization_enabled"),
			trogonerror.WithStackTrace(),
			trogonerror.WithDebugDetail("S3 backup failed: access denied to bucket shopify-backups-prod"),
			trogonerror.WithHelpLink("Backup Documentation", "https://shopify.dev/docs/admin-api/rest/reference/events/backup"))

		copied := original.WithChanges(trogonerror.WithChangeID("new-id"))

		assert.NotEqual(t, fmt.Sprintf("%p", original), fmt.Sprintf("%p", copied))

		assert.Empty(t, original.ID())
		assert.Equal(t, "new-id", copied.ID())

		assert.Equal(t, original.Domain(), copied.Domain())
		assert.Equal(t, original.Reason(), copied.Reason())
		assert.Equal(t, original.Code(), copied.Code())
		assert.Equal(t, len(original.Metadata()), len(copied.Metadata()))
		assert.Equal(t, original.DebugInfo().Detail(), copied.DebugInfo().Detail())
		assert.Equal(t, len(original.DebugInfo().StackEntries()), len(copied.DebugInfo().StackEntries()))
		assert.Equal(t, len(original.Help().Links()), len(copied.Help().Links()))
	})

	t.Run("WithChangeMetadataValue on empty metadata", func(t *testing.T) {
		original := trogonerror.NewError("shopify.cache", "CACHE_MISS")

		modified := original.WithChanges(trogonerror.WithChangeMetadataValue(trogonerror.VisibilityPublic, "configKey", "theme_customization_enabled"))

		assert.Empty(t, original.Metadata())
		assert.Contains(t, modified.Metadata(), "configKey")
		assert.Equal(t, "theme_customization_enabled", modified.Metadata()["configKey"].Value())
	})

	t.Run("copy with empty collections", func(t *testing.T) {
		original := trogonerror.NewError("shopify.validation", "INVALID_INPUT")
		copied := original.WithChanges(trogonerror.WithChangeID("test"))

		assert.Empty(t, copied.Metadata())
		assert.Empty(t, copied.Causes())
		assert.Nil(t, copied.Help())
		assert.Nil(t, copied.DebugInfo())
	})

	t.Run("copy with causes", func(t *testing.T) {
		cause := trogonerror.NewError("shopify.auth", "TOKEN_EXPIRED")
		original := trogonerror.NewError("shopify.session", "SESSION_EXPIRED",
			trogonerror.WithCause(cause))
		copied := original.WithChanges(trogonerror.WithChangeID("test"))

		assert.Len(t, copied.Causes(), 1)
		assert.Equal(t, cause, copied.Causes()[0])
	})

	t.Run("copy with empty help links", func(t *testing.T) {
		help := trogonerror.Help{}
		original := trogonerror.NewError("shopify.docs", "API_DOCS_UNAVAILABLE",
			trogonerror.WithHelp(help))
		copied := original.WithChanges(trogonerror.WithChangeID("test"))

		assert.NotNil(t, copied.Help())
		assert.Empty(t, copied.Help().Links())
	})

	t.Run("copy debugInfo with empty stack frames", func(t *testing.T) {
		original := trogonerror.NewError("shopify.logger", "LOG_WRITE_FAILED",
			trogonerror.WithDebugDetail("Log rotation failed: insufficient disk space (97% full)"))
		copied := original.WithChanges(trogonerror.WithChangeID("test"))

		assert.NotNil(t, copied.DebugInfo())
		assert.Equal(t, "Log rotation failed: insufficient disk space (97% full)", copied.DebugInfo().Detail())
		assert.Nil(t, copied.DebugInfo().StackFrames())
	})
}

func TestTrogonErrorCauses(t *testing.T) {
	t.Run("WithCause chains multiple error causes", func(t *testing.T) {
		cause1 := trogonerror.NewError("shopify.database", "CONNECTION_TIMEOUT")
		cause2 := trogonerror.NewError("shopify.network", "NETWORK_UNAVAILABLE")

		err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
			trogonerror.WithCause(cause1),
			trogonerror.WithCause(cause2))

		assert.Len(t, err.Causes(), 2)
		assert.Equal(t, "shopify.database", err.Causes()[0].Domain())
		assert.Equal(t, "shopify.network", err.Causes()[1].Domain())
		assert.Contains(t, err.Causes(), cause1)
		assert.Contains(t, err.Causes(), cause2)
		assert.NotEmpty(t, err.Causes())
	})
}

func TestTrogonErrorInterfaces(t *testing.T) {
	t.Run("Error() formats error string correctly", func(t *testing.T) {
		err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
			trogonerror.WithMessage("Payment gateway rejected transaction due to insufficient funds"))

		expected1 := `Payment gateway rejected transaction due to insufficient funds
  visibility: INTERNAL
  domain: shopify.payments
  reason: PAYMENT_DECLINED
  code: UNKNOWN`

		assert.Equal(t, expected1, err.Error())

		err2 := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
			trogonerror.WithCode(trogonerror.CodeNotFound))

		expected2 := `resource not found
  visibility: INTERNAL
  domain: shopify.payments
  reason: PAYMENT_DECLINED
  code: NOT_FOUND`

		assert.Equal(t, expected2, err2.Error())
	})

	t.Run("errors.As type assertion works correctly", func(t *testing.T) {
		uErr := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")

		var tErr *trogonerror.TrogonError
		assert.True(t, errors.As(uErr, &tErr))
		assert.Equal(t, "shopify.session", tErr.Domain())

		regularErr := errors.New("regular error")
		assert.False(t, errors.As(regularErr, &tErr))
	})

	t.Run("errors.Is compares by domain and reason", func(t *testing.T) {
		uErr1 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		uErr2 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		uErr3 := trogonerror.NewError("shopify.session", "SESSION_INVALID")

		assert.True(t, errors.Is(uErr1, uErr2))
		assert.False(t, errors.Is(uErr1, uErr3))
		assert.True(t, errors.Is(uErr1, uErr1))

		nonTrogonErr := errors.New("standard error")
		assert.False(t, errors.Is(uErr1, nonTrogonErr))

		uErr4 := *uErr1
		assert.True(t, errors.Is(uErr1, uErr4))
	})
}

func TestCode(t *testing.T) {
	t.Run("Code methods return correct values", func(t *testing.T) {
		code := trogonerror.CodeNotFound

		assert.Equal(t, 404, code.HttpStatusCode())
		assert.Equal(t, "NOT_FOUND", code.String())
		assert.Equal(t, "resource not found", code.Message())
		assert.NotEmpty(t, code.String())
		assert.Equal(t, code.HttpStatusCode(), 404)
	})

	t.Run("All code messages are tested", func(t *testing.T) {
		tests := []struct {
			code    trogonerror.Code
			message string
		}{
			{trogonerror.CodeCancelled, "the operation was cancelled"},
			{trogonerror.CodeUnknown, "unknown error"},
			{trogonerror.CodeInvalidArgument, "invalid argument provided"},
			{trogonerror.CodeDeadlineExceeded, "deadline exceeded"},
			{trogonerror.CodeNotFound, "resource not found"},
			{trogonerror.CodeAlreadyExists, "resource already exists"},
			{trogonerror.CodePermissionDenied, "permission denied"},
			{trogonerror.CodeUnauthenticated, "unauthenticated"},
			{trogonerror.CodeResourceExhausted, "resource exhausted"},
			{trogonerror.CodeFailedPrecondition, "failed precondition"},
			{trogonerror.CodeAborted, "operation aborted"},
			{trogonerror.CodeOutOfRange, "out of range"},
			{trogonerror.CodeUnimplemented, "not implemented"},
			{trogonerror.CodeInternal, "internal error"},
			{trogonerror.CodeUnavailable, "service unavailable"},
			{trogonerror.CodeDataLoss, "data loss or corruption"},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.message, tt.code.Message())
		}
	})
}

func TestTrogonErrorTimeFeatures(t *testing.T) {
	t.Run("WithTime sets current time", func(t *testing.T) {
		now := time.Now()
		err := trogonerror.NewError("shopify.scheduler", "SCHEDULE_CONFLICT",
			trogonerror.WithTime(now))

		assert.NotNil(t, err.Time())
		assert.True(t, err.Time().Equal(now))
	})

	t.Run("WithRetryTime sets retry time", func(t *testing.T) {
		retryTime := time.Now().Add(5 * time.Minute)
		err := trogonerror.NewError("shopify.queue", "QUEUE_FULL",
			trogonerror.WithRetryTime(retryTime))

		assert.NotNil(t, err.RetryInfo())
		assert.NotNil(t, err.RetryInfo().RetryTime())
		assert.True(t, err.RetryInfo().RetryTime().Equal(retryTime))
		assert.Nil(t, err.RetryInfo().RetryOffset())
	})

	t.Run("WithRetryInfoDuration sets retry duration", func(t *testing.T) {
		retryDuration := 30 * time.Second
		err := trogonerror.NewError("shopify.queue", "QUEUE_FULL",
			trogonerror.WithRetryInfoDuration(retryDuration))

		assert.NotNil(t, err.RetryInfo())
		assert.NotNil(t, err.RetryInfo().RetryOffset())
		assert.Equal(t, retryDuration, *err.RetryInfo().RetryOffset())
		assert.Nil(t, err.RetryInfo().RetryTime())
	})
}

func TestHTTPCodesMatchADR(t *testing.T) {
	tests := []struct {
		code     trogonerror.Code
		httpCode int
		name     string
	}{
		{trogonerror.CodeCancelled, 499, "CANCELLED"},
		{trogonerror.CodeUnknown, 500, "UNKNOWN"},
		{trogonerror.CodeInvalidArgument, 400, "INVALID_ARGUMENT"},
		{trogonerror.CodeDeadlineExceeded, 504, "DEADLINE_EXCEEDED"},
		{trogonerror.CodeNotFound, 404, "NOT_FOUND"},
		{trogonerror.CodeAlreadyExists, 409, "ALREADY_EXISTS"},
		{trogonerror.CodePermissionDenied, 403, "PERMISSION_DENIED"},
		{trogonerror.CodeUnauthenticated, 401, "UNAUTHENTICATED"},
		{trogonerror.CodeResourceExhausted, 429, "RESOURCE_EXHAUSTED"},
		{trogonerror.CodeFailedPrecondition, 400, "FAILED_PRECONDITION"},
		{trogonerror.CodeAborted, 409, "ABORTED"},
		{trogonerror.CodeOutOfRange, 400, "OUT_OF_RANGE"},
		{trogonerror.CodeUnimplemented, 501, "UNIMPLEMENTED"},
		{trogonerror.CodeInternal, 500, "INTERNAL"},
		{trogonerror.CodeUnavailable, 503, "UNAVAILABLE"},
		{trogonerror.CodeDataLoss, 500, "DATA_LOSS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.httpCode, tt.code.HttpStatusCode())
			assert.Equal(t, tt.name, tt.code.String())
		})
	}

	t.Run("Unknown code returns 500", func(t *testing.T) {
		var unknownCode trogonerror.Code = 999
		assert.Equal(t, 500, unknownCode.HttpStatusCode())
	})
}

func TestTrogonErrorWrapping(t *testing.T) {
	t.Run("WithWrap standard error preserves wrapped error for errors.Is", func(t *testing.T) {
		originalErr := fmt.Errorf("PostgreSQL connection failed: timeout after 30s")

		err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
			trogonerror.WithCode(trogonerror.CodeInternal),
			trogonerror.WithWrap(originalErr),
			trogonerror.WithErrorMessage(originalErr))

		assert.Equal(t, "PostgreSQL connection failed: timeout after 30s", err.Message())
		assert.Len(t, err.Causes(), 0)
	})

	t.Run("WithWrap TrogonError preserves wrapped error without affecting message", func(t *testing.T) {
		dbErr := trogonerror.NewError("shopify.database", "CONNECTION_FAILED",
			trogonerror.WithCode(trogonerror.CodeUnavailable),
			trogonerror.WithMessage("PostgreSQL connection timeout after 30 seconds"))

		err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
			trogonerror.WithCode(trogonerror.CodeInternal),
			trogonerror.WithWrap(dbErr))

		assert.Equal(t, "internal error", err.Message())
		assert.Len(t, err.Causes(), 0)
	})

	t.Run("WithWrap combined with other options like metadata and subject", func(t *testing.T) {
		originalErr := fmt.Errorf("Email validation failed")

		err := trogonerror.NewError("shopify.users", "USER_CREATION_FAILED",
			trogonerror.WithCode(trogonerror.CodeInvalidArgument),
			trogonerror.WithWrap(originalErr),
			trogonerror.WithErrorMessage(originalErr),
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "field", "email"),
			trogonerror.WithSubject("/email"))

		assert.Equal(t, "Email validation failed", err.Message())
		assert.Equal(t, "email", err.Metadata()["field"].Value())
		assert.Equal(t, "/email", err.Subject())
	})

	customErr := CustomError{msg: "custom error"}
	stdErr := fmt.Errorf("wrapped: %w", customErr)

	t.Run("WithWrap preserves wrapped error chain for errors.Is", func(t *testing.T) {
		err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
			trogonerror.WithCode(trogonerror.CodeInternal),
			trogonerror.WithWrap(stdErr))

		// Should match the wrapped error chain
		assert.True(t, errors.Is(err, customErr))
		assert.True(t, errors.Is(err, stdErr))

		// Should not match unrelated errors
		otherErr := fmt.Errorf("different error")
		assert.False(t, errors.Is(err, otherErr))
	})

	t.Run("WithWrap chaining TrogonErrors preserves full error chain", func(t *testing.T) {
		dbErr := trogonerror.NewError("shopify.database", "CONNECTION_FAILED",
			trogonerror.WithCode(trogonerror.CodeUnavailable),
			trogonerror.WithWrap(customErr))

		serviceErr := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
			trogonerror.WithCode(trogonerror.CodeInternal),
			trogonerror.WithWrap(dbErr))

		assert.True(t, errors.Is(serviceErr, dbErr))

		assert.True(t, errors.Is(serviceErr, customErr))

		assert.Len(t, serviceErr.Causes(), 0)

		assert.True(t, errors.Is(dbErr, customErr))
	})

	t.Run("TrogonError Is comparison based on domain and reason", func(t *testing.T) {
		err1 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		err2 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		err3 := trogonerror.NewError("shopify.session", "SESSION_INVALID")

		assert.True(t, errors.Is(err1, err2))
		assert.False(t, errors.Is(err1, err3))
	})
}

func TestTrogonErrorEdgeCases(t *testing.T) {
	t.Run("StackFrames returns nil for empty stack", func(t *testing.T) {
		err := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")

		if err.DebugInfo() != nil {
			frames := err.DebugInfo().StackFrames()
			assert.Nil(t, frames)
		}
	})

	t.Run("StackFrames returns copy of frames", func(t *testing.T) {
		err := trogonerror.NewError("shopify.renderer", "TEMPLATE_ERROR",
			trogonerror.WithStackTrace())

		frames1 := err.DebugInfo().StackFrames()
		frames2 := err.DebugInfo().StackFrames()

		assert.NotSame(t, &frames1, &frames2)
		assert.Equal(t, len(frames1), len(frames2))
	})

	t.Run("StackEntries returns nil for empty stack", func(t *testing.T) {
		err := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")

		if err.DebugInfo() != nil {
			entries := err.DebugInfo().StackEntries()
			assert.Nil(t, entries)
		}
	})

	t.Run("WithStackTraceDepth with zero depth", func(t *testing.T) {
		err := trogonerror.NewError("shopify.parser", "INVALID_JSON",
			trogonerror.WithStackTraceDepth(0))

		assert.NotNil(t, err.DebugInfo())
		assert.NotEmpty(t, err.DebugInfo().StackEntries())
	})

	t.Run("WithStackTraceDepth with negative depth", func(t *testing.T) {
		err := trogonerror.NewError("shopify.parser", "INVALID_JSON",
			trogonerror.WithStackTraceDepth(-5))

		assert.NotNil(t, err.DebugInfo())
		assert.NotEmpty(t, err.DebugInfo().StackEntries())
	})

	t.Run("WithStackTraceDepth updates existing debugInfo", func(t *testing.T) {
		err := trogonerror.NewError("shopify.logger", "LOG_WRITE_FAILED",
			trogonerror.WithDebugDetail("Stack trace captured during profiling"),
			trogonerror.WithStackTraceDepth(5))

		assert.NotNil(t, err.DebugInfo())
		assert.Equal(t, "Stack trace captured during profiling", err.DebugInfo().Detail())
		assert.NotEmpty(t, err.DebugInfo().StackEntries())
	})

	t.Run("Code Message returns default for unknown code", func(t *testing.T) {
		var unknownCode trogonerror.Code = 999
		assert.Equal(t, "unknown error", unknownCode.Message())
	})

	t.Run("Code String returns default for unknown code", func(t *testing.T) {
		var unknownCode trogonerror.Code = 999
		assert.Equal(t, "UNKNOWN", unknownCode.String())
	})

	t.Run("Visibility String returns default for unknown visibility", func(t *testing.T) {
		var unknownVisibility trogonerror.Visibility = 999
		assert.Equal(t, "UNKNOWN", unknownVisibility.String())
	})
}

// CustomError is a test error type
type CustomError struct {
	msg string
}

func (c CustomError) Error() string {
	return c.msg
}

func TestAs(t *testing.T) {
	t.Run("As returns TrogonError when error matches", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.inventory", "INSUFFICIENT_INVENTORY",
			trogonerror.TemplateWithCode(trogonerror.CodeFailedPrecondition))

		originalErr := template.NewError(
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "productId", "gid://shopify/Product/1234567890"))

		trogonErr, ok := trogonerror.As(originalErr, template)
		assert.True(t, ok)
		assert.NotNil(t, trogonErr)
		assert.Equal(t, "shopify.inventory", trogonErr.Domain())
		assert.Equal(t, "INSUFFICIENT_INVENTORY", trogonErr.Reason())
		assert.Equal(t, "gid://shopify/Product/1234567890", trogonErr.Metadata()["productId"].Value())

	})

	t.Run("As returns false when error doesn't match", func(t *testing.T) {
		template1 := trogonerror.NewErrorTemplate("shopify.inventory", "INSUFFICIENT_INVENTORY")
		template2 := trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND")

		err1 := template1.NewError()

		trogonErr, ok := trogonerror.As(err1, template2)
		assert.False(t, ok)
		assert.Nil(t, trogonErr)
	})

	t.Run("As returns false for non-TrogonError", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.inventory", "INSUFFICIENT_INVENTORY")
		regularErr := errors.New("regular error")

		trogonErr, ok := trogonerror.As(regularErr, template)
		assert.False(t, ok)
		assert.Nil(t, trogonErr)
	})

	t.Run("As works with WithChanges pattern", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.inventory", "INSUFFICIENT_INVENTORY",
			trogonerror.TemplateWithCode(trogonerror.CodeResourceExhausted))

		originalErr := template.NewError(
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "productId", "gid://shopify/Product/1234567890"))

		trogonErr, ok := trogonerror.As(originalErr, template)
		assert.True(t, ok)
		assert.NotNil(t, trogonErr)

		modifiedErr := trogonErr.WithChanges(
			trogonerror.WithChangeMetadataValue(trogonerror.VisibilityPublic, "main_order_id", "order_123"),
			trogonerror.WithChangeMetadataValue(trogonerror.VisibilityPublic, "listing_count", "5"),
		)

		assert.Equal(t, "order_123", modifiedErr.Metadata()["main_order_id"].Value())
		assert.Equal(t, "5", modifiedErr.Metadata()["listing_count"].Value())
		assert.Equal(t, "gid://shopify/Product/1234567890", modifiedErr.Metadata()["productId"].Value()) // Original preserved
	})

	t.Run("As works with TrogonError directly", func(t *testing.T) {
		originalErr := trogonerror.NewError("shopify.inventory", "INSUFFICIENT_INVENTORY",
			trogonerror.WithCode(trogonerror.CodeFailedPrecondition),
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "productId", "gid://shopify/Product/1234567890"))

		// Test with TrogonError as target (not template)
		trogonErr, ok := trogonerror.As(originalErr, originalErr)
		assert.True(t, ok)
		assert.NotNil(t, trogonErr)
		assert.Equal(t, "shopify.inventory", trogonErr.Domain())
		assert.Equal(t, "INSUFFICIENT_INVENTORY", trogonErr.Reason())
	})

	t.Run("As works with wrapped TrogonError", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.inventory", "INSUFFICIENT_INVENTORY")

		originalErr := template.NewError(
			trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "productId", "gid://shopify/Product/1234567890"))

		// Wrap the TrogonError with fmt.Errorf (this was the problematic scenario)
		wrappedErr := fmt.Errorf("context: %w", originalErr)

		// Test the As function with the wrapped error
		trogonErr, ok := trogonerror.As(wrappedErr, template)
		assert.True(t, ok, "As should return true for wrapped TrogonError")
		assert.NotNil(t, trogonErr, "As should return non-nil TrogonError")

		// Verify the extracted error has the correct properties
		assert.Equal(t, "shopify.inventory", trogonErr.Domain())
		assert.Equal(t, "INSUFFICIENT_INVENTORY", trogonErr.Reason())
		assert.Equal(t, "gid://shopify/Product/1234567890", trogonErr.Metadata()["productId"].Value())
	})

}

func TestInternalMethods(t *testing.T) {
	t.Run("TrogonError.is method delegates to Is", func(t *testing.T) {
		err1 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		err2 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		err3 := trogonerror.NewError("shopify.session", "SESSION_INVALID")

		// Test the internal is method by using it in the As function
		// This indirectly tests the is method since As calls target.is(err)
		trogonErr, ok := trogonerror.As(err2, err1)
		assert.True(t, ok)
		assert.NotNil(t, trogonErr)

		trogonErr2, ok2 := trogonerror.As(err3, err1)
		assert.False(t, ok2)
		assert.Nil(t, trogonErr2)
	})

	t.Run("ErrorTemplate.is method delegates to Is", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.session", "SESSION_EXPIRED")
		err1 := template.NewError()
		err2 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		err3 := trogonerror.NewError("shopify.session", "SESSION_INVALID")

		// Test the internal is method by using it in the As function
		// This indirectly tests the is method since As calls target.is(err)
		trogonErr, ok := trogonerror.As(err1, template)
		assert.True(t, ok)
		assert.NotNil(t, trogonErr)

		trogonErr2, ok2 := trogonerror.As(err2, template)
		assert.True(t, ok2)
		assert.NotNil(t, trogonErr2)

		trogonErr3, ok3 := trogonerror.As(err3, template)
		assert.False(t, ok3)
		assert.Nil(t, trogonErr3)
	})
}

func TestErrorTemplate(t *testing.T) {
	t.Run("NewErrorTemplate creates template with defaults", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.session", "SESSION_EXPIRED")
		err := template.NewError()

		assert.Equal(t, "shopify.session", err.Domain())
		assert.Equal(t, "SESSION_EXPIRED", err.Reason())
		assert.Equal(t, trogonerror.CodeUnknown, err.Code())
		assert.Equal(t, trogonerror.VisibilityInternal, err.Visibility())
	})

	t.Run("TemplateWithHelp sets help information", func(t *testing.T) {
		help := trogonerror.Help{}
		template := trogonerror.NewErrorTemplate("shopify.docs", "API_DOCS_UNAVAILABLE",
			trogonerror.TemplateWithHelp(help))
		err := template.NewError()

		assert.NotNil(t, err.Help())
	})

	t.Run("Template Is method compares domain and reason", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.session", "SESSION_EXPIRED")
		err1 := template.NewError()
		err2 := trogonerror.NewError("shopify.session", "SESSION_EXPIRED")
		err3 := trogonerror.NewError("shopify.session", "SESSION_INVALID")

		assert.True(t, template.Is(err1))
		assert.True(t, template.Is(err2))
		assert.False(t, template.Is(err3))

		err4 := *err2
		assert.True(t, template.Is(err4))

		regularErr := errors.New("regular error")
		assert.False(t, template.Is(regularErr))
	})

	t.Run("Template with all options", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.templates", "TEMPLATE_RENDERING_FAILED",
			trogonerror.TemplateWithCode(trogonerror.CodeInvalidArgument),
			trogonerror.TemplateWithMessage("Custom template message"),
			trogonerror.TemplateWithVisibility(trogonerror.VisibilityPublic),
			trogonerror.TemplateWithHelpLink("Docs", "https://example.com"))

		err := template.NewError(trogonerror.WithID("err_2024_01_15_template_rendering_abc123"))

		assert.Equal(t, trogonerror.CodeInvalidArgument, err.Code())
		assert.Equal(t, "Custom template message", err.Message())
		assert.Equal(t, trogonerror.VisibilityPublic, err.Visibility())
		assert.Equal(t, "err_2024_01_15_template_rendering_abc123", err.ID())
		assert.NotNil(t, err.Help())
		assert.Len(t, err.Help().Links(), 1)
	})
}
