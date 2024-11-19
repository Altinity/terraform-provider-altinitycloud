package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	env_aws "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/aws"
	env_azure "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/azure"
	env_gcp "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/gcp"
	env_hcloud "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/hcloud"
	env_k8s "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/k8s"
	env_certificate "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_certificate"
	env_status_aws "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/aws"
	env_status_azure "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/azure"
	env_status_gcp "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/gcp"
	env_status_k8s "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/k8s"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/secret"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/auth"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/crypto"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const DEFAULT_API_URL = "https://anywhere.altinity.cloud"
const GRAPHQL_API_PATH = "/api/v1/graphql"

const ENV_VAR_API_URL = "ALTINITYCLOUD_API_URL"
const ENV_VAR_API_TOKEN = "ALTINITYCLOUD_API_TOKEN"

var _ provider.Provider = &altinityCloudProvider{}

// altinityCloudProvider defines the provider implementation.
type altinityCloudProvider struct {
	version string
}

// altinityCloudProviderModel describes the provider data model.
type altinityCloudProviderModel struct {
	ApiURL   types.String `tfsdk:"api_url"`
	ApiToken types.String `tfsdk:"api_token"`
	CACrt    types.String `tfsdk:"ca_crt"`
}

func (p *altinityCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "altinitycloud"
	resp.Version = p.version
}

func (p *altinityCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Altinity.Cloud API URL. Defaults to `%s` unless `%s` env var is set.",
					DEFAULT_API_URL, ENV_VAR_API_URL),
				Optional: true,
			},
			"ca_crt": schema.StringAttribute{
				MarkdownDescription: "CA bundle for Altinity.Cloud.",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Altinity.Cloud API Token.\n"+
					"The value can be omitted if `%s` environment variable is set. ",
					ENV_VAR_API_TOKEN),
				Optional: true,
			},
		},
	}
}

func (p *altinityCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data altinityCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiToken := os.Getenv(ENV_VAR_API_TOKEN)
	apiUrl := os.Getenv(ENV_VAR_API_URL)
	caCrt := data.CACrt.ValueStringPointer()

	// Overwrite env variables with TF config values
	if !data.ApiToken.IsNull() {
		apiToken = data.ApiToken.ValueString()
	}

	if !data.ApiURL.IsNull() {
		apiUrl = data.ApiURL.ValueString()
	}

	// Use default value for API URL if is not set
	if apiUrl == "" {
		apiUrl = DEFAULT_API_URL
	}

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Altinity.Cloud API Token",
			fmt.Sprintf("%s environment variable or \"api_token\" provider attribute required.\n"+
				"See https://github.com/altinity/terraform-provider-altinitycloud for details.", ENV_VAR_API_TOKEN),
		)
	}

	var rootCAs *x509.CertPool
	if caCrt != nil {
		var err error
		rootCAs, err = loadCertPool(*caCrt)
		if err != nil {
			resp.Diagnostics.AddWarning("Failed to load CA certificate", err.Error())
		}
	}

	client := client.NewClient(
		&http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          16,
				IdleConnTimeout:       45 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig: &tls.Config{
					RootCAs: rootCAs,
				},
			},
			Timeout: time.Second * 60,
		},
		apiUrl+GRAPHQL_API_PATH,
		nil,
		client.WithBearerAuthorization(ctx, apiToken),
		client.WithUserAgent(ctx, userAgent(p.version)),
	)

	auth := auth.NewAuth(rootCAs, apiUrl, apiToken)
	crypto := crypto.NewCrypto(rootCAs, apiUrl)
	sdk := &sdk.AltinityCloudSDK{
		Client: client,
		Auth:   auth,
		Crypto: crypto,
	}

	resp.DataSourceData = sdk
	resp.ResourceData = sdk
}

func (p *altinityCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		env_aws.NewAWSEnvResource,
		env_gcp.NewGCPEnvResource,
		env_azure.NewAzureEnvResource,
		env_hcloud.NewHCloudEnvResource,
		env_k8s.NewK8SEnvResource,
		env_certificate.NewCertificateResource,
		secret.NewSecretResource,
	}
}

func (p *altinityCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		env_aws.NewAWSEnvDataSource,
		env_gcp.NewGCPEnvDataSource,
		env_azure.NewAzureEnvDataSource,
		env_hcloud.NewHCloudEnvDataSource,
		env_k8s.NewK8SEnvDataSource,

		env_status_azure.NewAzureEnvStatusDataSource,
		env_status_aws.NewAWSEnvStatusDataSource,
		env_status_gcp.NewGCPEnvStatusDataSource,
		env_status_k8s.NewK8SEnvStatusDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &altinityCloudProvider{
			version: version,
		}
	}
}

func loadCertPool(cert string) (*x509.CertPool, error) {
	clientCAs := x509.NewCertPool()
	ok := clientCAs.AppendCertsFromPEM([]byte(cert))
	if !ok {
		return nil, errors.New("pem: failed to append certificates")
	}
	return clientCAs, nil
}

func userAgent(version string) string {
	return fmt.Sprintf("%s/%s (%s; %s) go/%s",
		"terraform",
		version,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)
}
