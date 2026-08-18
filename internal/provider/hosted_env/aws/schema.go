package hosted_env

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	hosted "github.com/altinity/terraform-provider-altinitycloud/internal/provider/hosted_env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/modifiers"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *HostedAWSEnvResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: heredoc.Doc(`Altinity-hosted AWS environment resource. The environment runs in an AWS account owned by Altinity.`),
		Attributes: map[string]rschema.Attribute{
			"id":                  common.IDAttribute,
			"name":                common.NameAttribute,
			"region":              common.GetRegionAttribute(true, false, false, common.AWS_REGION_DESCRIPTION),
			"cidr":                getCIDRAttribute(),
			"zone_ids":            hosted.GetZoneIDsAttribute(true, false, false, common.HOSTED_AWS_ZONE_IDS_DESCRIPTION),
			"resource_prefix":     getResourcePrefixAttribute(false, true, true),
			"kms_key_arn":         getKmsKeyArnAttribute(false, true, false),
			"custom_domains":      getCustomDomainsAttribute(false, true, false),
			"load_balancers":      getLoadBalancersAttribute(false, true, true),
			"node_groups":         hosted.GetNodeGroupsAttribute(true, false, false, common.AWS_NODE_GROUP_NODE_TYPE_DESCRIPTION),
			"maintenance_windows": common.GetMaintenanceWindowAttribute(false, true, false),
			"endpoints":           getEndpointsAttribute(false, true, false),
			"external_buckets":    getExternalBucketsAttribute(false, true, false),
			"backups":             getBackupsAttribute(false, true, false),
			"iceberg":             getIcebergAttribute(false, true, false),
			"metrics_endpoint":    common.GetMetricsEndpointAttribute(false, true, true),
			"datadog":             common.GetDatadogAttribute(false, true, false),

			"spec_revision":                   common.SpecRevisionAttribute,
			"force_destroy":                   common.GetForceDestroyAttribute(false, true, true),
			"force_destroy_clusters":          common.GetForceDestroyClustersAttribute(false, true, true),
			"skip_deprovision_on_destroy":     common.GetSkipProvisioningOnDestroyAttribute(false, true, true),
			"allow_delete_while_disconnected": common.GetAllowDeleteWhileDisconnectedAttribute(false, true, true),
		},
		Blocks: map[string]rschema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Delete: true,
			}),
		},
	}
}

func (d *HostedAWSEnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: heredoc.Doc(`Altinity-hosted AWS environment data source.`),
		Attributes: map[string]dschema.Attribute{
			"id":                  common.IDAttribute,
			"name":                common.NameAttribute,
			"region":              common.GetRegionAttribute(false, false, true, common.AWS_REGION_DESCRIPTION),
			"cidr":                getCIDRAttribute(),
			"zone_ids":            hosted.GetZoneIDsAttribute(false, false, true, common.HOSTED_AWS_ZONE_IDS_DESCRIPTION),
			"resource_prefix":     getResourcePrefixAttribute(false, false, true),
			"kms_key_arn":         getKmsKeyArnAttribute(false, false, true),
			"custom_domains":      getCustomDomainsAttribute(false, false, true),
			"load_balancers":      getLoadBalancersAttribute(false, false, true),
			"node_groups":         hosted.GetNodeGroupsAttribute(false, false, true, common.AWS_NODE_GROUP_NODE_TYPE_DESCRIPTION),
			"maintenance_windows": common.GetMaintenanceWindowAttribute(false, false, true),
			"endpoints":           getEndpointsAttribute(false, false, true),
			"external_buckets":    getExternalBucketsAttribute(false, false, true),
			"backups":             getBackupsAttribute(false, false, true),
			"iceberg":             getIcebergAttribute(false, false, true),
			"metrics_endpoint":    common.GetMetricsEndpointAttribute(false, false, true),
			"datadog":             common.GetDatadogAttribute(false, false, true),
			"spec_revision":       common.SpecRevisionAttribute,

			// these options are not used in data sources,
			// but we need to include them in the schema to avoid conversion errors.
			"force_destroy":                   common.GetForceDestroyAttribute(false, false, true),
			"force_destroy_clusters":          common.GetForceDestroyClustersAttribute(false, false, true),
			"skip_deprovision_on_destroy":     common.GetSkipProvisioningOnDestroyAttribute(false, false, true),
			"allow_delete_while_disconnected": common.GetAllowDeleteWhileDisconnectedAttribute(false, false, true),
		},
	}
}

// The API assigns the VPC CIDR; there is no create/update input for it.
func getCIDRAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:            true,
		MarkdownDescription: common.HOSTED_AWS_CIDR_DESCRIPTION,
	}
}

func getResourcePrefixAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.HOSTED_RESOURCE_PREFIX_DESCRIPTION,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("resource_prefix"),
		},
	}
}

func getKmsKeyArnAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.KMS_KEY_ARN_DESCRIPTION,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("kms_key_arn"),
		},
	}
}

func getCustomDomainsAttribute(required, optional, computed bool) rschema.ListAttribute {
	attribute := rschema.ListAttribute{
		ElementType:         types.StringType,
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.CUSTOM_DOMAINS_DESCRIPTION,
	}

	if optional || required {
		attribute.Validators = []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.ValueStringsAre(
				stringvalidator.RegexMatches(common.DomainRegex, "invalid domain format"),
			),
		}
	}

	return attribute
}

func getLoadBalancersAttribute(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.LOAD_BALANCER_DESCRIPTION,
		PlanModifiers: []planmodifier.Object{
			modifiers.DefaultObject(map[string]attr.Value{
				"public":   loadBalancerPublicDefaultObject,
				"internal": loadBalancerInternalDefaultObject,
			}),
		},
		Attributes: map[string]rschema.Attribute{
			"public": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(loadBalancerPublicDefaultObject),
				MarkdownDescription: common.LOAD_BALANCER_PUBLIC_DESCRIPTION,
				Attributes: map[string]rschema.Attribute{
					"enabled":          common.EnabledAttribute,
					"source_ip_ranges": common.SourceIPRangesAttribute,
				},
			},
			"internal": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(loadBalancerInternalDefaultObject),
				MarkdownDescription: common.LOAD_BALANCER_INTERNAL_DESCRIPTION,
				Attributes: map[string]rschema.Attribute{
					"enabled":          common.EnabledAttribute,
					"source_ip_ranges": common.SourceIPRangesAttribute,
					"endpoint_service_allowed_principals": rschema.SetAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: common.AWS_LOAD_BALANCER_ENDPOINT_SERVICE_ALLOWED_PRINCIPALS_DESCRIPTION,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
					"endpoint_service_supported_regions": rschema.SetAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: common.AWS_LOAD_BALANCER_ENDPOINT_SERVICE_SUPPORTED_REGIONS_DESCRIPTION,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
				},
			},
		},
	}
}

func getEndpointsAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.ENDPOINT_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"service_name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: common.ENDPOINT_SERVICE_NAME_DESCRIPTION,
				},
				"alias": rschema.StringAttribute{
					Optional:            true,
					MarkdownDescription: common.ENDPOINT_ALIAS_DESCRIPTION,
				},
			},
		},
	}
}

func getExternalBucketsAttribute(required, optional, computed bool) rschema.SetNestedAttribute {
	return rschema.SetNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.HOSTED_EXTERNAL_BUCKET_DESCRIPTION,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
			validators.UniqueExternalBucketNames(),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: common.EXTERNAL_BUCKET_NAME_DESCRIPTION,
				},
				"kms_key_arn": rschema.StringAttribute{
					Optional:            true,
					MarkdownDescription: common.EXTERNAL_BUCKET_KMS_KEY_ARN_DESCRIPTION,
				},
			},
		},
	}
}

func getBackupsAttribute(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.AWS_BACKUPS_DESCRIPTION,
		Attributes: map[string]rschema.Attribute{
			"custom_bucket": rschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: common.AWS_BACKUPS_CUSTOM_BUCKET_DESCRIPTION,
				Attributes: map[string]rschema.Attribute{
					"name": rschema.StringAttribute{
						Required:            true,
						MarkdownDescription: common.AWS_BACKUPS_BUCKET_DESCRIPTION,
					},
					"region": rschema.StringAttribute{
						Required:            true,
						MarkdownDescription: common.AWS_BACKUPS_REGION_DESCRIPTION,
					},
					"role_arn": rschema.StringAttribute{
						Required:            true,
						MarkdownDescription: common.AWS_BACKUPS_AUTH_DESCRIPTION,
					},
				},
			},
		},
	}
}

func getIcebergAttribute(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.ICEBERG_DESCRIPTION,
		Attributes: map[string]rschema.Attribute{
			"catalogs": rschema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: common.ICEBERG_CATALOGS_DESCRIPTION,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"name": rschema.StringAttribute{
							Optional:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_NAME_DESCRIPTION,
						},
						"type": rschema.StringAttribute{
							Required:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_TYPE_DESCRIPTION,
							Validators: []validator.String{
								stringvalidator.OneOf("S3", "S3_TABLE"),
							},
						},
						"custom_s3_bucket": rschema.StringAttribute{
							Optional:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_CUSTOM_S3_BUCKET_DESCRIPTION,
						},
						"custom_s3_bucket_path": rschema.StringAttribute{
							Optional:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_CUSTOM_S3_BUCKET_PATH_DESCRIPTION,
						},
						"custom_s3_table_bucket_arn": rschema.StringAttribute{
							Optional:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_CUSTOM_S3_TABLE_BUCKET_ARN_DESCRIPTION,
						},
						"region": rschema.StringAttribute{
							Optional:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_AWS_REGION_DESCRIPTION,
						},
						"anonymous_access_enabled": rschema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_ANONYMOUS_ACCESS_ENABLED_DESCRIPTION,
							Default:             booldefault.StaticBool(false),
						},
						"maintenance": rschema.SingleNestedAttribute{
							Optional:            true,
							Computed:            true,
							Default:             objectdefault.StaticValue(maintenanceDefaultObject),
							MarkdownDescription: common.ICEBERG_CATALOG_MAINTENANCE_DESCRIPTION,
							Attributes: map[string]rschema.Attribute{
								"enabled": rschema.BoolAttribute{
									Optional:            true,
									Computed:            true,
									MarkdownDescription: common.ICEBERG_CATALOG_MAINTENANCE_ENABLED_DESCRIPTION,
									Default:             booldefault.StaticBool(true),
								},
							},
						},
						"watches": rschema.ListNestedAttribute{
							Optional:            true,
							MarkdownDescription: common.ICEBERG_CATALOG_WATCHES_DESCRIPTION,
							NestedObject: rschema.NestedAttributeObject{
								Attributes: map[string]rschema.Attribute{
									"table": rschema.StringAttribute{
										Required:            true,
										MarkdownDescription: common.ICEBERG_CATALOG_WATCH_TABLE_DESCRIPTION,
									},
									"paths_relative_to_table_location": rschema.ListAttribute{
										ElementType:         types.StringType,
										Optional:            true,
										MarkdownDescription: common.ICEBERG_CATALOG_WATCH_PATHS_DESCRIPTION,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

var loadBalancerPublicDefaultObject, _ = types.ObjectValue(
	map[string]attr.Type{
		"enabled": types.BoolType,
		"source_ip_ranges": types.ListType{
			ElemType: types.StringType,
		},
	},
	map[string]attr.Value{
		"enabled":          types.BoolValue(false),
		"source_ip_ranges": types.ListNull(types.StringType),
	},
)

var loadBalancerInternalDefaultObject, _ = types.ObjectValue(
	map[string]attr.Type{
		"enabled": types.BoolType,
		"source_ip_ranges": types.ListType{
			ElemType: types.StringType,
		},
		"endpoint_service_allowed_principals": types.SetType{
			ElemType: types.StringType,
		},
		"endpoint_service_supported_regions": types.SetType{
			ElemType: types.StringType,
		},
	},
	map[string]attr.Value{
		"enabled":                             types.BoolValue(false),
		"source_ip_ranges":                    types.ListNull(types.StringType),
		"endpoint_service_allowed_principals": types.SetNull(types.StringType),
		"endpoint_service_supported_regions":  types.SetNull(types.StringType),
	},
)

var maintenanceDefaultObject, _ = types.ObjectValue(
	map[string]attr.Type{
		"enabled": types.BoolType,
	},
	map[string]attr.Value{
		"enabled": types.BoolValue(true),
	},
)
