package certificate

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: heredoc.Doc(`Altinity.Cloud environment authentication certificate.`),

		Attributes: map[string]schema.Attribute{
			"env_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "",
			},
			"pem": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "",
			},
		},
	}
}
