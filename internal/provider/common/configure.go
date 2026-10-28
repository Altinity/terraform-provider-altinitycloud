package common

import (
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Reports false when the caller must stop: provider not configured yet, or wrong data type.
func SDKFromProviderData(providerData any, diags *diag.Diagnostics) (*sdk.AltinityCloudSDK, bool) {
	if providerData == nil {
		return nil, false
	}

	altinitySDK, ok := providerData.(*sdk.AltinityCloudSDK)
	if !ok {
		diags.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.AltinityCloudSDK, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil, false
	}

	return altinitySDK, true
}
