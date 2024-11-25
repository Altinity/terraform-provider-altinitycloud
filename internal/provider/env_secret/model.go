package secret

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecretResourceModel struct {
	PEM         types.String `tfsdk:"pem"`
	Value       types.String `tfsdk:"value"`
	SecretValue types.String `tfsdk:"secret_value"`
}
