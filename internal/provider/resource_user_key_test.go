// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"

	chefc "github.com/go-chef/chef"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUserKey_basic(t *testing.T) {
	var key chefUserKey

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccUserKeyCheckDestroy(&key),
		Steps: []resource.TestStep{
			{
				Config: testSuffixRender(testAccUserKeyConfig_basic),
				Check: resource.ComposeTestCheckFunc(
					testAccUserKeyCheckExists("chef_user_key.test", &key),
					func(s *terraform.State) error {
						if expected := "bdwyertech-github"; key.User != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, key.User)
						}
						if expected := "testing" + testSuffix; key.Key.Name != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, key.Key.Name)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccUserKeyCheckExists(rn string, key *chefUserKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("key id not set")
		}

		c := testAccProvider.Meta().(*chefClient)
		gotClient, err := c.Users.Get(rs.Primary.Attributes["user"])
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		k, err := c.Global.Users.GetKey(gotClient.UserName, rs.Primary.Attributes["key_name"])
		if err != nil {
			return fmt.Errorf("error getting client key: %s", err)
		}

		key.User = gotClient.UserName
		key.Key.Name = k.Name
		key.Key.PublicKey = k.PublicKey

		return nil
	}
}

func testAccUserKeyCheckDestroy(key *chefUserKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := testAccProvider.Meta().(*chefClient)
		o, err := c.Global.Users.GetKey(key.User, key.Key.Name)
		if err == nil {
			return fmt.Errorf("key still exists: %#v", o)
		}
		if _, ok := err.(*chefc.ErrorResponse); !ok {
			// A more specific check is tricky because Chef Server can return
			// a few different error codes in this case depending on which
			// part of its stack catches the error.
			return fmt.Errorf("got something other than an HTTP error (%v) when getting client", err)
		}

		return nil
	}
}

const testAccUserKeyConfig_basic = `
resource "chef_user_key" "test" {
	user = "bdwyertech-github"
	key_name = "testing{{.}}"
	public_key = <<-EOT
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoAGu+lOJXmbCpTpPxwv6
    OPTvr3AP8LqNYuP9t2qlgqichje2JFXaN8VXoF0LIz1+PIN5I/JjCLe0p2h/7FUf
    mU9OR+GRvtfUDHwiYa6v87UWtZhlQLI6D21J48vDc9h1KBmcATVEDXsPYj1wOis4
    QRn5iKkssSg7aZTUnT/AzElRU53rN2j+aL5/yRbtE2VOO2zv0jIG2F6CF3swCxY/
    HRS2vsH4yAQA+XDvmIPuPuqYGEQXszEMHImcIOE19bOONBvx1I3K86baFBpN2OMx
    U/3nHJLdFurCUZVpzQsP11SNcy0lmnr6lpfyn93EiQM7GXvS50cQd8caX2WzuV60
    fQIDAQAB
    -----END PUBLIC KEY-----
    EOT
}
`
