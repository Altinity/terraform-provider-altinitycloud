package env_status

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (r *AzureEnvStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: heredoc.Doc("Altinity.Cloud Azure environment status data source. It will long pool the status until `matching_spec` is `true`. Use this data source to wait for the environment is fully provisioned."),

		Attributes: map[string]schema.Attribute{
			"id":                             common.IDAttribute,
			"name":                           common.NameAttribute,
			"pending_delete":                 common.PendingDeleteAttribute,
			"applied_spec_revision":          common.AppliedSpecRevisionAttribute,
			"wait_for_applied_spec_revision": common.WaitForAppliedSpecRevisionAttribute,

			"load_balancers": schema.SingleNestedAttribute{
				Required:            false,
				Optional:            false,
				Computed:            true,
				MarkdownDescription: common.STATUS_LOAD_BALANCERS_DESCRIPTION,
				Attributes: map[string]schema.Attribute{
					"internal": schema.SingleNestedAttribute{
						Required:            false,
						Optional:            false,
						Computed:            true,
						MarkdownDescription: common.LOAD_BALANCER_INTERNAL_DESCRIPTION,
						Attributes: map[string]schema.Attribute{
							"private_link_service_alias": schema.StringAttribute{
								Required:            false,
								Optional:            false,
								Computed:            true,
								MarkdownDescription: common.AZURE_PRIVATE_LINK_SERVICE_ALIAS_DESCRIPTION,
							},
						},
					},
				},
			},
		},
	}
}
