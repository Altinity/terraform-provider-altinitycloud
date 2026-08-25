package common

import (
	"regexp"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var clickHouseNameRegex = regexp.MustCompile("^[a-z0-9][a-z0-9-]{0,13}[a-z0-9]$")
var clickHouseDiskNameRegex = regexp.MustCompile("^disk[a-z0-9-]{0,12}$")

func GetClickHouseClustersAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: CLICKHOUSE_CLUSTERS_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.RegexMatches(clickHouseNameRegex, "must be 2-15 lowercase alphanumerics or hyphens, starting and ending with an alphanumeric"),
					},
				},
				"mode": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(string(client.ClickHouseClusterModeSpecStandard)),
					MarkdownDescription: CLICKHOUSE_CLUSTER_MODE_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.OneOf(
							string(client.ClickHouseClusterModeSpecStandard),
							string(client.ClickHouseClusterModeSpecSwarm),
						),
					},
				},
				"image": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_IMAGE_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"instance_type": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_INSTANCE_TYPE_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"zones": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_ZONES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
				},
				"shards": rschema.Int64Attribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_SHARDS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"replicas": rschema.Int64Attribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_REPLICAS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"stopped": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: CLICKHOUSE_CLUSTER_STOPPED_DESCRIPTION,
				},
				"disk":             getClickHouseDiskAttribute(),
				"additional_disks": getClickHouseAdditionalDisksAttribute(),
				"keeper": rschema.SingleNestedAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_KEEPER_DESCRIPTION,
					Attributes: map[string]rschema.Attribute{
						"enabled": rschema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
							MarkdownDescription: CLICKHOUSE_CLUSTER_KEEPER_ENABLED_DESCRIPTION,
						},
						"name": rschema.StringAttribute{
							Required:            true,
							MarkdownDescription: CLICKHOUSE_CLUSTER_KEEPER_NAME_DESCRIPTION,
						},
					},
				},
				"settings": getClickHouseSettingsAttribute(CLICKHOUSE_CLUSTER_SETTINGS_DESCRIPTION),
				"profiles": rschema.ListNestedAttribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_PROFILES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
					NestedObject: rschema.NestedAttributeObject{
						Attributes: map[string]rschema.Attribute{
							"name": rschema.StringAttribute{
								Required:            true,
								MarkdownDescription: CLICKHOUSE_CLUSTER_PROFILE_NAME_DESCRIPTION,
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
							},
							"settings": getClickHouseSettingsAttribute(CLICKHOUSE_CLUSTER_SETTINGS_DESCRIPTION),
						},
					},
				},
				"users": getClickHouseUsersAttribute(),
			},
		},
	}
}

func GetClickHouseKeepersAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: CLICKHOUSE_KEEPERS_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_KEEPER_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.RegexMatches(clickHouseNameRegex, "must be 2-15 lowercase alphanumerics or hyphens, starting and ending with an alphanumeric"),
					},
				},
				"instance_type": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_KEEPER_INSTANCE_TYPE_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"zones": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_KEEPER_ZONES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
				},
				"ha": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
					MarkdownDescription: CLICKHOUSE_KEEPER_HA_DESCRIPTION,
				},
				"stopped": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: CLICKHOUSE_KEEPER_STOPPED_DESCRIPTION,
				},
				"disk": getClickHouseDiskAttribute(),
			},
		},
	}
}

func getClickHouseDiskAttribute() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Required:            true,
		MarkdownDescription: CLICKHOUSE_DISK_DESCRIPTION,
		Attributes: map[string]rschema.Attribute{
			"size":          getClickHouseDiskSizeAttribute(),
			"storage_class": getClickHouseDiskStorageClassAttribute(),
			"iops":          getClickHouseDiskIopsAttribute(),
			"throughput":    getClickHouseDiskThroughputAttribute(),
		},
	}
}

func getClickHouseAdditionalDisksAttribute() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_CLUSTER_ADDITIONAL_DISKS_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(8),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_DISK_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.RegexMatches(clickHouseDiskNameRegex, "must start with `disk` and cannot exceed 16 characters"),
					},
				},
				"size":          getClickHouseDiskSizeAttribute(),
				"storage_class": getClickHouseDiskStorageClassAttribute(),
				"iops":          getClickHouseDiskIopsAttribute(),
				"throughput":    getClickHouseDiskThroughputAttribute(),
			},
		},
	}
}

func getClickHouseDiskSizeAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Required:            true,
		MarkdownDescription: CLICKHOUSE_DISK_SIZE_DESCRIPTION,
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
	}
}

func getClickHouseDiskStorageClassAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: CLICKHOUSE_DISK_STORAGE_CLASS_DESCRIPTION,
	}
}

func getClickHouseDiskIopsAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: CLICKHOUSE_DISK_IOPS_DESCRIPTION,
	}
}

func getClickHouseDiskThroughputAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: CLICKHOUSE_DISK_THROUGHPUT_DESCRIPTION,
	}
}

func getClickHouseSettingsAttribute(description string) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: description,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"key": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_SETTING_KEY_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"value": rschema.StringAttribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_SETTING_VALUE_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_from_secret")),
					},
				},
				"value_from_secret": getClickHouseSecretRefAttribute(CLICKHOUSE_SETTING_VALUE_FROM_SECRET_DESCRIPTION),
			},
		},
	}
}

func getClickHouseUsersAttribute() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_CLUSTER_USERS_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_USER_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.NoneOf("grafana", "datadog"),
					},
				},
				"profile": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_USER_PROFILE_DESCRIPTION,
				},
				"quota": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_USER_QUOTA_DESCRIPTION,
				},
				"allowed_cidrs": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_USER_ALLOWED_CIDRS_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
				},
				"databases": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_USER_DATABASES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
				},
				"access_management": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: CLICKHOUSE_USER_ACCESS_MANAGEMENT_DESCRIPTION,
				},
				"named_collection_control": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: CLICKHOUSE_USER_NAMED_COLLECTION_CONTROL_DESCRIPTION,
				},
				"show_named_collections": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: CLICKHOUSE_USER_SHOW_NAMED_COLLECTIONS_DESCRIPTION,
				},
				"show_named_collections_secrets": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: CLICKHOUSE_USER_SHOW_NAMED_COLLECTIONS_SECRETS_DESCRIPTION,
				},
				"password_type": rschema.StringAttribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_USER_PASSWORD_TYPE_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.OneOf(
							string(client.ClickHouseUserPasswordTypeSpecInputSha256Hex),
							string(client.ClickHouseUserPasswordTypeSpecInputDoubleSha1Hex),
						),
					},
				},
				"password_value": rschema.StringAttribute{
					Optional:            true,
					Sensitive:           true,
					MarkdownDescription: CLICKHOUSE_USER_PASSWORD_VALUE_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_type")),
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_value_from_secret")),
					},
				},
				"password_value_from_secret": getClickHouseSecretRefAttribute(CLICKHOUSE_USER_PASSWORD_VALUE_FROM_SECRET_DESCRIPTION),
			},
		},
	}
}

func getClickHouseSecretRefAttribute(description string) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: description,
		Attributes: map[string]rschema.Attribute{
			"name": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: CLICKHOUSE_SECRET_REF_NAME_DESCRIPTION,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"key": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: CLICKHOUSE_SECRET_REF_KEY_DESCRIPTION,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}
