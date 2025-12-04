package env_status

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *AWSEnvStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: heredoc.Doc("Altinity.Cloud AWS environment status data source. It will long pool the status until `matching_spec` is `true`. Use this data source to wait for the environment is fully provisioned."),

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
							"endpoint_service_name": schema.StringAttribute{
								Required:            false,
								Optional:            false,
								Computed:            true,
								MarkdownDescription: common.STATUS_LOAD_BALANCERS_ENDPOINT_SERVICE_NAME_DESCRIPTION,
							},
						},
					},
				},
			},
			"peering_connections": rschema.ListNestedAttribute{
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"id": rschema.StringAttribute{
							Optional:            true,
							MarkdownDescription: common.PEERING_CONNECTION_ID_DESCRIPTION,
						},
						"vpc_id": rschema.StringAttribute{
							Required:            true,
							MarkdownDescription: common.PEERING_CONNECTION_VPC_ID_DESCRIPTION,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: common.PEERING_CONNECTION_DESCRIPTION,
			},
			"aws_resources": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: common.AWS_RESOURCES_ID_DESCRIPTION,
						},
						"arn": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: common.AWS_RESOURCES_ARN_DESCRIPTION,
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: common.AWS_RESOURCES_NAME_DESCRIPTION,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: common.AWS_RESOURCES_DESCRIPTION,
			},
		},
	}
}
