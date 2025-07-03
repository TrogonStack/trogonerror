# TrogonError

**TrogonError is a Go library that brings structured, secure error handling to distributed systems based on the [Straw Hat Error Specification](https://straw-hat-team.github.io/adr/adrs/0129349218/README.html).**

**TrogonError provides standardized error codes, rich metadata, and visibility controls to create consistent error handling across service boundaries.** It goes beyond simple error messages to include contextual information, retry guidance, and internationalization support while implementing a three-tier visibility system to control error disclosure.

**TrogonError addresses the critical challenge of inconsistent error handling in distributed systems, where different services often use incompatible error formats.** It prevents information leakage through visibility controls, enables effective debugging with error chaining and metadata, and improves user experience with localized messages and help links.

**TrogonError is particularly valuable for teams building microservices, REST APIs, or gRPC services that require consistent error responses.** Enterprise development teams benefit from its secure error handling with controlled information disclosure, while DevOps and SRE teams leverage its rich metadata for debugging distributed systems.

TrogonError provides structured error handling with:

- **Standardized error codes** mapped to common failure scenarios
- **Rich metadata** for debugging and error correlation
- **Three-tier visibility system** (Internal/Private/Public) for secure information disclosure
- **Error chaining** and stack traces for comprehensive debugging
- **Internationalization support** for user-facing error messages
- **Retry guidance** with duration or absolute time specifications
- **Help links** for enhanced user experience
- **Template system** for consistent error definitions across services

## Getting Started

### Installation

```bash
go get github.com/TrogonStack/trogonerror
```

### Production Templates (Recommended)

For production applications, use error templates to ensure consistency and maintainability. You may define
platform-specific error templates in a separate package and import them into your application. As well as define your
own error templates for your application.

```go
import (
    "context"
    "github.com/TrogonStack/trogonerror"
)

// Define reusable error templates
var (
    ErrUserNotFound = trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND",
        trogonerror.TemplateWithCode(trogonerror.NotFound))

    ErrValidationFailed = trogonerror.NewErrorTemplate("shopify", "VALIDATION_FAILED",
        trogonerror.TemplateWithCode(trogonerror.InvalidArgument))

    ErrOrderNotCancellable = trogonerror.NewErrorTemplate("shopify.orders", "ORDER_NOT_CANCELLABLE",
        trogonerror.TemplateWithCode(trogonerror.FailedPrecondition))

    ErrDatabaseError = trogonerror.NewErrorTemplate("shopify", "DATABASE_ERROR",
        trogonerror.TemplateWithCode(trogonerror.Internal))
)

type GetUser struct {
    UserID string
}
func GetUserHandler(ctx context.Context, cmd GetUser) (*User, error) {
    user, err := db.FindUser(ctx, "SELECT * FROM users WHERE id = $1", cmd.UserID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // This is a common pattern for not found errors
            return nil, ErrUserNotFound.NewError(
                trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/%s", cmd.UserID))
        }

        // This is a common pattern for database errors
        return nil, ErrDatabaseError.NewError(
            trogonerror.WithWrap(err))
    }

    // ... rest of the logic
}

type CreateUser struct {
    Email    string
}
func CreateUserHandler(ctx context.Context, cmd CreateUser) (*User, error) {
    if cmd.Email == "" {
        // This is a common pattern for validation errors
        return nil, ErrValidationFailed.NewError(
            trogonerror.WithSubject("/email"),
            trogonerror.WithMessage("email is required"),
            trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "validationType", "REQUIRED"))
    }

    // ... rest of the logic
}


type CancelOrder struct {
    OrderID    string
}
func CancelOrderHandler(ctx context.Context, cmd CancelOrder) error {
    order, err := db.FindOrder(ctx, "SELECT * FROM orders WHERE id = $1", cmd.OrderID)
    // ... rest of the logic

    if order.Status == "delivered" {
        // This is a common pattern for failed precondition errors
        return ErrOrderNotCancellable.NewError(
            trogonerror.WithMetadataValuef(trogonerror.VisibilityPublic, "orderId", "gid://shopify/Order/%s", cmd.OrderID),
            trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "reason", "ALREADY_DELIVERED"),
            trogonerror.WithHelpLinkf("Order Management", "https://admin.shopify.com/orders/%s", cmd.OrderID))
    }

    // ... rest of the logic
}
```

### Basic Usage without Template Errors

```go
import "github.com/TrogonStack/trogonerror"

func main() {
    // Create a simple error with clear domain and uppercase reason
    err := trogonerror.NewError("shopify.users", "NOT_FOUND",
        trogonerror.WithCode(trogonerror.NotFound),
        trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"),
        trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "requestedBy", "storefront-api"),
        trogonerror.WithMetadataValue(trogonerror.VisibilityInternal, "shopId", "mystore.myshopify.com"))

    fmt.Println(err.Error())   // "resource not found"
    fmt.Println(err.Domain())  // "shopify.users"
    fmt.Println(err.Reason())  // "NOT_FOUND"
}
```
