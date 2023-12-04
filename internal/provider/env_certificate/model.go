package certificate

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CertificateResourceModel struct {
	EnvironmentName types.String `tfsdk:"env_name"`
	PEM             types.String `tfsdk:"pem"`
}
