// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"reflect"
	"testing"

	chefc "github.com/go-chef/chef"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataEnvironment_basic(t *testing.T) {
	var env chefc.Environment

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccEnvironmentCheckDestroy(&env),
		Steps: []resource.TestStep{
			{
				Config: testSuffixRender(testAccDataEnvironmentConfig_basic),
				Check: resource.ComposeTestCheckFunc(
					testAccEnvironmentCheckExists("data.chef_environment.test", &env),
					func(s *terraform.State) error {

						if expected := "terraform-acc-test-basic-" + testSuffix; env.Name != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, env.Name)
						}
						if expected := "Terraform Acceptance Tests"; env.Description != expected {
							return fmt.Errorf("wrong description; expected %v, got %v", expected, env.Description)
						}

						expectedConstraints := map[string]string{
							"terraform": "= 1.0.0",
						}
						if !reflect.DeepEqual(env.CookbookVersions, expectedConstraints) {
							return fmt.Errorf("wrong cookbook constraints; expected %#v, got %#v", expectedConstraints, env.CookbookVersions)
						}

						expectedAttributes := map[string]interface{}{
							"terraform_acc_test": true,
						}
						if !reflect.DeepEqual(env.DefaultAttributes, expectedAttributes) {
							return fmt.Errorf("wrong default attributes; expected %#v, got %#v", expectedAttributes, env.DefaultAttributes)
						}
						if !reflect.DeepEqual(env.OverrideAttributes, expectedAttributes) {
							return fmt.Errorf("wrong override attributes; expected %#v, got %#v", expectedAttributes, env.OverrideAttributes)
						}

						return nil
					},
				),
			},
		},
	})
}

const testAccDataEnvironmentConfig_basic = `
resource "chef_environment" "test" {
  name = "terraform-acc-test-basic-{{.}}"
  description = "Terraform Acceptance Tests"
  default_attributes_json = <<EOT
{
     "terraform_acc_test": true
}
EOT
  override_attributes_json = <<EOT
{
     "terraform_acc_test": true
}
EOT
  cookbook_constraints = {
    "terraform" = "= 1.0.0"
  }
}

data "chef_environment" "test" {
	name = chef_environment.test.id
}
`
