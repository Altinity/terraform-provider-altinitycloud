package sdk

import (
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/auth"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/crypto"
)

type AltinityCloudSDK struct {
	Client *client.Client
	Auth   *auth.Auth
	Crypto *crypto.Crypto
}
