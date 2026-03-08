package env

import (
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
