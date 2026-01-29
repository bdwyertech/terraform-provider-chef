// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataChefEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadContext: ReadEnvironment,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_attributes_json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"override_attributes_json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cookbook_constraints": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
