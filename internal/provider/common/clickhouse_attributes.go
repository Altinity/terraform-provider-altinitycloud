package common

import (
	"fmt"
	"regexp"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/validators"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Platform-injected: the API rejects them on input and can return them in a read.
var ReservedClickHouseUsers = []string{"grafana", "datadog", "default", "clickhouse_operator", "swarm_access"}

var clickHouseNameRegex = regexp.MustCompile("^[a-z0-9][a-z0-9-]{0,13}[a-z0-9]$")

var clickHouseNameValidator = stringvalidator.RegexMatches(clickHouseNameRegex, "must be 2-15 lowercase alphanumerics or hyphens, starting and ending with an alphanumeric")

var clickHouseExtraDiskNameRegex = regexp.MustCompile("^disk")

// Every env spells the password type enum the same, so the shared attribute keys off AWS's.
var clickHousePasswordTypes = []string{
	string(client.AWSEnvClickHouseUserPasswordTypeSha256Hex),
	string(client.AWSEnvClickHouseUserPasswordTypeDoubleSha1Hex),
}

var clickHouseQuotaKeyedBy = []string{
	string(client.ClickHouseQuotaKeyedByNone),
	string(client.ClickHouseQuotaKeyedByClientKey),
	string(client.ClickHouseQuotaKeyedByIPAddress),
	string(client.ClickHouseQuotaKeyedByForwardedIPAddress),
	string(client.ClickHouseQuotaKeyedByNormalizedQueryHash),
}

// The API upserts by name, so a repeated name would silently keep only the last entry.
func uniqueClickHouseNames(attribute, entry string) validator.List {
	return validators.UniqueListObjectAttribute(attribute, fmt.Sprintf("Duplicate ClickHouse %s", entry),
		func(value string) string {
			return fmt.Sprintf("More than one %s is named %q. Names must be unique within the list.", entry, value)
		})
}

func GetClickHouseClustersAttribute(required, optional, computed bool) rschema.MapNestedAttribute {
	return rschema.MapNestedAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: CLICKHOUSE_CLUSTERS_DESCRIPTION,
		Validators: []validator.Map{
			mapvalidator.SizeAtLeast(1),
			mapvalidator.KeysAre(clickHouseNameValidator),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"image": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
					PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_CLUSTER_ZONES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.UniqueValues(),
					},
				},
				"shards": rschema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_CLUSTER_SHARDS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"replicas": rschema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
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
				"keeper": rschema.StringAttribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_KEEPER_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"disk":            getClickHouseDiskAttribute(),
				"extra_disks":     getClickHouseExtraDisksAttribute(),
				"users":           getClickHouseUsersAttribute(),
				"profiles":        getClickHouseProfilesAttribute(),
				"quotas":          getClickHouseQuotasAttribute(),
				"server_settings": getClickHouseSettingsAttribute(CLICKHOUSE_CLUSTER_SERVER_SETTINGS_DESCRIPTION),
			},
		},
	}
}

func GetClickHouseKeepersAttribute(required, optional, computed bool) rschema.MapNestedAttribute {
	return rschema.MapNestedAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: CLICKHOUSE_KEEPERS_DESCRIPTION,
		Validators: []validator.Map{
			mapvalidator.SizeAtLeast(1),
			mapvalidator.KeysAre(clickHouseNameValidator),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
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
					PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_KEEPER_ZONES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.UniqueValues(),
					},
				},
				"replicas": rschema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_KEEPER_REPLICAS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.OneOf(1, 3),
					},
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
			"size_in_gib":         getClickHouseDiskSizeAttribute(),
			"iops":                getClickHouseDiskIopsAttribute(),
			"throughput_in_mibps": getClickHouseDiskThroughputAttribute(),
		},
	}
}

func getClickHouseExtraDisksAttribute() rschema.MapNestedAttribute {
	return rschema.MapNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_EXTRA_DISKS_DESCRIPTION,
		Validators: []validator.Map{
			mapvalidator.SizeAtMost(8),
			mapvalidator.KeysAre(
				stringvalidator.LengthBetween(1, 16),
				stringvalidator.RegexMatches(clickHouseExtraDiskNameRegex, "must start with `disk`"),
			),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"size_in_gib":         getClickHouseDiskSizeAttribute(),
				"iops":                getClickHouseDiskIopsAttribute(),
				"throughput_in_mibps": getClickHouseDiskThroughputAttribute(),
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

func getClickHouseDiskIopsAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		MarkdownDescription: CLICKHOUSE_DISK_IOPS_DESCRIPTION,
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
	}
}

func getClickHouseDiskThroughputAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		MarkdownDescription: CLICKHOUSE_DISK_THROUGHPUT_DESCRIPTION,
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
	}
}

func getClickHouseSettingsAttribute(description string) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: description,
		Validators: []validator.List{
			uniqueClickHouseNames("key", "setting"),
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
					Required:            true,
					MarkdownDescription: CLICKHOUSE_SETTING_VALUE_DESCRIPTION,
				},
			},
		},
	}
}

func getClickHouseProfilesAttribute() rschema.MapNestedAttribute {
	return rschema.MapNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_PROFILES_DESCRIPTION,
		Validators: []validator.Map{
			mapvalidator.KeysAre(stringvalidator.LengthAtLeast(1)),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"settings": getClickHouseSettingsAttribute(CLICKHOUSE_PROFILE_SETTINGS_DESCRIPTION),
			},
		},
	}
}

func getClickHouseUsersAttribute() rschema.MapNestedAttribute {
	return rschema.MapNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_USERS_DESCRIPTION,
		Validators: []validator.Map{
			mapvalidator.KeysAre(
				stringvalidator.LengthAtLeast(1),
				stringvalidator.NoneOf(ReservedClickHouseUsers...),
			),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"profile": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_USER_PROFILE_DESCRIPTION,
				},
				"quota": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_USER_QUOTA_DESCRIPTION,
				},
				"networks": rschema.SingleNestedAttribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_USER_NETWORKS_DESCRIPTION,
					Attributes: map[string]rschema.Attribute{
						"ips": rschema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							MarkdownDescription: CLICKHOUSE_USER_NETWORKS_IPS_DESCRIPTION,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.UniqueValues(),
							},
						},
					},
				},
				"access_management": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_USER_ACCESS_MANAGEMENT_DESCRIPTION,
				},
				"named_collection_control": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_USER_NAMED_COLLECTION_CONTROL_DESCRIPTION,
				},
				"show_named_collections_secrets": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_USER_SHOW_NAMED_COLLECTIONS_SECRETS_DESCRIPTION,
				},
				"databases": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_USER_DATABASES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.UniqueValues(),
					},
				},
				"grants": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_USER_GRANTS_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.UniqueValues(),
						listvalidator.ConflictsWith(
							path.MatchRelative().AtParent().AtName("databases"),
							path.MatchRelative().AtParent().AtName("access_management"),
							path.MatchRelative().AtParent().AtName("named_collection_control"),
							path.MatchRelative().AtParent().AtName("show_named_collections_secrets"),
						),
					},
				},
				"password": rschema.SingleNestedAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_USER_PASSWORD_DESCRIPTION,
					Attributes: map[string]rschema.Attribute{
						"type": rschema.StringAttribute{
							Required:            true,
							MarkdownDescription: CLICKHOUSE_USER_PASSWORD_TYPE_DESCRIPTION,
							Validators: []validator.String{
								stringvalidator.OneOf(clickHousePasswordTypes...),
							},
						},
						"enc_value": rschema.StringAttribute{
							Optional:            true,
							Sensitive:           true,
							MarkdownDescription: CLICKHOUSE_USER_PASSWORD_ENC_VALUE_DESCRIPTION,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
			},
		},
	}
}

func getClickHouseQuotasAttribute() rschema.MapNestedAttribute {
	return rschema.MapNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_QUOTAS_DESCRIPTION,
		Validators: []validator.Map{
			mapvalidator.KeysAre(stringvalidator.LengthAtLeast(1)),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"keyed_by": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					MarkdownDescription: CLICKHOUSE_QUOTA_KEYED_BY_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.OneOf(clickHouseQuotaKeyedBy...),
					},
				},
				"ipv4_prefix_bits": rschema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_QUOTA_IPV4_PREFIX_BITS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.Between(0, 32),
					},
				},
				"ipv6_prefix_bits": rschema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_QUOTA_IPV6_PREFIX_BITS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.Between(0, 128),
					},
				},
				"intervals": getClickHouseQuotaIntervalsAttribute(),
			},
		},
	}
}

func getClickHouseQuotaIntervalsAttribute() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Required:            true,
		MarkdownDescription: CLICKHOUSE_QUOTA_INTERVALS_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"duration": rschema.Int64Attribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_QUOTA_INTERVAL_DURATION_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"queries":                           getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_QUERIES_DESCRIPTION),
				"query_selects":                     getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_QUERY_SELECTS_DESCRIPTION),
				"query_inserts":                     getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_QUERY_INSERTS_DESCRIPTION),
				"errors":                            getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_ERRORS_DESCRIPTION),
				"result_rows":                       getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_RESULT_ROWS_DESCRIPTION),
				"result_bytes":                      getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_RESULT_BYTES_DESCRIPTION),
				"read_rows":                         getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_READ_ROWS_DESCRIPTION),
				"read_bytes":                        getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_READ_BYTES_DESCRIPTION),
				"written_bytes":                     getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_WRITTEN_BYTES_DESCRIPTION),
				"failed_sequential_authentications": getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_FAILED_AUTH_DESCRIPTION),
				"queries_per_normalized_hash":       getClickHouseQuotaLimitAttribute(CLICKHOUSE_QUOTA_QUERIES_PER_HASH_DESCRIPTION),
				"execution_time": rschema.Float64Attribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_QUOTA_EXECUTION_TIME_DESCRIPTION,
					Validators: []validator.Float64{
						float64validator.AtLeast(0),
					},
				},
			},
		},
	}
}

func getClickHouseQuotaLimitAttribute(description string) rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:            true,
		MarkdownDescription: description,
		Validators: []validator.Int64{
			int64validator.AtLeast(0),
		},
	}
}
