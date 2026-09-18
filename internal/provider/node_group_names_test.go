package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const duplicateNodeGroupNameSummary = "Duplicate node group name"

func nodeGroupValue(t *testing.T, elementType tftypes.Type, name, nodeType string) tftypes.Value {
	t.Helper()
	objectType, ok := elementType.(tftypes.Object)
	if !ok {
		t.Fatalf("node_groups element type is %T, want tftypes.Object", elementType)
	}

	fields := map[string]tftypes.Value{}
	for attribute, attributeType := range objectType.AttributeTypes {
		switch attribute {
		case "name":
			fields[attribute] = tftypes.NewValue(attributeType, name)
		case "node_type":
			fields[attribute] = tftypes.NewValue(attributeType, nodeType)
		default:
			fields[attribute] = tftypes.NewValue(attributeType, nil)
		}
	}

	return tftypes.NewValue(objectType, fields)
}

func configWithNodeGroups(t *testing.T, objectType tftypes.Object, groups ...tftypes.Value) tftypes.Value {
	t.Helper()
	attributes := map[string]tftypes.Value{}
	for attribute, attributeType := range objectType.AttributeTypes {
		attributes[attribute] = tftypes.NewValue(attributeType, nil)
	}
	attributes["node_groups"] = tftypes.NewValue(objectType.AttributeTypes["node_groups"], groups)

	return tftypes.NewValue(objectType, attributes)
}

// node_groups is a list on most envs and a set on hcloud.
func nodeGroupElementType(t *testing.T, attributeType tftypes.Type) tftypes.Type {
	t.Helper()
	switch collection := attributeType.(type) {
	case tftypes.List:
		return collection.ElementType
	case tftypes.Set:
		return collection.ElementType
	}

	t.Fatalf("node_groups is %T, want tftypes.List or tftypes.Set", attributeType)
	return nil
}

func hasDuplicateNameError(resp *resource.ValidateConfigResponse) bool {
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), duplicateNodeGroupNameSummary) {
			return true
		}
	}

	return false
}

// Guards the wiring: a resource exposing node_groups that forgets ValidateNodeGroupNames fails here.
func TestResourcesRejectDuplicateNodeGroupNames(t *testing.T) {
	ctx := context.Background()

	for _, newResource := range (&altinityCloudProvider{}).Resources(ctx) {
		res := newResource()

		metadataResp := &resource.MetadataResponse{}
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "altinitycloud"}, metadataResp)

		schemaResp := &resource.SchemaResponse{}
		res.Schema(ctx, resource.SchemaRequest{}, schemaResp)
		objectType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
		if !ok {
			t.Fatalf("%s schema type is %T, want tftypes.Object", metadataResp.TypeName, schemaResp.Schema.Type().TerraformType(ctx))
		}
		if _, ok := objectType.AttributeTypes["node_groups"]; !ok {
			continue
		}

		t.Run(metadataResp.TypeName, func(t *testing.T) {
			validating, ok := res.(resource.ResourceWithValidateConfig)
			if !ok {
				t.Fatalf("%s exposes node_groups but does not implement ValidateConfig", metadataResp.TypeName)
			}
			elementType := nodeGroupElementType(t, objectType.AttributeTypes["node_groups"])

			validate := func(groups ...tftypes.Value) *resource.ValidateConfigResponse {
				cfg := tfsdk.Config{Schema: schemaResp.Schema, Raw: configWithNodeGroups(t, objectType, groups...)}
				resp := &resource.ValidateConfigResponse{}
				validating.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
				return resp
			}

			resp := validate(
				nodeGroupValue(t, elementType, "workers", "m6i.large"),
				nodeGroupValue(t, elementType, "workers", "m6i.xlarge"),
			)
			if !hasDuplicateNameError(resp) {
				t.Errorf("duplicate names accepted, diagnostics: %v", resp.Diagnostics)
			}

			resp = validate(
				nodeGroupValue(t, elementType, "workers-a", "m6i.large"),
				nodeGroupValue(t, elementType, "workers-b", "m6i.large"),
			)
			if hasDuplicateNameError(resp) {
				t.Errorf("distinct names rejected, diagnostics: %v", resp.Diagnostics)
			}
		})
	}
}
