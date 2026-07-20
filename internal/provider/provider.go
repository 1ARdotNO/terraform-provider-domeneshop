package provider

import (
	"context"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/innovationnorway/go-domeneshop/api/v0/domeneshop"
)

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"token": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("DOMENESHOP_TOKEN", nil),
					Description: "A Domeneshop API token. This can also be set with the `DOMENESHOP_TOKEN` environment variable.",
				},
				"secret": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("DOMENESHOP_SECRET", nil),
					Description: "A Domeneshop API secret. This can also be set with the `DOMENESHOP_SECRET` environment variable.",
				},
				"host": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("DOMENESHOP_HOST", nil),
					Description: "The base URL of the Domeneshop API. Defaults to `https://api.domeneshop.no/v0`. This can also be set with the `DOMENESHOP_HOST` environment variable. Mainly useful for testing.",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"domeneshop_domain":   dataSourceDomain(),
				"domeneshop_domains":  dataSourceDomains(),
				"domeneshop_forward":  dataSourceForward(),
				"domeneshop_forwards": dataSourceForwards(),
				"domeneshop_record":   dataSourceRecord(),
				"domeneshop_records":  dataSourceRecords(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"domeneshop_forward": resourceForward(),
				"domeneshop_record":  resourceRecord(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	client *domeneshop.APIClient
	auth   context.Context
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		config := domeneshop.NewConfiguration()
		config.UserAgent = p.UserAgent("terraform-provider-domeneshop", version)
		config.HTTPClient = cleanhttp.DefaultClient()
		config.HTTPClient.Transport = logging.NewTransport("Domeneshop", config.HTTPClient.Transport)

		if v, ok := d.GetOk("host"); ok {
			config.Servers = domeneshop.ServerConfigurations{
				{URL: v.(string)},
			}
		}

		client := domeneshop.NewAPIClient(config)
		auth := context.WithValue(context.Background(), domeneshop.ContextBasicAuth, domeneshop.BasicAuth{
			// With the HTTP Basic Auth authentication method,
			// the token is the username, and the secret is the password.
			// See https://api.domeneshop.no/docs/#section/Authentication/basicAuth
			UserName: d.Get("token").(string),
			Password: d.Get("secret").(string),
		})

		return &apiClient{
			client: client,
			auth:   auth,
		}, nil
	}
}
