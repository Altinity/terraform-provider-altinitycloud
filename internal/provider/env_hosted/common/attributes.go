package hosted_env

import (
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func GetNodeGroupsAttribute(required, optional, computed bool, nodeTypeDescription string) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.NODE_GROUP_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: common.NODE_GROUP_NAME_DESCRIPTION,
				},
				"node_type": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: nodeTypeDescription,
				},
				"capacity_per_zone": rschema.Int64Attribute{
					Required:            true,
					MarkdownDescription: common.NODE_GROUP_CAPACITY_PER_ZONE_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"zone_ids": rschema.ListAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					Computed:            true,
					MarkdownDescription: common.HOSTED_NODE_GROUP_ZONE_IDS_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
				},
				"reservations": common.GetReservationsAttribute(true, false, false),
			},
		},
	}
}

func GetZoneIDsAttribute(required, optional, computed bool, description string) rschema.ListAttribute {
	attribute := rschema.ListAttribute{
		ElementType:         types.StringType,
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: description,
	}

	if optional || required {
		attribute.Validators = []validator.List{
			listvalidator.SizeAtLeast(2),
		}
	}

	return attribute
}
