package trogonerror_test

import (
	"testing"

	"github.com/TrogonStack/trogonerror"
	testdatav1 "github.com/TrogonStack/trogonerror/internal/testdata/gen/trogonerror/testdata/v1"
	"github.com/stretchr/testify/assert"
)

var userNotFoundTemplate = trogonerror.NewErrorTemplateFromProto[*testdatav1.UserNotFound]()

func TestNewErrorTemplateFromProto_TemplateLevel(t *testing.T) {
	err := userNotFoundTemplate.NewError()

	assert.Equal(t, "shopify.users", err.Domain())
	assert.Equal(t, "USER_NOT_FOUND", err.Reason())
	assert.Equal(t, trogonerror.CodeNotFound, err.Code())
	assert.Equal(t, "User does not exist.", err.Message())
	assert.Equal(t, trogonerror.VisibilityPublic, err.Visibility())

	help := err.Help()
	if assert.NotNil(t, help) {
		assert.Equal(t, "https://docs.shopify.com/users", help.Links()[0].URL())
	}
}

func TestNewErrorTemplateFromProto_TemplateMetadata(t *testing.T) {
	md := userNotFoundTemplate.NewError().Metadata()

	assert.Equal(t, "users", md["component"].Value())
	assert.Equal(t, trogonerror.VisibilityPublic, md["component"].Visibility())
	assert.Equal(t, "platform-identity", md["team"].Value())
	assert.Equal(t, trogonerror.VisibilityInternal, md["team"].Visibility())
}

func TestNewErrorTemplateFromProto_FieldFixedValueBaked(t *testing.T) {
	md := userNotFoundTemplate.NewError().Metadata()

	assert.Equal(t, "us-east-1", md["region"].Value(), "value_policy=value should bake into template")
	assert.Equal(t, trogonerror.VisibilityPublic, md["region"].Visibility())
}

func TestErrorTemplate_FromProto_RuntimeFieldValues(t *testing.T) {
	err := userNotFoundTemplate.FromProto(&testdatav1.UserNotFound{
		UserId:   "gid://shopify/Customer/1234",
		TenantId: "acme",
	})

	md := err.Metadata()
	assert.Equal(t, "gid://shopify/Customer/1234", md["userId"].Value())
	assert.Equal(t, "acme", md["tenantId"].Value())
	assert.Equal(t, "us-east-1", md["region"].Value())
}

func TestErrorTemplate_FromProto_DefaultValueFallback(t *testing.T) {
	err := userNotFoundTemplate.FromProto(&testdatav1.UserNotFound{
		UserId: "gid://shopify/Customer/1234",
	})

	md := err.Metadata()
	assert.Equal(t, "default-tenant", md["tenantId"].Value(), "empty field should fall back to default_value")
}

func TestErrorTemplate_FromProto_FixedValueWinsOverRuntime(t *testing.T) {
	err := userNotFoundTemplate.FromProto(&testdatav1.UserNotFound{
		UserId: "gid://shopify/Customer/1234",
		Region: "ignored",
	})

	assert.Equal(t, "us-east-1", err.Metadata()["region"].Value(),
		"value_policy=value must ignore runtime field")
}

func TestErrorTemplate_FromProto_OptionsOverride(t *testing.T) {
	err := userNotFoundTemplate.FromProto(
		&testdatav1.UserNotFound{UserId: "u1"},
		trogonerror.WithMetadataValue(trogonerror.VisibilityPublic, "userId", "u2"),
	)

	assert.Equal(t, "u2", err.Metadata()["userId"].Value(),
		"caller options must win over proto-derived metadata")
}

func TestNewErrorTemplateFromProto_TemplateOptionOverride(t *testing.T) {
	template := trogonerror.NewErrorTemplateFromProto[*testdatav1.UserNotFound](
		trogonerror.TemplateWithMessage("custom message"),
	)

	assert.Equal(t, "custom message", template.NewError().Message())
}
