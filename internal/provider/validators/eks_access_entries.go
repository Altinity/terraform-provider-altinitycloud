package validators

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func UniqueEKSAccessEntryPrincipals() validator.Set {
	return UniqueObjectAttribute(
		"principal_arn",
		"Duplicate EKS Access Entry Principal",
		func(arn string) string {
			return fmt.Sprintf("eks_access_entries contains more than one entry for principal %q. Each principal can be granted a single access level.", arn)
		},
	)
}
