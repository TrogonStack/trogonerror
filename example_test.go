package trogonerror_test

import (
	"fmt"
	"time"

	"github.com/TrogonStack/trogonerror"
)

func ExampleNewError() {
	err := trogonerror.NewError("shopify.users", "NOT_FOUND",
		trogonerror.WithCode(trogonerror.NotFound),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))

	fmt.Println(err.Error())
	fmt.Println(err.Domain())
	fmt.Println(err.Reason())
	// Output:
	// resource not found
	//   visibility: INTERNAL
	//   domain: shopify.users
	//   reason: NOT_FOUND
	//   code: NOT_FOUND
	//   metadata:
	//     - userId: gid://shopify/Customer/1234567890 visibility=PUBLIC
	//
	// shopify.users
	// NOT_FOUND
}

func ExampleErrorTemplate() {
	template := trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND",
		trogonerror.TemplateWithCode(trogonerror.NotFound))

	err := template.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))

	fmt.Println(err.Error())
	fmt.Println(err.Domain())
	fmt.Println(err.Reason())
	// Output:
	// resource not found
	//   visibility: INTERNAL
	//   domain: shopify.users
	//   reason: NOT_FOUND
	//   code: NOT_FOUND
	//   metadata:
	//     - userId: gid://shopify/Customer/1234567890 visibility=PUBLIC
	//
	// shopify.users
	// NOT_FOUND
}

func ExampleWithCause() {
	dbErr := trogonerror.NewError("shopify.database", "CONNECTION_FAILED",
		trogonerror.WithCode(trogonerror.Internal),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "host", "postgres-primary.shopify.com"))

	serviceErr := trogonerror.NewError("shopify.users", "USER_FETCH_FAILED",
		trogonerror.WithCode(trogonerror.Internal),
		trogonerror.WithMessage("Failed to fetch user data"),
		trogonerror.WithCause(dbErr))

	fmt.Println(serviceErr.Error())
	fmt.Println(len(serviceErr.Causes()))
	fmt.Println(serviceErr.Causes()[0].Domain())
	// Output:
	// Failed to fetch user data
	//   visibility: INTERNAL
	//   domain: shopify.users
	//   reason: USER_FETCH_FAILED
	//   code: INTERNAL
	// 1
	// shopify.database
}

func ExampleCode_methods() {
	code := trogonerror.NotFound

	fmt.Println(code.String())
	fmt.Println(code.HttpStatusCode())
	// Output:
	// NOT_FOUND
	// 404
}

func ExampleNewError_basic() {
	err := trogonerror.NewError("shopify.users", "NOT_FOUND",
		trogonerror.WithCode(trogonerror.NotFound),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))

	fmt.Println("Error message:", err.Error())
	fmt.Println("Domain:", err.Domain())
	fmt.Println("Reason:", err.Reason())
	fmt.Println("Code:", err.Code().String())

	// Output:
	// Error message: resource not found
	//   visibility: INTERNAL
	//   domain: shopify.users
	//   reason: NOT_FOUND
	//   code: NOT_FOUND
	//   metadata:
	//     - userId: gid://shopify/Customer/1234567890 visibility=PUBLIC
	//
	// Domain: shopify.users
	// Reason: NOT_FOUND
	// Code: NOT_FOUND
}

func ExampleErrorTemplate_production() {
	var (
		ErrUserNotFound = trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND",
			trogonerror.TemplateWithCode(trogonerror.NotFound),
			trogonerror.TemplateWithCode(trogonerror.NotFound))

		ErrInvalidInput = trogonerror.NewErrorTemplate("shopify.validation", "INVALID_INPUT",
			trogonerror.TemplateWithCode(trogonerror.InvalidArgument))
	)

	userErr := ErrUserNotFound.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/4567890123"))

	inputErr := ErrInvalidInput.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "fieldName", "email"),
		trogonerror.WithSubject("/email"))

	fmt.Println("User error:", userErr.Domain(), userErr.Reason())
	fmt.Println("Input error:", inputErr.Domain(), inputErr.Reason())

	// Output:
	// User error: shopify.users NOT_FOUND
	// Input error: shopify.validation INVALID_INPUT
}

func ExampleNewError_richMetadata() {
	err := trogonerror.NewError("shopify.payments", "PAYMENT_DECLINED",
		trogonerror.WithCode(trogonerror.Internal),
		trogonerror.WithMessage("Payment processing failed due to upstream service error"),
		trogonerror.WithVisibility(trogonerror.VisibilityPrivate),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "paymentId", "pay_2024_01_15_abc123def456"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "amount", "299.99"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "currency", "USD"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "merchantId", "gid://shopify/Shop/1234567890"),
		trogonerror.WithSubject("/payment/amount"),
		trogonerror.WithTime(time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)),
		trogonerror.WithSourceID("payment-gateway-prod-01"))

	fmt.Printf("Payment ID: %s\n", err.Metadata()["paymentId"].Value())
	fmt.Printf("Currency: %s\n", err.Metadata()["currency"].Value())
	fmt.Printf("Subject: %s\n", err.Subject())

	// Output:
	// Payment ID: pay_2024_01_15_abc123def456
	// Currency: USD
	// Subject: /payment/amount
}

func ExampleWithStackTrace_debugging() {
	// Error with stack trace for debugging
	err := trogonerror.NewError("shopify.database", "QUERY_TIMEOUT",
		trogonerror.WithCode(trogonerror.Internal),
		trogonerror.WithStackTrace(),
		trogonerror.WithDebugDetail("Database query failed with timeout"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "query", "SELECT * FROM customers WHERE id = $1"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "requestId", "req_2024_01_15_db_query_abc123"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "duration", "1.5s"))

	fmt.Printf("Has debug info: %v\n", err.DebugInfo() != nil)
	if err.DebugInfo() != nil {
		fmt.Printf("Debug detail: %s\n", err.DebugInfo().Detail())
		fmt.Printf("Stack entries count: %d\n", len(err.DebugInfo().StackEntries()))
	}

	// Output:
	// Has debug info: true
	// Debug detail: Database query failed with timeout
	// Stack entries count: 9
}

func ExampleWithCause_errorChaining() {
	dbErr := trogonerror.NewError("shopify.database", "CONNECTION_FAILED",
		trogonerror.WithCode(trogonerror.Unavailable),
		trogonerror.WithMessage("Database connection timeout"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "host", "postgres-primary.shopify.com"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "port", "5432"))

	serviceErr := trogonerror.NewError("shopify.users", "USER_FETCH_FAILED",
		trogonerror.WithCode(trogonerror.Internal),
		trogonerror.WithMessage("Failed to fetch user data"),
		trogonerror.WithCause(dbErr))

	fmt.Printf("Service error: %s\n", serviceErr.Reason())
	fmt.Printf("Has causes: %v\n", len(serviceErr.Causes()) > 0)
	if len(serviceErr.Causes()) > 0 {
		fmt.Printf("Root cause: %s\n", serviceErr.Causes()[0].Reason())
		fmt.Printf("Root cause domain: %s\n", serviceErr.Causes()[0].Domain())
	}

	// Output:
	// Service error: USER_FETCH_FAILED
	// Has causes: true
	// Root cause: CONNECTION_FAILED
	// Root cause domain: shopify.database
}

func ExampleWithRetryInfoDuration_retryLogic() {
	// Error with retry information for rate limiting
	err := trogonerror.NewError("shopify.api", "RATE_LIMIT_EXCEEDED",
		trogonerror.WithCode(trogonerror.ResourceExhausted),
		trogonerror.WithMessage("API rate limit exceeded"),
		trogonerror.WithRetryInfoDuration(60*time.Second),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "limit", "100"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "timeWindow", "1m"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "remaining", "0"))

	if retryInfo := err.RetryInfo(); retryInfo != nil {
		if retryOffset := retryInfo.RetryOffset(); retryOffset != nil {
			fmt.Printf("Retry after: %s\n", retryOffset.String())
		}
	}
	fmt.Printf("Rate limit: %s\n", err.Metadata()["limit"].Value())
	fmt.Printf("Window: %s\n", err.Metadata()["timeWindow"].Value())

	// Output:
	// Retry after: 1m0s
	// Rate limit: 100
	// Window: 1m
}

func ExampleWithRetryTime_absoluteRetry() {
	retryTime := time.Date(2024, 1, 15, 14, 35, 0, 0, time.UTC)

	err := trogonerror.NewError("shopify.maintenance", "SERVICE_UNAVAILABLE",
		trogonerror.WithCode(trogonerror.Unavailable),
		trogonerror.WithMessage("Service temporarily unavailable for maintenance"),
		trogonerror.WithRetryTime(retryTime),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "maintenanceWindow", "30min"))

	if retryInfo := err.RetryInfo(); retryInfo != nil {
		if retryTime := retryInfo.RetryTime(); retryTime != nil {
			fmt.Printf("Retry at: %s\n", retryTime.Format("2006-01-02 15:04:05 UTC"))
		}
	}
	fmt.Printf("Maintenance window: %s\n", err.Metadata()["maintenanceWindow"].Value())

	// Output:
	// Retry at: 2024-01-15 14:35:00 UTC
	// Maintenance window: 30min
}

func ExampleWithHelpLink_documentation() {
	// Error with help links for user guidance
	err := trogonerror.NewError("shopify.users", "INVALID_EMAIL",
		trogonerror.WithCode(trogonerror.InvalidArgument),
		trogonerror.WithMessage("Email address format is invalid"),
		trogonerror.WithSubject("/email"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "fieldName", "email"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "providedValue", "invalid-email"),
		trogonerror.WithHelpLink("Fix Email", "https://admin.shopify.com/customers/1234567890/edit#email"),
		trogonerror.WithHelpLink("Contact Support", "https://admin.shopify.com/support/new?customer_id=1234567890"))

	if help := err.Help(); help != nil {
		fmt.Printf("Help links available: %d\n", len(help.Links()))
		for _, link := range help.Links() {
			fmt.Printf("- %s: %s\n", link.Description(), link.URL())
		}
	}

	// Output:
	// Help links available: 2
	// - Fix Email: https://admin.shopify.com/customers/1234567890/edit#email
	// - Contact Support: https://admin.shopify.com/support/new?customer_id=1234567890
}

func ExampleWithLocalizedMessage_internationalization() {
	err := trogonerror.NewError("shopify.users", "NOT_FOUND",
		trogonerror.WithCode(trogonerror.NotFound),
		trogonerror.WithLocalizedMessage("es-ES", "Usuario no encontrado"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/7890123456"))

	fmt.Printf("Default message: %s\n", err.Message())
	if localizedMsg := err.LocalizedMessage(); localizedMsg != nil {
		fmt.Printf("Localized (%s): %s\n", localizedMsg.Locale(), localizedMsg.Message())
	}

	// Output:
	// Default message: resource not found
	// Localized (es-ES): Usuario no encontrado
}

func ExampleCode_utilities() {
	// Working with error codes
	code := trogonerror.NotFound

	fmt.Printf("Code string: %s\n", code.String())
	fmt.Printf("HTTP status: %d\n", code.HttpStatusCode())
	fmt.Printf("Default message: %s\n", code.Message())

	err := trogonerror.NewError("shopify.core", "INVALID_REQUEST", trogonerror.WithCode(trogonerror.InvalidArgument))
	if err.Code() == trogonerror.InvalidArgument {
		fmt.Println("This is an invalid argument error")
	}

	// Output:
	// Code string: NOT_FOUND
	// HTTP status: 404
	// Default message: resource not found
	// This is an invalid argument error
}

func ExampleWithWrap_standardErrors() {
	originalErr := fmt.Errorf("connection timeout after 30s")

	wrappedErr := trogonerror.NewError("shopify.database", "CONNECTION_TIMEOUT",
		trogonerror.WithCode(trogonerror.Unavailable),
		trogonerror.WithWrap(originalErr),
		trogonerror.WithErrorMessage(originalErr),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "timeout", "30s"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "host", "postgres-primary.shopify.com"))

	fmt.Printf("Wrapped error domain: %s\n", wrappedErr.Domain())
	fmt.Printf("Wrapped error reason: %s\n", wrappedErr.Reason())
	fmt.Printf("Original message preserved: %v\n", wrappedErr.Message() == originalErr.Error())

	// Output:
	// Wrapped error domain: shopify.database
	// Wrapped error reason: CONNECTION_TIMEOUT
	// Original message preserved: true
}

func ExampleTrogonError_visibilityControl() {
	// Demonstrate visibility controls
	err := trogonerror.NewError("shopify.auth", "ACCESS_DENIED",
		trogonerror.WithCode(trogonerror.PermissionDenied),
		trogonerror.WithVisibility(trogonerror.VisibilityPublic),
		trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "userId", "gid://shopify/Customer/1234567890"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "resource", "/admin/customers"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "action", "DELETE"))

	fmt.Printf("Error visibility: %s\n", err.Visibility().String())

	// Show metadata with different visibility levels in alphabetical order
	metadataKeys := []string{"action", "resource", "userId"}
	for _, key := range metadataKeys {
		if value, exists := err.Metadata()[key]; exists {
			fmt.Printf("Metadata %s (%s): %s\n", key, value.Visibility().String(), value.Value())
		}
	}

	// Output:
	// Error visibility: PUBLIC
	// Metadata action (PUBLIC): DELETE
	// Metadata resource (PRIVATE): /admin/customers
	// Metadata userId (INTERNAL): gid://shopify/Customer/1234567890
}

func ExampleWithMetadataValuef_formattedValues() {
	// Using printf-style formatting for metadata values
	userID := "1234567890"
	orderID := "5432109876"
	amount := 29999 // cents
	
	err := trogonerror.NewError("shopify.orders", "ORDER_PROCESSING_FAILED",
		trogonerror.WithCode(trogonerror.Internal),
		trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/%s", userID),
		trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/%s", orderID),
		trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "amount", "$%.2f", float64(amount)/100),
		trogonerror.WithMetadataValuef(trogonerror.VisibilityInternal, "requestId", "req_%s_%s", userID, orderID))

	fmt.Printf("Customer ID: %s\n", err.Metadata()["customerId"].Value())
	fmt.Printf("Order ID: %s\n", err.Metadata()["orderId"].Value())
	fmt.Printf("Amount: %s\n", err.Metadata()["amount"].Value())
	fmt.Printf("Request ID: %s\n", err.Metadata()["requestId"].Value())

	// Output:
	// Customer ID: gid://shopify/Customer/1234567890
	// Order ID: gid://shopify/Order/5432109876
	// Amount: $299.99
	// Request ID: req_1234567890_5432109876
}
