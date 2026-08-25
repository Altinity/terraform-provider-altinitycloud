package env

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClickHouseClusterModel struct {
	Name            types.String                    `tfsdk:"name"`
	Mode            types.String                    `tfsdk:"mode"`
	Image           types.String                    `tfsdk:"image"`
	InstanceType    types.String                    `tfsdk:"instance_type"`
	Zones           types.List                      `tfsdk:"zones"`
	Shards          types.Int64                     `tfsdk:"shards"`
	Replicas        types.Int64                     `tfsdk:"replicas"`
	Stopped         types.Bool                      `tfsdk:"stopped"`
	Disk            *ClickHouseDiskModel            `tfsdk:"disk"`
	AdditionalDisks []ClickHouseAdditionalDiskModel `tfsdk:"additional_disks"`
	Keeper          *ClickHouseKeeperRefModel       `tfsdk:"keeper"`
	Settings        []ClickHouseSettingModel        `tfsdk:"settings"`
	Profiles        []ClickHouseProfileModel        `tfsdk:"profiles"`
	Users           []ClickHouseUserModel           `tfsdk:"users"`
}

type ClickHouseKeeperModel struct {
	Name         types.String         `tfsdk:"name"`
	InstanceType types.String         `tfsdk:"instance_type"`
	Zones        types.List           `tfsdk:"zones"`
	HA           types.Bool           `tfsdk:"ha"`
	Stopped      types.Bool           `tfsdk:"stopped"`
	Disk         *ClickHouseDiskModel `tfsdk:"disk"`
}

// The main volume is always named `default`, so the name is not part of the config.
type ClickHouseDiskModel struct {
	Size         types.Int64  `tfsdk:"size"`
	StorageClass types.String `tfsdk:"storage_class"`
	Iops         types.Int64  `tfsdk:"iops"`
	Throughput   types.Int64  `tfsdk:"throughput"`
}

type ClickHouseAdditionalDiskModel struct {
	Name         types.String `tfsdk:"name"`
	Size         types.Int64  `tfsdk:"size"`
	StorageClass types.String `tfsdk:"storage_class"`
	Iops         types.Int64  `tfsdk:"iops"`
	Throughput   types.Int64  `tfsdk:"throughput"`
}

type ClickHouseKeeperRefModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Name    types.String `tfsdk:"name"`
}

type ClickHouseSettingModel struct {
	Key             types.String              `tfsdk:"key"`
	Value           types.String              `tfsdk:"value"`
	ValueFromSecret *ClickHouseSecretRefModel `tfsdk:"value_from_secret"`
}

type ClickHouseProfileModel struct {
	Name     types.String             `tfsdk:"name"`
	Settings []ClickHouseSettingModel `tfsdk:"settings"`
}

type ClickHouseUserModel struct {
	Name                        types.String              `tfsdk:"name"`
	Profile                     types.String              `tfsdk:"profile"`
	Quota                       types.String              `tfsdk:"quota"`
	AllowedCIDRs                types.List                `tfsdk:"allowed_cidrs"`
	Databases                   types.List                `tfsdk:"databases"`
	AccessManagement            types.Bool                `tfsdk:"access_management"`
	NamedCollectionControl      types.Bool                `tfsdk:"named_collection_control"`
	ShowNamedCollections        types.Bool                `tfsdk:"show_named_collections"`
	ShowNamedCollectionsSecrets types.Bool                `tfsdk:"show_named_collections_secrets"`
	PasswordType                types.String              `tfsdk:"password_type"`
	PasswordValue               types.String              `tfsdk:"password_value"`
	PasswordValueFromSecret     *ClickHouseSecretRefModel `tfsdk:"password_value_from_secret"`
}

type ClickHouseSecretRefModel struct {
	Name types.String `tfsdk:"name"`
	Key  types.String `tfsdk:"key"`
}

// Env-agnostic view of the read fragments, which gqlgenc generates per env. Each
// env converts its own fragment into these so the model mapping is shared.
type ClickHouseClusterSpec struct {
	Name            string
	Mode            string
	Image           string
	InstanceType    string
	Zones           []string
	Shards          int64
	Replicas        int64
	Stopped         bool
	Disk            ClickHouseDiskSpec
	AdditionalDisks []ClickHouseDiskSpec
	KeeperName      *string
	Settings        []ClickHouseSettingSpec
	Profiles        []ClickHouseProfileSpec
	Users           []ClickHouseUserSpec
}

type ClickHouseKeeperSpec struct {
	Name         string
	InstanceType string
	Zones        []string
	HA           bool
	Stopped      bool
	Disk         ClickHouseDiskSpec
}

type ClickHouseDiskSpec struct {
	Name         string
	Size         int64
	StorageClass string
	Iops         int64
	Throughput   int64
}

type ClickHouseSettingSpec struct {
	Key             string
	Value           string
	ValueFromSecret *ClickHouseSecretRefSpec
}

type ClickHouseProfileSpec struct {
	Name     string
	Settings []ClickHouseSettingSpec
}

type ClickHouseUserSpec struct {
	Name                        string
	Profile                     string
	Quota                       string
	AllowedCIDRs                []string
	Databases                   []string
	AccessManagement            bool
	NamedCollectionControl      bool
	ShowNamedCollections        bool
	ShowNamedCollectionsSecrets bool
	PasswordType                *string
	PasswordValueFromSecret     *ClickHouseSecretRefSpec
}

type ClickHouseSecretRefSpec struct {
	Name string
	Key  string
}
