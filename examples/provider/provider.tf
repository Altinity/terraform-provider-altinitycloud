terraform {
  required_providers {
    altinitycloud = {
      source = "altinity/altinitycloud"
      # https://github.com/altinity/terraform-provider-altinitycloud/blob/master/CHANGELOG.md
      version = "0.4.5"
    }
  }
}

provider "altinitycloud" {
  # `api_token` can be omitted if ALTINITYCLOUD_API_TOKEN env var is set.
  api_token = "XXXXXXXXXXXXXXXXXXXXXXXX"
}
