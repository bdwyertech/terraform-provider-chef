// Copyright © Brian Dwyer - Intelligent Digital Services 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	chefc "github.com/go-chef/chef"
)

func resourceChefClient() *schema.Resource {
	return &schema.Resource{
		Create: CreateClient,
		Update: UpdateClient,
		Read:   ReadClient,
		Delete: DeleteClient,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"validator": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func CreateClient(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*chefClient)

	client, err := clientFromResourceData(d)
	if err != nil {
		return err
	}

	if _, err := c.Clients.Create(*client); err != nil {
		return err
	}

	d.SetId(client.Name)
	return ReadClient(d, meta)
}

func UpdateClient(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*chefClient)

	client, err := clientFromResourceData(d)
	if err != nil {
		return err
	}

	_, err = c.Clients.Update(client.Name, *client)
	if err != nil {
		return err
	}

	d.SetId(client.Name)
	return ReadClient(d, meta)
}

func ReadClient(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*chefClient)

	name := d.Id()

	client, err := c.Clients.Get(name)
	if err != nil {
		if errRes, ok := err.(*chefc.ErrorResponse); ok {
			if errRes.Response.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		} else {
			return err
		}
	}

	d.Set("name", client.Name)
	d.Set("validator", client.Validator)

	return nil
}

func DeleteClient(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*chefClient)

	name := d.Id()
	err := c.Clients.Delete(name)

	if err == nil {
		d.SetId("")
	}

	return err
}

func clientFromResourceData(d *schema.ResourceData) (*chefc.ApiNewClient, error) {
	client := &chefc.ApiNewClient{
		Name:      d.Get("name").(string),
		Validator: d.Get("validator").(bool),
		CreateKey: false,
	}
	return client, nil
}
