package common

import (
	"regexp"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Platform-injected: the API rejects them on input and can return them in a read.
var ReservedClickHouseUsers = []string{"grafana", "datadog"}

var clickHouseNameRegex = regexp.MustCompile("^[a-z0-9][a-z0-9-]{0,13}[a-z0-9]$")

// The password type enum is generated per env, but every env spells the values
// the same, so a shared attribute can key off the AWS one.
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
				"image": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
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
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_CLUSTER_SHARDS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"replicas": rschema.Int64Attribute{
					Optional:            true,
					Computed:            true,
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
					Computed:            true,
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
				"replicas": rschema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_KEEPER_REPLICAS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
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
			"size_gib":         getClickHouseDiskSizeAttribute(),
			"iops":             getClickHouseDiskIopsAttribute(),
			"throughput_mibps": getClickHouseDiskThroughputAttribute(),
		},
	}
}

func getClickHouseExtraDisksAttribute() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_EXTRA_DISKS_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtMost(8),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_EXTRA_DISK_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 16),
					},
				},
				"size_gib":         getClickHouseDiskSizeAttribute(),
				"iops":             getClickHouseDiskIopsAttribute(),
				"throughput_mibps": getClickHouseDiskThroughputAttribute(),
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
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_SETTING_VALUE_DESCRIPTION,
				},
			},
		},
	}
}

func getClickHouseProfilesAttribute() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_PROFILES_DESCRIPTION,
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_PROFILE_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"settings": getClickHouseSettingsAttribute(CLICKHOUSE_PROFILE_SETTINGS_DESCRIPTION),
			},
		},
	}
}

func getClickHouseUsersAttribute() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_USERS_DESCRIPTION,
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_USER_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
						stringvalidator.NoneOf(ReservedClickHouseUsers...),
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
							},
						},
					},
				},
				"access_management": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_USER_ACCESS_MANAGEMENT_DESCRIPTION,
				},
				"named_collection_control": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_USER_NAMED_COLLECTION_CONTROL_DESCRIPTION,
				},
				"show_named_collections_secrets": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_USER_SHOW_NAMED_COLLECTIONS_SECRETS_DESCRIPTION,
				},
				"databases": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_USER_DATABASES_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
				},
				"grants": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					Computed:            true,
					MarkdownDescription: CLICKHOUSE_USER_GRANTS_DESCRIPTION,
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

func getClickHouseQuotasAttribute() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: CLICKHOUSE_QUOTAS_DESCRIPTION,
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: CLICKHOUSE_QUOTA_NAME_DESCRIPTION,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"keyed_by": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
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
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"duration_seconds": rschema.Int64Attribute{
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
				"execution_time_seconds": rschema.Float64Attribute{
					Optional:            true,
					MarkdownDescription: CLICKHOUSE_QUOTA_EXECUTION_TIME_DESCRIPTION,
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
