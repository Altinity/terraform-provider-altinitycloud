package validators

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func UniqueExternalBucketNames() validator.Set {
	return UniqueObjectAttribute(
		"name",
		"Duplicate External Bucket Name",
		func(name string) string {
			return fmt.Sprintf("external_buckets contains more than one entry for bucket %q. Bucket names must be unique.", name)
		},
	)
}
