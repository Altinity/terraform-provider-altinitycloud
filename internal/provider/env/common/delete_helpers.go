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

// ValidateDisconnected checks if the environment is DISCONNECTED and returns
// differentiated error messages based on whether the env was ever provisioned.
func ValidateDisconnected(envName string, errorCode string, appliedSpecRevision int64, skipDeprovision, allowDeleteDisconnected bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if (errorCode == "DISCONNECTED" || errorCode == "K8S_DISCONNECTED") && !skipDeprovision && !allowDeleteDisconnected {
		msg := fmt.Sprintf("Unable to delete env %s, environment is DISCONNECTED.\n", envName)
		if appliedSpecRevision == 0 {
			msg += "The environment was never fully provisioned. Use `skip_deprovision_on_destroy=true` together with `allow_delete_while_disconnected=true` to clean up."
		} else {
			msg += "Check environment's `cloudconnect` or use `allow_delete_while_disconnected=true` to continue with the delete operation."
		}
		diags.AddError("Client Error", msg)
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
