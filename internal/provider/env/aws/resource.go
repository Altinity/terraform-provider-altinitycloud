package env

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &AWSEnvResource{}
var _ resource.ResourceWithConfigure = &AWSEnvResource{}
var _ resource.ResourceWithImportState = &AWSEnvResource{}
var _ resource.ResourceWithModifyPlan = &AWSEnvResource{}
var _ resource.ResourceWithValidateConfig = &AWSEnvResource{}

const adoptClickHouseKey = "clickhouse_adopt"

func NewAWSEnvResource() resource.Resource {
	return &AWSEnvResource{}
}

type AWSEnvResource struct {
	common.EnvResourceBase
}

func (r *AWSEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_aws"
}

// Import is the one read that adopts clusters and Keepers Terraform did not create.
func (r *AWSEnvResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.EnvResourceBase.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, adoptClickHouseKey, []byte("1"))...)
}

func (r *AWSEnvResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.EnvResourceBase.ModifyPlan(ctx, req, resp)
	if req.Plan.Raw.IsNull() || resp.Diagnostics.HasError() {
		return
	}

	plannedClusters, priorClusters, diags := planClickHouse[common.ClickHouseClusterModel](ctx, req, resp, "clickhouse_clusters", clientsupport.GetClickHouseClustersAttribute(false, true, false))
	resp.Diagnostics.Append(diags...)
	plannedKeepers, priorKeepers, diags := planClickHouse[common.ClickHouseKeeperModel](ctx, req, resp, "clickhouse_keepers", clientsupport.GetClickHouseKeepersAttribute(false, true, false))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(common.ValidateClickHouseNewUserPasswords(path.Root("clickhouse_clusters"), priorClusters, plannedClusters)...)
	if req.State.Raw.IsNull() {
		return
	}

	var envName types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("name"), &envName)...)
	resp.Diagnostics.Append(r.rejectClickHouseCollisions(ctx, envName.ValueString(),
		common.ClickHouseClusterNamesAdded(priorClusters, plannedClusters),
		common.ClickHouseKeeperNamesAdded(priorKeepers, plannedKeepers),
	)...)
}

// Computed values the config omits keep their state value, paired by name rather than list index.
// schema is the same constructor schema.go uses, so the computed/default flags match the real schema.
func planClickHouse[M any](ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, attribute string, schema rschema.ListNestedAttribute) (planned, prior []M, diags diag.Diagnostics) {
	p := path.Root(attribute)

	var plan, state, config types.List
	diags.Append(req.Plan.GetAttribute(ctx, p, &plan)...)
	diags.Append(req.Config.GetAttribute(ctx, p, &config)...)
	if !req.State.Raw.IsNull() {
		diags.Append(req.State.GetAttribute(ctx, p, &state)...)
	}
	if diags.HasError() {
		return nil, nil, diags
	}

	filled := common.UseStateForUnknownByKeyList(ctx, attribute, schema, plan, state, config)
	diags.Append(resp.Plan.SetAttribute(ctx, p, filled)...)

	var d diag.Diagnostics
	planned, d = common.ObjectsFromList[M](ctx, filled)
	diags.Append(d...)
	prior, d = common.ObjectsFromList[M](ctx, state)
	diags.Append(d...)
	return planned, prior, diags
}

func (r *AWSEnvResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Read only datadog: a full Config.Get panics on unknown nested struct-pointer attrs.
	var datadogObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("datadog"), &datadogObj)...)
	if resp.Diagnostics.HasError() || datadogObj.IsNull() || datadogObj.IsUnknown() {
		return
	}

	var datadog *common.DatadogModel
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("datadog"), &datadog)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(common.ValidateDatadog(datadog)...)
}

func (r *AWSEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AWSEnvResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "creating resource", map[string]interface{}{"name": envName})

	sdkEnv, _, diags := data.toSDK(ctx, nil)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiResp, err := r.Client.CreateAWSEnv(ctx, sdkEnv)

	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to create env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	// Reorder node groups, zones and tags to respect order in the user's configuration
	apiResp.CreateAWSEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.CreateAWSEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AWSEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	resp.Diagnostics.Append(reorderNodeGroupZones(ctx, data.NodeGroups, apiResp.CreateAWSEnv.Spec.NodeGroups)...)
	apiResp.CreateAWSEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.CreateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.CreateAWSEnv.Spec.Tags = common.ReorderByKey(data.Tags, apiResp.CreateAWSEnv.Spec.Tags,
		func(m common.KeyValueModel) string { return m.Key.ValueString() },
		func(s *client.AWSEnvSpecFragment_Tags) string { return s.Key },
	)
	apiResp.CreateAWSEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.CreateAWSEnv.Spec.PeeringConnections,
		func(m AWSEnvPeeringConnectionModel) string { return m.VpcID.ValueString() },
		func(s *client.AWSEnvSpecFragment_PeeringConnections) string { return s.VpcID },
	)
	data.Id = data.Name
	data.Zones, diags = common.ListToModel(apiResp.CreateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.CreateAWSEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.CreateAWSEnv.SpecRevision)
	data.ResourcePrefix = types.StringValue(apiResp.CreateAWSEnv.Spec.ResourcePrefix)
	data.ClickHouseClusters = common.ClickHouseClustersToModel(data.ClickHouseClusters, clickHouseClustersToSpec(apiResp.CreateAWSEnv.Spec.ClickHouseClusters))
	data.ClickHouseKeepers = common.ClickHouseKeepersToModel(data.ClickHouseKeepers, clickHouseKeepersToSpec(apiResp.CreateAWSEnv.Spec.ClickHouseKeepers))

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": envName})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AWSEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AWSEnvResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "getting environment", map[string]interface{}{"name": envName})
	apiResp, err := r.Client.GetAWSEnv(ctx, envName)

	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "removing resource from state", map[string]interface{}{"name": envName})
			resp.State.RemoveResource(ctx)
		} else {
			clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName)))
		}
		return
	}

	if apiResp.AWSEnv == nil {
		tflog.Trace(ctx, "removing resource from state", map[string]interface{}{"name": envName})
		resp.State.RemoveResource(ctx)
		return
	}

	// Reorder node groups, zones and tags to respect order in the user's configuration
	apiResp.AWSEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.AWSEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AWSEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	resp.Diagnostics.Append(reorderNodeGroupZones(ctx, data.NodeGroups, apiResp.AWSEnv.Spec.NodeGroups)...)
	apiResp.AWSEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.AWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.AWSEnv.Spec.Tags = common.ReorderByKey(data.Tags, apiResp.AWSEnv.Spec.Tags,
		func(m common.KeyValueModel) string { return m.Key.ValueString() },
		func(s *client.AWSEnvSpecFragment_Tags) string { return s.Key },
	)
	apiResp.AWSEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.AWSEnv.Spec.PeeringConnections,
		func(m AWSEnvPeeringConnectionModel) string { return m.VpcID.ValueString() },
		func(s *client.AWSEnvSpecFragment_PeeringConnections) string { return s.VpcID },
	)
	diags = data.toModel(*apiResp.AWSEnv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	adopt, diags := req.Private.GetKey(ctx, adoptClickHouseKey)
	resp.Diagnostics.Append(diags...)
	if len(adopt) > 0 {
		data.ClickHouseClusters = common.AdoptClickHouseClusters(clickHouseClustersToSpec(apiResp.AWSEnv.Spec.ClickHouseClusters))
		data.ClickHouseKeepers = common.AdoptClickHouseKeepers(clickHouseKeepersToSpec(apiResp.AWSEnv.Spec.ClickHouseKeepers))
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, adoptClickHouseKey, nil)...)
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AWSEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AWSEnvResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *AWSEnvResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "updating resource", map[string]interface{}{"name": envName})

	_, sdkEnv, diags := data.toSDK(ctx, &state.AWSEnvModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiResp, err := r.Client.UpdateAWSEnv(ctx, sdkEnv)

	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to update env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	// Reorder node groups, zones and tags to respect order in the user's configuration
	apiResp.UpdateAWSEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.UpdateAWSEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AWSEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	resp.Diagnostics.Append(reorderNodeGroupZones(ctx, data.NodeGroups, apiResp.UpdateAWSEnv.Spec.NodeGroups)...)
	apiResp.UpdateAWSEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.UpdateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.UpdateAWSEnv.Spec.Tags = common.ReorderByKey(data.Tags, apiResp.UpdateAWSEnv.Spec.Tags,
		func(m common.KeyValueModel) string { return m.Key.ValueString() },
		func(s *client.AWSEnvSpecFragment_Tags) string { return s.Key },
	)
	apiResp.UpdateAWSEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.UpdateAWSEnv.Spec.PeeringConnections,
		func(m AWSEnvPeeringConnectionModel) string { return m.VpcID.ValueString() },
		func(s *client.AWSEnvSpecFragment_PeeringConnections) string { return s.VpcID },
	)
	data.Zones, diags = common.ListToModel(apiResp.UpdateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.UpdateAWSEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.UpdateAWSEnv.SpecRevision)
	data.ResourcePrefix = types.StringValue(apiResp.UpdateAWSEnv.Spec.ResourcePrefix)
	data.ClickHouseClusters = common.ClickHouseClustersToModel(data.ClickHouseClusters, clickHouseClustersToSpec(apiResp.UpdateAWSEnv.Spec.ClickHouseClusters))
	data.ClickHouseKeepers = common.ClickHouseKeepersToModel(data.ClickHouseKeepers, clickHouseKeepersToSpec(apiResp.UpdateAWSEnv.Spec.ClickHouseKeepers))

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": envName})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Patch Replace upserts by name, so a "pure create" plan can overwrite an unmanaged entry.
func (r *AWSEnvResource) rejectClickHouseCollisions(ctx context.Context, envName string, addedClusters, addedKeepers []string) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(addedClusters) == 0 && len(addedKeepers) == 0 {
		return diags
	}

	current, err := r.Client.GetAWSEnv(ctx, envName)
	if err != nil {
		clientsupport.AddClientError(&diags, fmt.Sprintf("Unable to read env %s before adding ClickHouse entries, got error: %s", envName, client.FormatError(err, envName)))
		return diags
	}
	if current.AWSEnv == nil {
		return diags
	}

	var clusterNames, keeperNames []string
	for _, c := range current.AWSEnv.Spec.ClickHouseClusters {
		if c != nil {
			clusterNames = append(clusterNames, c.Name)
		}
	}
	for _, k := range current.AWSEnv.Spec.ClickHouseKeepers {
		if k != nil {
			keeperNames = append(keeperNames, k.Name)
		}
	}

	for _, name := range common.ClickHouseCollisions(addedClusters, clusterNames) {
		diags.AddAttributeError(path.Root("clickhouse_clusters"), "ClickHouse Cluster Already Exists",
			fmt.Sprintf("Cluster %q already exists in env %s but is not in the Terraform state, so applying would overwrite its configuration. Rename the new cluster, remove the unmanaged one first, or adopt every existing cluster and Keeper by re-importing the environment (`terraform state rm`, then `terraform import`).", name, envName))
	}
	for _, name := range common.ClickHouseCollisions(addedKeepers, keeperNames) {
		diags.AddAttributeError(path.Root("clickhouse_keepers"), "ClickHouse Keeper Already Exists",
			fmt.Sprintf("Keeper %q already exists in env %s but is not in the Terraform state, so applying would overwrite its configuration. Rename the new Keeper, remove the unmanaged one first, or adopt every existing cluster and Keeper by re-importing the environment (`terraform state rm`, then `terraform import`).", name, envName))
	}

	return diags
}

func (r *AWSEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AWSEnvResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	envStatus, err := r.Client.GetAWSEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env status %s, got error: %s", envName, err))
		}
		return
	}

	if envStatus.AWSEnv == nil {
		tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		return
	}

	resp.Diagnostics.Append(common.ValidateForceDestroy(envName, data.ForceDestroy.ValueBool())...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(envStatus.AWSEnv.Status.Errors) > 0 {
		for _, err := range envStatus.AWSEnv.Status.Errors {
			resp.Diagnostics.Append(common.ValidateDisconnected(
				envName,
				string(err.Code),
				envStatus.AWSEnv.Status.AppliedSpecRevision,
				data.SkipDeprovisionOnDestroy.ValueBool(),
				data.AllowDeleteWhileDisconnected.ValueBool(),
			)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	apiResp, err := r.Client.DeleteAWSEnv(ctx, client.DeleteAWSEnvInput{
		Name:                 envName,
		Force:                data.SkipDeprovisionOnDestroy.ValueBoolPointer(),
		ForceDestroyClusters: data.ForceDestroyClusters.ValueBoolPointer(),
	})

	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, common.FormatDeleteError(envName, err))
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, common.DeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	common.WaitForDeletion(ctx, resp, envName, apiResp.DeleteAWSEnv.PendingMfa,
		func(ctx context.Context, name string) (bool, error) {
			status, err := r.Client.GetAWSEnvStatus(ctx, name)
			if err != nil {
				return false, err
			}
			if status.AWSEnv == nil {
				return false, common.ErrEnvNotFound
			}
			return status.AWSEnv.Status.PendingDelete, nil
		},
		deleteTimeout,
		common.MFATimeout,
	)
}
