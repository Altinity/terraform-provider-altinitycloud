package env

import (
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ValidateForceDestroy checks if the environment is protected from deletion.
func ValidateForceDestroy(envName string, forceDestroy bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if !forceDestroy {
		diags.AddError("Env Locked", fmt.Sprintf("env %s is protected for deletion, set `force_destroy` property to `true` and run `terraform apply` to unlock it", envName))
	}
	return diags
}

// FormatDeleteError returns a user-friendly error message for delete failures.
func FormatDeleteError(envName string, err error) string {
	activeClusters, _ := client.IsActiveClustersError(err)
	if activeClusters {
		return fmt.Sprintf("Unable to delete env %s, it has active ClickHouse/Zookeeper clusters (use force_destroy_clusters=true to force delete them)", envName)
	}
	return fmt.Sprintf("Unable to delete env %s, got error: %s", envName, client.FormatError(err, envName))
}
