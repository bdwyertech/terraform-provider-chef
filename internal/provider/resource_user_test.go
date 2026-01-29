package provider

import (
	"fmt"
	"testing"

	chefc "github.com/go-chef/chef"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUser_basic(t *testing.T) {
	var user chefc.User

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccUserCheckDestroy(&user),
		Steps: []resource.TestStep{
			{
				Config: testSuffixRender(testAccUserConfig_basic),
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheckExists("chef_user.test", &user),
					func(s *terraform.State) error {

						if expected := "terraform-acc-user-test-basic-" + testSuffix; user.UserName != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, user.UserName)
						}
						// if expected := true; client.Validator != expected {
						// 	return fmt.Errorf("wrong environment; expected %v, got %v", expected, client.Validator)
						// }

						return nil
					},
				),
			},
		},
	})
}

func testAccUserCheckExists(rn string, user *chefc.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("user id not set")
		}

		c := testAccProvider.Meta().(*chefClient)
		gotUser, err := c.Global.Users.Get(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting user: %s", err)
		}

		user.UserName = gotUser.UserName

		if _, err := c.Users.Get(user.UserName); err != nil {
			return fmt.Errorf("error getting org user: %s", err)
		}

		return nil
	}
}

func testAccUserCheckDestroy(user *chefc.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := testAccProvider.Meta().(*chefClient)
		o, err := c.Users.Get(user.UserName)
		if err == nil {
			return fmt.Errorf("user still exists: %#v", o)
		}
		if _, ok := err.(*chefc.ErrorResponse); !ok {
			// A more specific check is tricky because Chef Server can return
			// a few different error codes in this case depending on which
			// part of its stack catches the error.
			return fmt.Errorf("got something other than an HTTP error (%v) when getting user", err)
		}

		return nil
	}
}

const testAccUserConfig_basic = `
resource "chef_user" "test" {
  username = "terraform-acc-user-test-basic-{{.}}"
  email = "terraform-acc-user-test-basic-{{.}}@test.local"
  password = "Abc123456!"
}
`
