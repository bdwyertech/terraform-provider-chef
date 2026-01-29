// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataChefNode() *schema.Resource {
	return &schema.Resource{
		ReadContext: ReadNode,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_attributes_json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"normal_attributes_json": {
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
			"run_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type:      schema.TypeString,
					StateFunc: runListEntryStateFunc,
				},
			},
		},
	}
}
