package sdk

import (
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/auth"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

type AltinityCloudSDK struct {
	Client *client.Client
	Auth   *auth.Auth
}
