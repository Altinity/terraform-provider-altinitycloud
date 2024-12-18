package env_status

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (r *HCloudEnvStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: heredoc.Doc("Altinity.Cloud HCloud environment status data source. It will long pool the status until `matching_spec` is `true`. Use this data source to wait for the environment is fully provisioned."),

		Attributes: map[string]schema.Attribute{
			"id":                             common.IDAttribute,
			"name":                           common.NameAttribute,
			"pending_delete":                 common.PendingDeleteAttribute,
			"applied_spec_revision":          common.AppliedSpecRevisionAttribute,
			"wait_for_applied_spec_revision": common.WaitForAppliedSpecRevisionAttribute,
		},
	}
}
