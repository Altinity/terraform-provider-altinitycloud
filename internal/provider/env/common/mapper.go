package env

import (
	"context"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// SetToModel/ListToModel always return non-null; unsafe for plain Optional attrs where null must stay null.
func SetToModel(input []string) (types.Set, diag.Diagnostics) {
	zones := []attr.Value{}
	for _, str := range input {
		zones = append(zones, types.StringValue(str))
	}

	list, diags := types.SetValue(types.StringType, zones)
	return list, diags
}

func ListToModel(input []string) (types.List, diag.Diagnostics) {
	zones := []attr.Value{}
	for _, str := range input {
		zones = append(zones, types.StringValue(str))
	}

	list, diags := types.ListValue(types.StringType, zones)
	return list, diags
}

// Prefers custom_domains over the deprecated custom_domain (mutually exclusive via validator); unknown-safe.
func CustomDomainsToSDK(ctx context.Context, customDomain types.String, customDomains types.List) (*string, []string, diag.Diagnostics) {
	var allDiags diag.Diagnostics

	if !customDomains.IsUnknown() && !customDomains.IsNull() {
		var domains []string
		allDiags.Append(customDomains.ElementsAs(ctx, &domains, false)...)
		return nil, domains, allDiags
	}

	if customDomain.IsUnknown() || customDomain.IsNull() {
		return nil, nil, allDiags
	}

	return customDomain.ValueStringPointer(), nil, allDiags
}

// Refreshes whichever field prior state manages and keeps the other null, so the
// API's customDomains[0] echo can't flip a list-managed resource into a permanent
// diff on the deprecated scalar. Resources only; data sources expose both fields.
func CustomDomainsToModel(priorCustomDomains types.List, specCustomDomain *string, specCustomDomains []string) (types.String, types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !priorCustomDomains.IsNull() {
		list, d := ListToModel(specCustomDomains)
		diags.Append(d...)
		return types.StringNull(), list, diags
	}

	return types.StringPointerValue(specCustomDomain), types.ListNull(types.StringType), diags
}

// Data sources are read-only with no user intent to preserve, so expose both attributes as returned.
func DataSourceCustomDomainsToModel(specCustomDomain *string, specCustomDomains []string) (types.String, types.List, diag.Diagnostics) {
	list, diags := ListToModel(specCustomDomains)
	return types.StringPointerValue(specCustomDomain), list, diags
}

// Always non-null; safe only while reservations attrs are Required or defaulted.
func ReservationsToModel(input []client.NodeReservation) (types.Set, diag.Diagnostics) {
	reservations := []attr.Value{}
	for _, reservation := range input {
		reservations = append(reservations, types.StringValue(string(reservation)))
	}

	list, diags := types.SetValue(types.StringType, reservations)
	return list, diags
}

func ListStringToSDK(input []basetypes.StringValue) []string {
	var list []string
	for _, str := range input {
		list = append(list, str.ValueString())
	}

	return list
}

func ListStringToModel(input []string) []types.String {
	var list []types.String
	for _, str := range input {
		list = append(list, types.StringValue(str))
	}

	return list
}

func KeyValueToSDK(input []KeyValueModel) []*client.KeyValueInput {
	var list []*client.KeyValueInput
	for _, element := range input {
		list = append(list, &client.KeyValueInput{
			Key:   element.Key.ValueString(),
			Value: element.Value.ValueString(),
		})
	}

	return list
}

func DatadogToSDK(datadog *DatadogModel) *client.DatadogSpecInput {
	if datadog == nil {
		return nil
	}

	return &client.DatadogSpecInput{
		Enabled:        datadog.Enabled.ValueBoolPointer(),
		EncAPIKey:      datadog.EncAPIKey.ValueStringPointer(),
		Domain:         datadog.Domain.ValueStringPointer(),
		LogsEnabled:    datadog.LogsEnabled.ValueBoolPointer(),
		MetricsEnabled: datadog.MetricsEnabled.ValueBoolPointer(),
	}
}

func MaintenanceWindowsToSDK(maintenanceWindows []MaintenanceWindowModel) []*client.MaintenanceWindowSpecInput {
	var sdkMaintenanceWindows []*client.MaintenanceWindowSpecInput
	for _, mw := range maintenanceWindows {
		var days []client.Day
		for _, day := range mw.Days {
			days = append(days, client.Day(day.ValueString()))
		}

		sdkMaintenanceWindows = append(sdkMaintenanceWindows, &client.MaintenanceWindowSpecInput{
			Name:          mw.Name.ValueString(),
			Enabled:       mw.Enabled.ValueBoolPointer(),
			Hour:          mw.Hour.ValueInt64(),
			LengthInHours: mw.LengthInHours.ValueInt64(),
			Days:          days,
		})
	}

	return sdkMaintenanceWindows
}
