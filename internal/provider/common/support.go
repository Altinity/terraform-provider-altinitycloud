package common

const SupportMessage = `

If you need help, reach out to us via Slack:
  - Enterprise customers: Use your organization's dedicated Altinity Slack channel.
  - Community: Join the AltinityDB workspace (https://altinitydbworkspace.slack.com/) and post in the #terraform channel.`

func FormatClientError(detail string) string {
	return detail + SupportMessage
}
