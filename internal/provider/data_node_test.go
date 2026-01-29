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

func TestAccDataNode_basic(t *testing.T) {
	var node chefc.Node

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccNodeCheckDestroy(&node),
		Steps: []resource.TestStep{
			{
				Config: testSuffixRender(testAccDataNodeConfig_basic),
				Check: resource.ComposeTestCheckFunc(
					testAccNodeCheckExists("data.chef_node.test", &node),
					func(s *terraform.State) error {

						if expected := "terraform-acc-test-basic-" + testSuffix; node.Name != expected {
							return fmt.Errorf("wrong name; expected %v, got %v", expected, node.Name)
						}
						if expected := "terraform-acc-test-node-basic-" + testSuffix; node.Environment != expected {
							return fmt.Errorf("wrong environment; expected %v, got %v", expected, node.Environment)
						}

						expectedRunList := []string{
							"recipe[terraform@1.0.0]",
							"recipe[consul]",
							"role[foo]",
						}
						if !reflect.DeepEqual(node.RunList, expectedRunList) {
							return fmt.Errorf("wrong runlist; expected %#v, got %#v", expectedRunList, node.RunList)
						}

						expectedAttributes := map[string]interface{}{
							"terraform_acc_test": true,
						}
						if !reflect.DeepEqual(node.AutomaticAttributes, expectedAttributes) {
							return fmt.Errorf("wrong automatic attributes; expected %#v, got %#v", expectedAttributes, node.AutomaticAttributes)
						}
						if !reflect.DeepEqual(node.NormalAttributes, expectedAttributes) {
							return fmt.Errorf("wrong normal attributes; expected %#v, got %#v", expectedAttributes, node.NormalAttributes)
						}
						if !reflect.DeepEqual(node.DefaultAttributes, expectedAttributes) {
							return fmt.Errorf("wrong default attributes; expected %#v, got %#v", expectedAttributes, node.DefaultAttributes)
						}
						if !reflect.DeepEqual(node.OverrideAttributes, expectedAttributes) {
							return fmt.Errorf("wrong override attributes; expected %#v, got %#v", expectedAttributes, node.OverrideAttributes)
						}

						return nil
					},
				),
			},
		},
	})
}

const testAccDataNodeConfig_basic = `
resource "chef_node" "test" {
  name = "terraform-acc-test-basic-{{.}}"
  environment_name = "terraform-acc-test-node-basic-{{.}}"
  automatic_attributes_json = <<EOT
{
     "terraform_acc_test": true
}
EOT
  normal_attributes_json = <<EOT
{
     "terraform_acc_test": true
}
EOT
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
  run_list = ["terraform@1.0.0", "recipe[consul]", "role[foo]"]
}

data "chef_node" "test" {
	name = chef_node.test.id
}
`
