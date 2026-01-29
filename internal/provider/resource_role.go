// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	chefc "github.com/go-chef/chef"
)

func resourceChefRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: CreateRole,
		UpdateContext: UpdateRole,
		ReadContext:   ReadRole,
		DeleteContext: DeleteRole,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
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

func CreateRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	role, err := roleFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error reading Chef Role from Resource Data",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("name"),
			},
		}
	}

	_, err = client.Roles.Create(role)
	if err != nil {
		resp := diag.Diagnostic{Severity: diag.Error, Summary: "Error creating Chef Role", AttributePath: cty.GetAttrPath("name")}
		if cheferr, ok := err.(*chefc.ErrorResponse); ok {
			resp.Detail = fmt.Sprintln(cheferr.ErrorMsg, cheferr)
		} else {
			resp.Detail = fmt.Sprint(err)
		}
		return diag.Diagnostics{resp}
	}

	d.SetId(role.Name)
	return ReadRole(ctx, d, meta)
}

func UpdateRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	role, err := roleFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error reading Chef Role from Resource Data",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("name"),
			},
		}
	}

	_, err = client.Roles.Put(role)
	if err != nil {
		resp := diag.Diagnostic{Severity: diag.Error, Summary: "Error updating Chef Role", AttributePath: cty.GetAttrPath("name")}
		if cheferr, ok := err.(*chefc.ErrorResponse); ok {
			resp.Detail = fmt.Sprintln(cheferr.ErrorMsg, cheferr)
		} else {
			resp.Detail = fmt.Sprint(err)
		}
		return diag.Diagnostics{resp}
	}

	d.SetId(role.Name)
	return ReadRole(ctx, d, meta)
}

func ReadRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	name := d.Id()

	role, err := client.Roles.Get(name)
	if err != nil {
		if errRes, ok := err.(*chefc.ErrorResponse); ok {
			if errRes.Response.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		} else {
			return diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Error reading Chef Role",
					Detail:        fmt.Sprint(err),
					AttributePath: cty.GetAttrPath("name"),
				},
			}
		}
	}

	d.Set("name", role.Name)
	d.Set("description", role.Description)

	defaultAttrJson, err := json.Marshal(role.DefaultAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error reading Chef Role default attributes as JSON",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("name"),
			},
		}
	}
	d.Set("default_attributes_json", string(defaultAttrJson))

	overrideAttrJson, err := json.Marshal(role.OverrideAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error reading Chef Role override attributes as JSON",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("name"),
			},
		}
	}
	d.Set("override_attributes_json", string(overrideAttrJson))

	runList := make([]string, len(role.RunList))
	copy(runList, role.RunList)

	d.Set("run_list", runList)

	return nil
}

func DeleteRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	name := d.Id()

	err := client.Roles.Delete(name)
	if err != nil {
		resp := diag.Diagnostic{Severity: diag.Error, Summary: "Error deleting Chef Role", AttributePath: cty.GetAttrPath("name")}
		if cheferr, ok := err.(*chefc.ErrorResponse); ok {
			resp.Detail = fmt.Sprintln(cheferr.ErrorMsg, cheferr)
		} else {
			resp.Detail = fmt.Sprint(err)
		}
		return diag.Diagnostics{resp}
	}
	return nil
}

func roleFromResourceData(d *schema.ResourceData) (*chefc.Role, error) {

	role := &chefc.Role{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ChefType:    "role",
	}

	err := json.Unmarshal(
		[]byte(d.Get("default_attributes_json").(string)),
		&role.DefaultAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("default_attributes_json: %s", err)
	}

	err = json.Unmarshal(
		[]byte(d.Get("override_attributes_json").(string)),
		&role.OverrideAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("override_attributes_json: %s", err)
	}

	runListI := d.Get("run_list").([]interface{})
	role.RunList = make([]string, len(runListI))
	for i, vI := range runListI {
		role.RunList[i] = vI.(string)
	}

	return role, nil
}
