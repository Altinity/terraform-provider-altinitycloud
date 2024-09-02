terraform {
  required_providers {
    altinitycloud = {
      source = "altinity/altinitycloud"
      # https://github.com/altinity/terraform-provider-altinitycloud/blob/master/CHANGELOG.md
      version = "0.2.8"
    }
  }
}

provider "altinitycloud" {
  # `api_token` can be omitted if ALTINITYCLOUD_API_TOKEN env var is set.
  api_token = "XXXXXXXXXXXXXXXXXXXXXXXX"
}
