package secret

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: heredoc.Doc(`Altinity.Cloud secret resource.`),

		Attributes: map[string]schema.Attribute{
			"pem": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "",
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "",
			},
			"secret_value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "",
			},
		},
	}
}
