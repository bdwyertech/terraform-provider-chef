// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	chefc "github.com/go-chef/chef"
)

func resourceChefNode() *schema.Resource {
	return &schema.Resource{
		CreateContext: CreateNode,
		UpdateContext: UpdateNode,
		ReadContext:   ReadNode,
		DeleteContext: DeleteNode,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "_default",
			},
			"automatic_attributes_json": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "{}",
				StateFunc: jsonStateFunc,
			},
			"normal_attributes_json": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "{}",
				StateFunc: jsonStateFunc,
			},
			"default_attributes_json": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "{}",
				StateFunc: jsonStateFunc,
			},
			"override_attributes_json": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "{}",
				StateFunc: jsonStateFunc,
			},
			"run_list": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:      schema.TypeString,
					StateFunc: runListEntryStateFunc,
				},
			},
		},
	}
}

func CreateNode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	node, err := nodeFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error loading node from resource data",
				Detail:   fmt.Sprint(err),
			},
		}
	}

	_, err = client.Nodes.Post(*node)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error creating node",
				Detail:   fmt.Sprint(err),
			},
		}
	}

	return ReadNode(ctx, d, meta)
}

func UpdateNode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	node, err := nodeFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error loading node from resource data",
				Detail:   fmt.Sprint(err),
			},
		}
	}

	_, err = client.Nodes.Put(*node)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error updating node",
				Detail:   fmt.Sprint(err),
			},
		}
	}

	return ReadNode(ctx, d, meta)
}

func ReadNode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	node, err := client.Nodes.Get(d.Get("name").(string))
	if err != nil {
		if errRes, ok := err.(*chefc.ErrorResponse); ok {
			if errRes.Response.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		} else {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Error reading node",
					Detail:   fmt.Sprint(err),
				},
			}
		}
	}

	d.SetId(node.Name)
	d.Set("name", node.Name)
	d.Set("environment_name", node.Environment)

	automaticAttrJson, err := json.Marshal(node.AutomaticAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error parsing automatic attributes as JSON",
				Detail:   fmt.Sprint(err),
			},
		}
	}
	d.Set("automatic_attributes_json", string(automaticAttrJson))

	normalAttrJson, err := json.Marshal(node.NormalAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error parsing normal attributes as JSON",
				Detail:   fmt.Sprint(err),
			},
		}
	}
	d.Set("normal_attributes_json", string(normalAttrJson))

	defaultAttrJson, err := json.Marshal(node.DefaultAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error parsing default attributes as JSON",
				Detail:   fmt.Sprint(err),
			},
		}
	}
	d.Set("default_attributes_json", string(defaultAttrJson))

	overrideAttrJson, err := json.Marshal(node.OverrideAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error parsing override attributes as JSON",
				Detail:   fmt.Sprint(err),
			},
		}
	}
	d.Set("override_attributes_json", string(overrideAttrJson))

	runListI := make([]interface{}, len(node.RunList))
	for i, v := range node.RunList {
		runListI[i] = v
	}
	d.Set("run_list", runListI)

	return nil
}

func DeleteNode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	name := d.Id()
	if err := client.Nodes.Delete(name); err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error deleting node",
				Detail:   fmt.Sprint(err),
			},
		}
	}

	d.SetId("")

	return nil
}

func nodeFromResourceData(d *schema.ResourceData) (*chefc.Node, error) {

	node := &chefc.Node{
		Name:        d.Get("name").(string),
		Environment: d.Get("environment_name").(string),
		ChefType:    "node",
		JsonClass:   "Chef::Node",
	}

	var err error

	err = json.Unmarshal(
		[]byte(d.Get("automatic_attributes_json").(string)),
		&node.AutomaticAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("automatic_attributes_json: %s", err)
	}

	err = json.Unmarshal(
		[]byte(d.Get("normal_attributes_json").(string)),
		&node.NormalAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("normal_attributes_json: %s", err)
	}

	err = json.Unmarshal(
		[]byte(d.Get("default_attributes_json").(string)),
		&node.DefaultAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("default_attributes_json: %s", err)
	}

	err = json.Unmarshal(
		[]byte(d.Get("override_attributes_json").(string)),
		&node.OverrideAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("override_attributes_json: %s", err)
	}

	runListI := d.Get("run_list").([]interface{})
	node.RunList = make([]string, len(runListI))
	for i, vI := range runListI {
		node.RunList[i] = vI.(string)
	}

	return node, nil
}
