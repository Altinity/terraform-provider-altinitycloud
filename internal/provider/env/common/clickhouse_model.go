package env

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClickHouseClusterModel struct {
	Name           types.String               `tfsdk:"name"`
	Image          types.String               `tfsdk:"image"`
	InstanceType   types.String               `tfsdk:"instance_type"`
	Zones          types.List                 `tfsdk:"zones"`
	Shards         types.Int64                `tfsdk:"shards"`
	Replicas       types.Int64                `tfsdk:"replicas"`
	Stopped        types.Bool                 `tfsdk:"stopped"`
	Keeper         types.String               `tfsdk:"keeper"`
	Disk           *ClickHouseDiskModel       `tfsdk:"disk"`
	ExtraDisks     []ClickHouseNamedDiskModel `tfsdk:"extra_disks"`
	Users          []ClickHouseUserModel      `tfsdk:"users"`
	Profiles       []ClickHouseProfileModel   `tfsdk:"profiles"`
	Quotas         []ClickHouseQuotaModel     `tfsdk:"quotas"`
	ServerSettings []ClickHouseSettingModel   `tfsdk:"server_settings"`
}

type ClickHouseKeeperModel struct {
	Name         types.String         `tfsdk:"name"`
	InstanceType types.String         `tfsdk:"instance_type"`
	Zones        types.List           `tfsdk:"zones"`
	Replicas     types.Int64          `tfsdk:"replicas"`
	Stopped      types.Bool           `tfsdk:"stopped"`
	Disk         *ClickHouseDiskModel `tfsdk:"disk"`
}

type ClickHouseDiskModel struct {
	SizeGiB        types.Int64 `tfsdk:"size_gib"`
	Iops           types.Int64 `tfsdk:"iops"`
	ThroughputMiBs types.Int64 `tfsdk:"throughput_mibps"`
}

type ClickHouseNamedDiskModel struct {
	Name           types.String `tfsdk:"name"`
	SizeGiB        types.Int64  `tfsdk:"size_gib"`
	Iops           types.Int64  `tfsdk:"iops"`
	ThroughputMiBs types.Int64  `tfsdk:"throughput_mibps"`
}

type ClickHouseSettingModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type ClickHouseProfileModel struct {
	Name     types.String             `tfsdk:"name"`
	Settings []ClickHouseSettingModel `tfsdk:"settings"`
}

type ClickHouseUserModel struct {
	Name                        types.String                 `tfsdk:"name"`
	Profile                     types.String                 `tfsdk:"profile"`
	Quota                       types.String                 `tfsdk:"quota"`
	Networks                    *ClickHouseUserNetworksModel `tfsdk:"networks"`
	AccessManagement            types.Bool                   `tfsdk:"access_management"`
	NamedCollectionControl      types.Bool                   `tfsdk:"named_collection_control"`
	ShowNamedCollectionsSecrets types.Bool                   `tfsdk:"show_named_collections_secrets"`
	Databases                   types.List                   `tfsdk:"databases"`
	Grants                      types.List                   `tfsdk:"grants"`
	Password                    *ClickHouseUserPasswordModel `tfsdk:"password"`
}

type ClickHouseUserNetworksModel struct {
	Ips types.List `tfsdk:"ips"`
}

// `enc_value` is write-only: the API never returns it, the way datadog's
// enc_api_key does not come back either.
type ClickHouseUserPasswordModel struct {
	Type     types.String `tfsdk:"type"`
	EncValue types.String `tfsdk:"enc_value"`
}

type ClickHouseQuotaModel struct {
	Name           types.String                   `tfsdk:"name"`
	KeyedBy        types.String                   `tfsdk:"keyed_by"`
	IPv4PrefixBits types.Int64                    `tfsdk:"ipv4_prefix_bits"`
	IPv6PrefixBits types.Int64                    `tfsdk:"ipv6_prefix_bits"`
	Intervals      []ClickHouseQuotaIntervalModel `tfsdk:"intervals"`
}

type ClickHouseQuotaIntervalModel struct {
	DurationSeconds                 types.Int64   `tfsdk:"duration_seconds"`
	Queries                         types.Int64   `tfsdk:"queries"`
	QuerySelects                    types.Int64   `tfsdk:"query_selects"`
	QueryInserts                    types.Int64   `tfsdk:"query_inserts"`
	Errors                          types.Int64   `tfsdk:"errors"`
	ResultRows                      types.Int64   `tfsdk:"result_rows"`
	ResultBytes                     types.Int64   `tfsdk:"result_bytes"`
	ReadRows                        types.Int64   `tfsdk:"read_rows"`
	ReadBytes                       types.Int64   `tfsdk:"read_bytes"`
	WrittenBytes                    types.Int64   `tfsdk:"written_bytes"`
	ExecutionTimeSeconds            types.Float64 `tfsdk:"execution_time_seconds"`
	FailedSequentialAuthentications types.Int64   `tfsdk:"failed_sequential_authentications"`
	QueriesPerNormalizedHash        types.Int64   `tfsdk:"queries_per_normalized_hash"`
}

// Env-agnostic view of the read fragments, which gqlgenc generates per env.
type ClickHouseClusterSpec struct {
	Name           string
	Image          string
	InstanceType   string
	Zones          []string
	Shards         int64
	Replicas       int64
	Stopped        bool
	Keeper         string
	Disk           ClickHouseDiskSpec
	ExtraDisks     []ClickHouseNamedDiskSpec
	Users          []ClickHouseUserSpec
	Profiles       []ClickHouseProfileSpec
	Quotas         []ClickHouseQuotaSpec
	ServerSettings []ClickHouseSettingSpec
}

type ClickHouseKeeperSpec struct {
	Name         string
	InstanceType string
	Zones        []string
	Replicas     int64
	Stopped      bool
	Disk         ClickHouseDiskSpec
}

type ClickHouseDiskSpec struct {
	SizeGiB        int64
	Iops           int64
	ThroughputMiBs int64
}

type ClickHouseNamedDiskSpec struct {
	Name           string
	SizeGiB        int64
	Iops           int64
	ThroughputMiBs int64
}

type ClickHouseSettingSpec struct {
	Key   string
	Value string
}

type ClickHouseProfileSpec struct {
	Name     string
	Settings []ClickHouseSettingSpec
}

type ClickHouseUserSpec struct {
	Name                        string
	Profile                     string
	Quota                       string
	Ips                         []string
	AccessManagement            bool
	NamedCollectionControl      bool
	ShowNamedCollectionsSecrets bool
	Databases                   []string
	Grants                      []string
	PasswordType                string
}

type ClickHouseQuotaSpec struct {
	Name           string
	KeyedBy        string
	IPv4PrefixBits *int64
	IPv6PrefixBits *int64
	Intervals      []ClickHouseQuotaIntervalSpec
}

type ClickHouseQuotaIntervalSpec struct {
	DurationSeconds                 uint64
	Queries                         *uint64
	QuerySelects                    *uint64
	QueryInserts                    *uint64
	Errors                          *uint64
	ResultRows                      *uint64
	ResultBytes                     *uint64
	ReadRows                        *uint64
	ReadBytes                       *uint64
	WrittenBytes                    *uint64
	ExecutionTimeSeconds            *float64
	FailedSequentialAuthentications *uint64
	QueriesPerNormalizedHash        *uint64
}
