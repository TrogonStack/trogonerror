package errproto_test

import (
	"fmt"

	"github.com/TrogonStack/trogonerror/errproto"
	testdatav1 "github.com/TrogonStack/trogonerror/internal/testdata/gen/trogonerror/testdata/v1"
)

var userNotFoundTpl = errproto.NewErrorTemplateFromProto[*testdatav1.UserNotFound]()

func ExampleNewErrorTemplateFromProto() {
	err := userNotFoundTpl.FromProto(&testdatav1.UserNotFound{
		UserId: "gid://shopify/Customer/1234567890",
	})

	fmt.Println(err.Error())
	// Output:
	// User does not exist.
	//   visibility: PUBLIC
	//   domain: shopify.users
	//   reason: USER_NOT_FOUND
	//   code: NOT_FOUND
	//   metadata:
	//     - component: users visibility=PUBLIC
	//     - region: us-east-1 visibility=PUBLIC
	//     - team: platform-identity visibility=INTERNAL
	//     - tenantId: default-tenant visibility=PUBLIC
	//     - userId: gid://shopify/Customer/1234567890 visibility=PUBLIC
	//
	// - User Docs: https://docs.shopify.com/users
}
