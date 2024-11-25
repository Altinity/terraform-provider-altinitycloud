package secret

import (
	"context"
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/crypto"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

type SecretResource struct {
	client *client.Client
	crypto *crypto.Crypto
}

func (r *SecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_secret"
}

func (r *SecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*sdk.AltinityCloudSDK)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = sdk.Client
	r.crypto = sdk.Crypto
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SecretResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "creating resource")
	secretValue, err := r.crypto.Encrypt(data.PEM.ValueString(), data.Value.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable encrypt, got error: %s", err))
		return
	}
	data.SecretValue = types.StringValue(secretValue)
	tflog.Trace(ctx, "created resource")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SecretResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(diags...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SecretResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updating resource")

	secretValue, err := r.crypto.Encrypt(data.PEM.ValueString(), data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable generate secret, got error: %s", err))
		return
	}
	data.SecretValue = types.StringValue(secretValue)
	tflog.Trace(ctx, "updated resource")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecretResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
