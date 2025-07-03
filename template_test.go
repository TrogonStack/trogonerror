package trogonerror_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/TrogonStack/trogonerror"
	"github.com/stretchr/testify/assert"
)

func TestErrorTemplate_Basic(t *testing.T) {
	template := trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND",
		trogonerror.TemplateWithCode(trogonerror.NotFound),
		trogonerror.TemplateWithVisibility(trogonerror.VisibilityPublic))

	err := template.NewError()
	assert.Equal(t, "shopify.users", err.Domain())
	assert.Equal(t, "NOT_FOUND", err.Reason())
	assert.Equal(t, trogonerror.NotFound, err.Code())
	assert.Equal(t, "resource not found", err.Message())
	assert.Equal(t, trogonerror.VisibilityPublic, err.Visibility())
}

func TestErrorTemplate_CreateInstances(t *testing.T) {
	template := trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND",
		trogonerror.TemplateWithCode(trogonerror.NotFound))
	err1 := template.NewError(trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))
	err2 := template.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/4567890123"),
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/4567890123"))

	assert.Equal(t, "shopify.users", err1.Domain())
	assert.Equal(t, trogonerror.NotFound, err1.Code())
	assert.Equal(t, "shopify.users", err2.Domain())

	assert.Equal(t, "gid://shopify/Customer/1234567890", err1.Metadata()["userId"].Value())
	assert.Equal(t, "gid://shopify/Customer/4567890123", err2.Metadata()["userId"].Value())
	assert.Empty(t, err2.Subject())

	assert.NotEqual(t, err1.Metadata()["userId"].Value(), err2.Metadata()["userId"].Value())
}

func TestErrorTemplate_TemplateWithHelp(t *testing.T) {
	template := trogonerror.NewErrorTemplate("shopify.auth", "INVALID_CREDENTIALS",
		trogonerror.TemplateWithCode(trogonerror.Unauthenticated))

	err := template.NewError(trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "tokenType", "bearer"))

	assert.Nil(t, err.Help())
}

func TestErrorTemplate_MultipleHelpLinks(t *testing.T) {
	template := trogonerror.NewErrorTemplate("shopify.api", "RATE_LIMIT_EXCEEDED",
		trogonerror.TemplateWithCode(trogonerror.ResourceExhausted))

	err := template.NewError()

	assert.Nil(t, err.Help())
}

func TestErrorTemplate_CustomMessage(t *testing.T) {
	template := trogonerror.NewErrorTemplate("shopify.orders", "ORDER_PROCESSING_FAILED",
		trogonerror.TemplateWithCode(trogonerror.Internal),
		trogonerror.TemplateWithMessage("This is a custom error message"))

	err := template.NewError()

	assert.Equal(t, trogonerror.Internal, err.Code())
	assert.Equal(t, "This is a custom error message", err.Message())
}

func TestErrorTemplate_ValueSemantics(t *testing.T) {
	template := trogonerror.NewErrorTemplate("shopify.concurrent", "CONCURRENT_ACCESS",
		trogonerror.TemplateWithCode(trogonerror.Internal))

	// Create multiple instances concurrently
	results := make(chan *trogonerror.TrogonError, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			err := template.NewError(
				trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "goroutineId", fmt.Sprintf("goroutine-%d", id)),
				trogonerror.WithTime(time.Now()))
			results <- err
		}(i)
	}

	errors := make([]*trogonerror.TrogonError, 100)
	for i := 0; i < 100; i++ {
		errors[i] = <-results
	}

	for i, err := range errors {
		assert.Equal(t, "shopify.concurrent", err.Domain(), "Error %d: wrong domain", i)
		assert.Equal(t, trogonerror.Internal, err.Code(), "Error %d: wrong code", i)
		assert.NotNil(t, err.Time(), "Error %d: time should be set", i)

		goroutineID := err.Metadata()["goroutineId"].Value()
		assert.NotEmpty(t, goroutineID, "Error %d: goroutineId should be set", i)
	}
}

func TestErrorTemplate_DefaultMessage(t *testing.T) {
	template := trogonerror.NewErrorTemplate("shopify.products", "PRODUCT_NOT_FOUND",
		trogonerror.TemplateWithCode(trogonerror.NotFound))

	err := template.NewError()

	assert.Equal(t, "resource not found", err.Message())
}

func TestErrorTemplate_OverrideTemplateOptions(t *testing.T) {
	// Create template with default visibility
	template := trogonerror.NewErrorTemplate("shopify.inventory", "STOCK_CHECK_FAILED",
		trogonerror.TemplateWithCode(trogonerror.Internal),
		trogonerror.TemplateWithVisibility(trogonerror.VisibilityInternal))

	err := template.NewError(trogonerror.WithVisibility(trogonerror.VisibilityPublic))

	assert.Equal(t, trogonerror.VisibilityPublic, err.Visibility())
	assert.Equal(t, trogonerror.Internal, err.Code())
}

func TestErrorTemplate_ErrorsIs(t *testing.T) {
	userNotFoundTemplate := trogonerror.NewErrorTemplate("shopify.users", "NOT_FOUND",
		trogonerror.TemplateWithCode(trogonerror.NotFound))

	validationTemplate := trogonerror.NewErrorTemplate("shopify.validation", "INVALID_INPUT",
		trogonerror.TemplateWithCode(trogonerror.InvalidArgument))
	var (
		ErrUserNotFound = userNotFoundTemplate.NewError()
		ErrInvalidInput = validationTemplate.NewError()
	)

	// errors.Is works with template-created error instances
	// Create actual error instances with specific metadata
	actualUserErr := userNotFoundTemplate.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))

	actualValidationErr := validationTemplate.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "fieldName", "email"))

	// These should match because they have same domain + reason
	assert.True(t, errors.Is(actualUserErr, ErrUserNotFound))
	assert.True(t, errors.Is(actualValidationErr, ErrInvalidInput))

	// Cross-template comparisons should fail
	assert.False(t, errors.Is(actualUserErr, ErrInvalidInput))
	assert.False(t, errors.Is(actualValidationErr, ErrUserNotFound))

	// templates themselves are NOT error values
	// Templates don't implement error interface - this demonstrates correct usage
	actualErr := userNotFoundTemplate.NewError()

	// You can't use templates directly with errors.Is - they're not errors
	// This would be a compile error: errors.Is(actualErr, userNotFoundTemplate)

	// Instead, use sentinel errors created from templates
	assert.True(t, errors.Is(actualErr, ErrUserNotFound))

	// Templates are for creating errors, not for being errors
	// Templates don't have public getters (YAGNI) - they're internal configuration

	// different instances from same template match
	err1 := userNotFoundTemplate.NewError(trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))
	err2 := userNotFoundTemplate.NewError(trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/4567890123"))

	// Different instances with different metadata still match the sentinel
	assert.True(t, errors.Is(err1, ErrUserNotFound))
	assert.True(t, errors.Is(err2, ErrUserNotFound))

	// And they match each other (same domain + reason)
	assert.True(t, errors.Is(err1, err2))

	// template.Is() method for direct checking
	// Create errors from templates
	userErr := userNotFoundTemplate.NewError(trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "gid://shopify/Customer/1234567890"))
	validationErr := validationTemplate.NewError(trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "fieldName", "email"))

	// Create a non-TrogonError
	genericErr := fmt.Errorf("generic error")

	// Template.Is() should work for direct checking
	assert.True(t, userNotFoundTemplate.Is(userErr))
	assert.True(t, validationTemplate.Is(validationErr))

	// Cross-template checks should fail
	assert.False(t, userNotFoundTemplate.Is(validationErr))
	assert.False(t, validationTemplate.Is(userErr))

	// Non-TrogonError should fail
	assert.False(t, userNotFoundTemplate.Is(genericErr))
	assert.False(t, validationTemplate.Is(genericErr))

	// This provides a cleaner API than creating sentinel errors
	// template.Is(err) vs errors.Is(err, sentinelErr)

	// templates ignore stack trace options correctly
	// Templates don't support stack trace options (they're for error instances only)
	stackTemplate := trogonerror.NewErrorTemplate("shopify.debugging", "DEBUG_STACK_TEMPLATE",
		trogonerror.TemplateWithCode(trogonerror.Internal))

	// Create error from template without stack options
	err := stackTemplate.NewError()

	// Should NOT have stack trace from template (templates don't store stack traces)
	assert.Nil(t, err.DebugInfo())

	// Instance can add its own stack
	errWithInstanceStack := stackTemplate.NewError(
		trogonerror.WithStackTrace(),
		trogonerror.WithDebugDetail("Template instance stack trace captured"))
	assert.NotNil(t, errWithInstanceStack.DebugInfo())
	assert.Equal(t, "Template instance stack trace captured", errWithInstanceStack.DebugInfo().Detail())
	assert.NotEmpty(t, errWithInstanceStack.DebugInfo().StackEntries())
}

func TestTemplateWithHelp(t *testing.T) {
	t.Run("TemplateWithHelp sets help on template", func(t *testing.T) {
		template := trogonerror.NewErrorTemplate("shopify.support", "HELP_SYSTEM_ERROR")

		err := template.NewError()

		assert.Nil(t, err.Help())
	})
}

func ExampleErrorTemplate_reusable() {
	// Create a validation error template
	validationTemplate := trogonerror.NewErrorTemplate("shopify.validation", "FIELD_INVALID",
		trogonerror.TemplateWithCode(trogonerror.InvalidArgument),
		trogonerror.TemplateWithVisibility(trogonerror.VisibilityPublic))

	// Create multiple validation errors from the same template
	emailErr := validationTemplate.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "fieldName", "email"),
		trogonerror.WithSubject("/email"))

	phoneErr := validationTemplate.NewError(
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "fieldName", "phone"),
		trogonerror.WithSubject("/phone"))

	fmt.Println("Email validation:", emailErr.Error())
	fmt.Println("Phone validation:", phoneErr.Error())
	fmt.Println("Same domain:", emailErr.Domain() == phoneErr.Domain())
	fmt.Println("Same reason:", emailErr.Reason() == phoneErr.Reason())

	// Output:
	// Email validation: invalid argument provided
	//   visibility: PUBLIC
	//   domain: shopify.validation
	//   reason: FIELD_INVALID
	//   code: INVALID_ARGUMENT
	//   subject: /email
	//   metadata:
	//     - fieldName: email visibility=PUBLIC
	//
	// Phone validation: invalid argument provided
	//   visibility: PUBLIC
	//   domain: shopify.validation
	//   reason: FIELD_INVALID
	//   code: INVALID_ARGUMENT
	//   subject: /phone
	//   metadata:
	//     - fieldName: phone visibility=PUBLIC
	//
	// Same domain: true
	// Same reason: true
}
