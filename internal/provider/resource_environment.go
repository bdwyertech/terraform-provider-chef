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

func resourceChefEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: CreateEnvironment,
		UpdateContext: UpdateEnvironment,
		ReadContext:   ReadEnvironment,
		DeleteContext: DeleteEnvironment,
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
			"allow_overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			"cookbook_constraints": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func CreateEnvironment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	env, err := environmentFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error loading environment from resource data",
				Detail:   fmt.Sprint(err),
			},
		}
	}

	// Check if the environment already exists
	if d.Id() != "" {
		if d.Get("allow_overwrite").(bool) {
			return UpdateEnvironment(ctx, d, meta)
		} else {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Environment already exists and allow_overwrite is set to false",
				},
			}
		}
	}

	_, err = client.Environments.Create(env)
	if err != nil {
		resp := diag.Diagnostic{Severity: diag.Error, Summary: "Error creating environment"}
		if cheferr, ok := err.(*chefc.ErrorResponse); ok {
			resp.Detail = fmt.Sprintln(cheferr.ErrorMsg, cheferr)
		} else {
			resp.Detail = fmt.Sprint(err)
		}
		return diag.Diagnostics{resp}
	}

	return ReadEnvironment(ctx, d, meta)
}

func UpdateEnvironment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	env, err := environmentFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error loading environment from resource data",
				Detail:   fmt.Sprint(err),
			},
		}
	}

	_, err = client.Environments.Put(env)
	if err != nil {
		resp := diag.Diagnostic{Severity: diag.Error, Summary: "Error updating environment"}
		if cheferr, ok := err.(*chefc.ErrorResponse); ok {
			resp.Detail = fmt.Sprintln(cheferr.ErrorMsg, cheferr)
		} else {
			resp.Detail = fmt.Sprint(err)
		}
		return diag.Diagnostics{resp}
	}

	return ReadEnvironment(ctx, d, meta)
}

func ReadEnvironment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	env, err := client.Environments.Get(d.Get("name").(string))
	if err != nil {
		resp := diag.Diagnostic{Severity: diag.Error, Summary: "Error reading environment"}
		if cheferr, ok := err.(*chefc.ErrorResponse); ok {
			if cheferr.Response.StatusCode == 404 {
				d.SetId("")
				return nil
			}
			resp.Detail = fmt.Sprintln(cheferr.ErrorMsg, cheferr)
		} else {
			resp.Detail = fmt.Sprint(err)
		}
		return diag.Diagnostics{resp}
	}

	d.SetId(env.Name)
	d.Set("name", env.Name)
	d.Set("description", env.Description)
	envJson, err := json.Marshal(env)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error converting environment into JSON",
				Detail:   fmt.Sprint(err),
			},
		}
	}
	d.Set("json", string(envJson))

	defaultAttrJson, err := json.Marshal(env.DefaultAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error parsing default attributes",
				Detail:   fmt.Sprint(err),
			},
		}
	}
	d.Set("default_attributes_json", string(defaultAttrJson))

	overrideAttrJson, err := json.Marshal(env.OverrideAttributes)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Error parsing override attributes",
				Detail:   fmt.Sprint(err),
			},
		}
	}
	d.Set("override_attributes_json", string(overrideAttrJson))

	cookbookVersionsI := map[string]interface{}{}
	for k, v := range env.CookbookVersions {
		cookbookVersionsI[k] = v
	}
	d.Set("cookbook_constraints", cookbookVersionsI)

	return nil
}

func DeleteEnvironment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*chefClient)

	name := d.Id()

	if _, err := client.Environments.Delete(name); err == nil {
		d.SetId("")
	} else {
		resp := diag.Diagnostic{Severity: diag.Error, Summary: "Error deleting environment"}
		if cheferr, ok := err.(*chefc.ErrorResponse); ok {
			resp.Detail = fmt.Sprintln(cheferr.ErrorMsg, cheferr)
		} else {
			resp.Detail = fmt.Sprint(err)
		}
		return diag.Diagnostics{resp}
	}

	return nil
}

func environmentFromResourceData(d *schema.ResourceData) (*chefc.Environment, error) {

	env := &chefc.Environment{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ChefType:    "environment",
	}

	var err error

	err = json.Unmarshal(
		[]byte(d.Get("default_attributes_json").(string)),
		&env.DefaultAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("default_attributes_json: %s", err)
	}

	err = json.Unmarshal(
		[]byte(d.Get("override_attributes_json").(string)),
		&env.OverrideAttributes,
	)
	if err != nil {
		return nil, fmt.Errorf("override_attributes_json: %s", err)
	}

	env.CookbookVersions = make(map[string]string)
	for k, vI := range d.Get("cookbook_constraints").(map[string]interface{}) {
		env.CookbookVersions[k] = vI.(string)
	}

	return env, nil
}
