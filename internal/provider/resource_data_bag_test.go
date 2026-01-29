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

func TestAccDataBag_basic(t *testing.T) {
	dataBagName := "terraform-acc-test-basic-" + testSuffix
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccDataBagCheckDestroy(dataBagName),
		Steps: []resource.TestStep{
			{
				Config: testSuffixRender(testAccDataBagConfig_basic),
				Check:  testAccDataBagCheckExists("chef_data_bag.test"),
			},
		},
	})
}

func testAccDataBagCheckExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data bag id not set")
		}

		client := testAccProvider.Meta().(*chefClient)
		_, err := client.DataBags.ListItems(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting data bag: %s", err)
		}

		return nil
	}
}

func testAccDataBagCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*chefClient)
		result, err := client.DataBags.ListItems(name)
		if err == nil && len(*result) != 0 {
			return fmt.Errorf("data bag still exists")
		}
		if _, ok := err.(*chefc.ErrorResponse); err != nil && !ok {
			return fmt.Errorf("got something other than an HTTP error (%v) when getting data bag", err)
		}

		return nil
	}
}

const testAccDataBagConfig_basic = `
resource "chef_data_bag" "test" {
  name = "terraform-acc-test-basic-{{.}}"
}
`
