resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

# If you need to re-generate certificate, use the `terraform taint` command to force the resource to be re-created.
# https://developer.hashicorp.com/terraform/cli/commands/taint
#
# $ terraform taint altinitycloud_env_certificate.this
