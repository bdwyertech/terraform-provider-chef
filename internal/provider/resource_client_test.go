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

func TestAccClient_basic(t *testing.T) {
	var client chefc.ApiNewClient

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccClientCheckDestroy(&client),
		Steps: []resource.TestStep{
			{
				Config: testSuffixRender(testAccClientConfig_basic),
				Check: resource.ComposeTestCheckFunc(
					testAccClientCheckExists("chef_client.test", &client),
					func(s *terraform.State) error {

						if expected := "terraform-acc-client-test-basic-" + testSuffix; client.Name != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, client.Name)
						}
						if expected := true; client.Validator != expected {
							return fmt.Errorf("wrong environment; expected %v, got %v", expected, client.Validator)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccClientCheckExists(rn string, client *chefc.ApiNewClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("client id not set")
		}

		c := testAccProvider.Meta().(*chefClient)
		gotClient, err := c.Clients.Get(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		client.Name = gotClient.Name
		client.Validator = gotClient.Validator

		return nil
	}
}

func testAccClientCheckDestroy(client *chefc.ApiNewClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := testAccProvider.Meta().(*chefClient)
		o, err := c.Clients.Get(client.Name)
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

const testAccClientConfig_basic = `
resource "chef_client" "test" {
  name = "terraform-acc-client-test-basic-{{.}}"
  validator = true
}
`
