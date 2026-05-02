package common

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// supportMessage is appended to API-related error diagnostics.
// It uses plain text (not Markdown) because it renders in Terraform CLI output.
const supportMessage = `

If you need help, reach out to us via:
  - Slack (Enterprise customers): Use your organization's dedicated Altinity Slack channel.
  - Slack (Community): Join the AltinityDB workspace (https://altinitydbworkspace.slack.com/) and post in the #terraform channel.
  - GitHub: Open an issue at https://github.com/altinity/terraform-provider-altinitycloud/issues/new to report bugs or request features.`

// AddClientError appends a "Client Error" diagnostic to diags. The support
// message is only appended once per diagnostics collection to avoid duplicates
// when multiple errors are emitted during the same operation.
func AddClientError(diags *diag.Diagnostics, detail string) {
	AddSupportError(diags, "Client Error", detail)
}

// AddSupportError appends an error diagnostic with the given summary and
// attaches the support contact message to the detail, but only if no other
// diagnostic in the collection already includes it.
func AddSupportError(diags *diag.Diagnostics, summary, detail string) {
	if hasSupportMessage(*diags) {
		diags.AddError(summary, detail)
		return
	}
	diags.AddError(summary, detail+supportMessage)
}

func hasSupportMessage(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if strings.Contains(d.Detail(), supportMessage) {
			return true
		}
	}
	return false
}
