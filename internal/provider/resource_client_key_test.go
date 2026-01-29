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

func TestAccClientKey_basic(t *testing.T) {
	var key chefClientKey

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccClientKeyCheckDestroy(&key),
		Steps: []resource.TestStep{
			{
				Config: testSuffixRender(testAccClientKeyConfig_basic),
				Check: resource.ComposeTestCheckFunc(
					testAccClientKeyCheckExists("chef_client_key.test", &key),
					func(s *terraform.State) error {
						if expected := "terraform-acc-client-key-test-basic-" + testSuffix; key.Client != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, key.Client)
						}
						if expected := "default"; key.Key.Name != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, key.Key.Name)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccClientKeyCheckExists(rn string, key *chefClientKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("key id not set")
		}

		c := testAccProvider.Meta().(*chefClient)
		gotClient, err := c.Clients.Get(rs.Primary.Attributes["client"])
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		k, err := c.Clients.GetKey(gotClient.Name, rs.Primary.Attributes["key_name"])
		if err != nil {
			return fmt.Errorf("error getting client key: %s", err)
		}

		key.Client = gotClient.Name
		key.Key.Name = k.Name
		key.Key.PublicKey = k.PublicKey

		return nil
	}
}

func testAccClientKeyCheckDestroy(key *chefClientKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := testAccProvider.Meta().(*chefClient)
		o, err := c.Clients.Get(key.Client)
		if err == nil {
			return fmt.Errorf("client still exists: %#v", o)
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

const testAccClientKeyConfig_basic = `
resource "chef_client" "test-key" {
  name = "terraform-acc-client-key-test-basic-{{.}}"
  validator = true
}

resource "chef_client_key" "test" {
	depends_on = [chef_client.test-key]
	client = chef_client.test-key.name
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
