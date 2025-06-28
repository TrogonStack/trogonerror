// Package trogonerror provides a comprehensive, standardized error handling system
// for distributed applications based on the TrogonError specification.
//
// TrogonError offers structured error handling with standardized error codes,
// rich metadata, visibility controls, internationalization support, and
// cause chaining for better debugging and error correlation.
//
// # Production Usage with Error Templates (Recommended)
//
// For production applications, use error templates to define reusable error patterns.
// Templates ensure consistency across your application and reduce code duplication:
//
//	var (
//		ErrUserNotFound = trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND",
//			trogonerror.TemplateWithCode(trogonerror.NotFound))
//
//		ErrValidationFailed = trogonerror.NewErrorTemplate("shopify.validation", "INVALID_INPUT",
//			trogonerror.TemplateWithCode(trogonerror.InvalidArgument))
//
//		ErrDatabaseError = trogonerror.NewErrorTemplate("shopify.database", "CONNECTION_FAILED",
//			trogonerror.TemplateWithCode(trogonerror.Internal))
//	)
//
//	func GetUser(id string) (*User, error) {
//		user, err := db.FindUser(id)
//		if err != nil {
//			if errors.Is(err, sql.ErrNoRows) {
//				return nil, ErrUserNotFound.NewError(
//					trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/%s", id))
//			}
//			return nil, ErrDatabaseError.NewError(trogonerror.WithWrap(err))
//		}
//		return user, nil
//	}
//
// # Basic Error Creation
//
// Create errors using functional options:
//
//	err := trogonerror.NewError("shopify.users", "NOT_FOUND",
//		trogonerror.WithCode(trogonerror.NotFound),
//		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))
//
// # Error Codes and HTTP Status Mapping
//
// TrogonError supports 16 standardized error codes that map to HTTP status codes:
//
//	Code                    HTTP Status    Description
//	----                    -----------    -----------
//	Cancelled               499           Request cancelled
//	Unknown                 500           Unknown error
//	InvalidArgument         400           Invalid request parameters
//	DeadlineExceeded        504           Request timeout
//	NotFound                404           Resource not found
//	AlreadyExists           409           Resource already exists
//	PermissionDenied        403           Access denied
//	Unauthenticated         401           Authentication required
//	ResourceExhausted       429           Rate limit exceeded
//	FailedPrecondition      422           Precondition failed
//	Aborted                 409           Operation aborted
//	OutOfRange              400           Value out of range
//	Unimplemented           501           Not implemented
//	Internal                500           Internal server error
//	Unavailable             503           Service unavailable
//	DataLoss                500           Data corruption
//
// Access code information:
//
//	code := trogonerror.NotFound
//	fmt.Println(code.String())         // "NOT_FOUND"
//	fmt.Println(code.HttpStatusCode()) // 404
//	fmt.Println(code.Message())        // "resource not found"
//
// # Visibility Control System
//
// TrogonError implements a three-tier visibility system to control information disclosure:
//
//	VisibilityInternal: Only visible within the same service/process
//	VisibilityPrivate:  Visible across internal services (not to external users)
//	VisibilityPublic:   Safe to expose to external users
//
//	err := trogonerror.NewError("shopify.auth", "ACCESS_DENIED",
//		trogonerror.WithCode(trogonerror.PermissionDenied),
//		trogonerror.WithVisibility(trogonerror.VisibilityPublic),
//		trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "userId", "gid://shopify/Customer/1234567890"),
//		trogonerror.WithMetadataValue(trogonerror.VisibilityPrivate, "resource", "/admin/customers"),
//		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "action", "DELETE"))
//
// # Rich Metadata with Formatting
//
// Add structured metadata with visibility controls and printf-style formatting:
//
//	err := trogonerror.NewError("shopify.orders", "ORDER_PROCESSING_FAILED",
//		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/5432109876"),
//		trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/%s", userID),
//		trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "amount", "$%.2f", float64(amount)/100))
//
// # Error Chaining and Wrapping
//
// Chain errors to preserve context while wrapping standard Go errors:
//
//	dbErr := trogonerror.NewError("shopify.database", "CONNECTION_FAILED",
//		trogonerror.WithCode(trogonerror.Internal))
//
//	serviceErr := trogonerror.NewError("shopify.users", "USER_FETCH_FAILED",
//		trogonerror.WithCode(trogonerror.Internal),
//		trogonerror.WithCause(dbErr))
//
//	// Wrap standard Go errors
//	originalErr := fmt.Errorf("connection timeout")
//	wrappedErr := trogonerror.NewError("shopify.database", "CONNECTION_TIMEOUT",
//		trogonerror.WithCode(trogonerror.Unavailable),
//		trogonerror.WithWrap(originalErr),
//		trogonerror.WithErrorMessage(originalErr))
//
// # Debugging Support
//
// Capture stack traces and debug information for internal debugging:
//
//	err := trogonerror.NewError("shopify.database", "QUERY_TIMEOUT",
//		trogonerror.WithCode(trogonerror.Internal),
//		trogonerror.WithStackTrace(),
//		trogonerror.WithDebugDetail("Query execution exceeded 30s timeout"),
//		trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "query", "SELECT * FROM users WHERE id = $1"))
//
//	// Access debug information
//	if debugInfo := err.DebugInfo(); debugInfo != nil {
//		fmt.Println("Detail:", debugInfo.Detail())
//		fmt.Println("Stack entries:", len(debugInfo.StackEntries()))
//	}
//
// # Retry Guidance
//
// Provide retry guidance with duration offset or absolute time:
//
//	// Retry after a duration
//	err := trogonerror.NewError("shopify.api", "RATE_LIMIT_EXCEEDED",
//		trogonerror.WithCode(trogonerror.ResourceExhausted),
//		trogonerror.WithRetryInfoDuration(60*time.Second),
//		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "limit", "1000"))
//
//	// Retry at specific time
//	retryTime := time.Now().Add(5 * time.Minute)
//	err := trogonerror.NewError("shopify.maintenance", "SERVICE_UNAVAILABLE",
//		trogonerror.WithCode(trogonerror.Unavailable),
//		trogonerror.WithRetryTime(retryTime))
//
// # Help Links and User Guidance
//
// Provide actionable help links with formatting support:
//
//	err := trogonerror.NewError("shopify.users", "INVALID_EMAIL",
//		trogonerror.WithCode(trogonerror.InvalidArgument),
//		trogonerror.WithHelpLink("Fix Email", "https://admin.shopify.com/customers/1234567890/edit#email"),
//		trogonerror.WithHelpLinkf("Customer Console", "https://admin.shopify.com/customers/%s/help", userID))
//
// # Internationalization
//
// Support localized error messages:
//
//	err := trogonerror.NewError("shopify.users", "NOT_FOUND",
//		trogonerror.WithCode(trogonerror.NotFound),
//		trogonerror.WithLocalizedMessage("es-ES", "Usuario no encontrado"))
//
//	fmt.Println(err.Message())                    // "resource not found" (default)
//	fmt.Println(err.LocalizedMessage().Message()) // "Usuario no encontrado"
//
// # Error Mutation with WithChanges
//
// Create modified copies of errors efficiently:
//
//	original := trogonerror.NewError("shopify.orders", "ORDER_FAILED",
//		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/12345"))
//
//	modified := original.WithChanges(
//		trogonerror.WithChangeID("error-123"),
//		trogonerror.WithChangeSourceID("payment-service"),
//		trogonerror.WithChangeMetadataValuef(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/%s", userID))
//
// # Standard Go Error Compatibility
//
// TrogonError implements the standard Go error interface and works with
// errors.Is, errors.As, and error wrapping:
//
//	// Type assertion
//	if tErr, ok := err.(*trogonerror.TrogonError); ok {
//		fmt.Printf("Domain: %s, Reason: %s\n", tErr.Domain(), tErr.Reason())
//	}
//
//	// errors.Is comparison (by domain + reason)
//	if errors.Is(err, ErrUserNotFound.NewError()) {
//		fmt.Println("User not found")
//	}
//
//	// errors.As type assertion
//	var tErr *trogonerror.TrogonError
//	if errors.As(err, &tErr) {
//		fmt.Printf("Code: %s, HTTP Status: %d\n", tErr.Code().String(), tErr.Code().HttpStatusCode())
//	}
//
// # Template-Based Error Checking
//
// Use template.Is() for efficient error type checking:
//
//	if ErrUserNotFound.Is(err) {
//		fmt.Println("This is a user not found error")
//	}
//
// # Complete Example
//
// Here's a complete example showing production usage:
//
//	var ErrPaymentFailed = trogonerror.NewErrorTemplate("shopify.payments", "PAYMENT_DECLINED",
//		trogonerror.TemplateWithCode(trogonerror.Internal))
//
//	func ProcessPayment(orderID, userID string, amount int) error {
//		// Simulate payment processing
//		if amount > 100000 { // Over $1000
//			return ErrPaymentFailed.NewError(
//				trogonerror.WithMessage("Payment amount exceeds limit"),
//				trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/%s", orderID),
//				trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "customerId", "gid://shopify/Customer/%s", userID),
//				trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "amount", "$%.2f", float64(amount)/100),
//				trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "currency", "USD"),
//				trogonerror.WithHelpLinkf("Contact Support", "https://admin.shopify.com/support/new?order_id=%s", orderID),
//				trogonerror.WithRetryInfoDuration(30*time.Minute))
//		}
//		return nil
//	}
package trogonerror
