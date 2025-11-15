package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	chefc "github.com/go-chef/chef"
)

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		return &schema.Provider{
			ConfigureContextFunc: providerConfigure,
			DataSourcesMap: map[string]*schema.Resource{
				"chef_environment": dataChefEnvironment(),
				"chef_node":        dataChefNode(),
				"chef_search":      dataChefSearch(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"chef_data_bag":      resourceChefDataBag(),
				"chef_data_bag_item": resourceChefDataBagItem(),
				"chef_environment":   resourceChefEnvironment(),
				"chef_client":        resourceChefClient(),
				"chef_client_key":    resourceChefClientKey(),
				"chef_node":          resourceChefNode(),
				"chef_role":          resourceChefRole(),
				"chef_user_key":      resourceChefUserKey(),
			},
			Schema: map[string]*schema.Schema{
				"server_url": {
					Type:         schema.TypeString,
					Required:     true,
					DefaultFunc:  schema.EnvDefaultFunc("CHEF_SERVER_URL", nil),
					Description:  "URL of the root of the target Chef server or organization.",
					ValidateFunc: validateServerURL,
				},
				"client_name": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("CHEF_CLIENT_NAME", nil),
					Description: "Name of a registered client within the Chef server.",
				},
				"private_key_pem": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: providerPrivateKeyEnvDefault,
					Deprecated:  "Please use key_material instead",
				},
				"key_material": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("CHEF_KEY_MATERIAL", ""),
					Description: "PEM-formatted private key for client authentication.",
				},
				"key_material_base64": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("CHEF_KEY_MATERIAL_BASE64", ""),
					Description: "Base64-encoded PEM-formatted private key for client authentication. This is useful when storing keys in systems that don't preserve newlines (e.g., Jenkins Secret Text).",
				},

				"allow_unverified_ssl": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "If set, the Chef client will permit unverifiable SSL certificates.",
				},
			},
		}
	}
}

type chefClient struct {
	*chefc.Client
	Global *chefc.Client
}

func validateServerURL(val interface{}, key string) (warns []string, errs []error) {
	url := val.(string)
	if !strings.HasSuffix(url, "/") {
		errs = append(errs, fmt.Errorf("chef server_url %s must end with a slash", url))
	}
	return
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := &chefc.Config{
		Name:    d.Get("client_name").(string),
		BaseURL: d.Get("server_url").(string),
		SkipSSL: d.Get("allow_unverified_ssl").(bool),
		Timeout: 10,
	}

	if v, ok := d.GetOk("private_key_pem"); ok {
		config.Key = v.(string)
	}

	if v, ok := d.GetOk("key_material"); ok {
		config.Key = v.(string)
	}

	if v, ok := d.GetOk("key_material_base64"); ok {
		decoded, err := base64.StdEncoding.DecodeString(v.(string))
		if err != nil {
			return nil, diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Error decoding base64-encoded private key",
					Detail:        fmt.Sprintf("Failed to decode key_material_base64: %s", err),
					AttributePath: cty.GetAttrPath("key_material_base64"),
				},
			}
		}
		config.Key = string(decoded)
	}

	client, err := chefc.NewClient(config)
	if err != nil {
		return nil, diag.Diagnostics{
			{
				Severity:      diag.Error,
				Summary:       "Error creating Chef Client",
				Detail:        fmt.Sprint(err),
				AttributePath: cty.GetAttrPath("client_name"),
			},
		}
	}
	if split := strings.Split(config.BaseURL, "/organizations/"); len(split) > 1 {
		config.BaseURL = split[0]
		globalClient, err := chefc.NewClient(config)
		if err != nil {
			return nil, diag.Diagnostics{
				{
					Severity:      diag.Error,
					Summary:       "Error creating Chef Client",
					Detail:        fmt.Sprint(err),
					AttributePath: cty.GetAttrPath("client_name"),
				},
			}
		}
		return &chefClient{client, globalClient}, nil
	}

	return &chefClient{client, client}, nil
}

func providerPrivateKeyEnvDefault() (interface{}, error) {
	if fn := os.Getenv("CHEF_PRIVATE_KEY_FILE"); fn != "" {
		contents, err := os.ReadFile(fn)
		if err != nil {
			return nil, err
		}
		return string(contents), nil
	}

	return nil, nil
}

func jsonStateFunc(value interface{}) string {
	// Parse and re-stringify the JSON to make sure it's always kept
	// in a normalized form.
	jsonValue, err := structure.NormalizeJsonString(value)
	if err != nil {
		return "null"
	}

	return jsonValue
}

func runListEntryStateFunc(value interface{}) string {
	// Recipes in run lists can either be naked, like "foo", or can
	// be explicitly qualified as "recipe[foo]". Whichever form we use,
	// the server will always normalize to the explicit form,
	// so we'll normalize too and then we won't generate unnecessary
	// diffs when we refresh.
	in := value.(string)
	if !strings.Contains(in, "[") {
		return fmt.Sprintf("recipe[%s]", in)
	}
	return in
}
