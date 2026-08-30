package env

import (
	"context"
	"fmt"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &K8SEnvDataSource{}
	_ datasource.DataSourceWithConfigure = &K8SEnvDataSource{}
)

func NewK8SEnvDataSource() datasource.DataSource {
	return &K8SEnvDataSource{}
}

type K8SEnvDataSource struct {
	common.EnvDataSourceBase
}

func (d *K8SEnvDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_k8s"
}

func (d *K8SEnvDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading k8s env data source")

	var data K8SEnvDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetK8SEnv(ctx, envName)
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.K8sEnv == nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	diags = data.toModel(apiResp.K8sEnv.Name, apiResp.K8sEnv.SpecRevision, *apiResp.K8sEnv.Spec)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.CustomDomain, data.CustomDomains, diags = common.DataSourceCustomDomainsToModel(apiResp.K8sEnv.Spec.CustomDomain, apiResp.K8sEnv.Spec.CustomDomains)
	resp.Diagnostics.Append(diags...)
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
