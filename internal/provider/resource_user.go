// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	chefc "github.com/go-chef/chef"
)

func resourceChefUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: CreateUser,
		UpdateContext: UpdateUser,
		ReadContext:   ReadUser,
		DeleteContext: DeleteUser,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"associate": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ExactlyOneOf: []string{
					"password",
					"external_auth_uid",
				},
			},
			"external_auth_uid": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func CreateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*chefClient)

	user, err := userFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error reading Chef User from Resource Data",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("username"),
			},
		}
	}

	if _, err := c.Global.Users.Create(*user); err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error creating Chef User",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("username"),
			},
		}
	}

	if invitation, err := c.Associations.Invite(chefc.Request{User: user.UserName}); err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error associating Chef User with Organization",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("username"),
			},
		}
	} else {
		if _, err = c.Associations.AcceptInvite(path.Base(invitation.Uri)); err != nil {
			return diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Error accepting Organization invite",
					Detail:        fmt.Sprint(err),
					AttributePath: cty.GetAttrPath("username"),
				},
			}
		}
	}

	d.SetId(user.UserName)
	return ReadUser(ctx, d, meta)
}

func UpdateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*chefClient)

	user, err := userFromResourceData(d)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error reading Chef User from Resource Data",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("username"),
			},
		}
	}

	_, err = c.Global.Users.Update(user.UserName, *user)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error updating Chef User",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("username"),
			},
		}
	}

	d.SetId(user.UserName)
	return ReadUser(ctx, d, meta)
}

func ReadUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*chefClient)

	name := d.Id()

	user, err := c.Global.Users.Get(name)
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
					Summary:       "Error reading Chef User from Resource Data",
					Detail:        fmt.Sprint(err),
					AttributePath: cty.GetAttrPath("username"),
				},
			}
		}
	}

	d.Set("username", user.UserName)
	d.Set("email", user.Email)
	d.Set("display_name", user.DisplayName)
	d.Set("first_name", user.FirstName)
	d.Set("last_name", user.LastName)

	return nil
}

func DeleteUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*chefClient)

	name := d.Id()
	err := c.Global.Users.Delete(name)

	if err == nil {
		d.SetId("")
		return nil
	}

	return diag.Diagnostics{
		{
			Severity:      diag.Error,
			Summary:       "Error deleting Chef User",
			Detail:        fmt.Sprint(err),
			AttributePath: cty.GetAttrPath("username"),
		},
	}
}

func userFromResourceData(d *schema.ResourceData) (*chefc.User, error) {
	user := &chefc.User{
		UserName: d.Get("username").(string),
		Password: d.Get("password").(string),
		// CreateKey: false,
		// PublicKey:   "-----BEGIN RSA PUBLIC KEY-----",
		Email:       d.Get("email").(string),
		FirstName:   d.Get("first_name").(string),
		LastName:    d.Get("last_name").(string),
		DisplayName: d.Get("display_name").(string),
	}
	if user.DisplayName == "" {
		user.DisplayName = user.UserName
	}
	return user, nil
}
