package secret

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: heredoc.Doc(`Altinity.Cloud secret resource.`),

		Attributes: map[string]schema.Attribute{
			"pem": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The Altinity.Cloud PEM certificate required to encrypt the value.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The value to be encrypted.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"secret_value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The encrypted value.",
			},
		},
	}
}
