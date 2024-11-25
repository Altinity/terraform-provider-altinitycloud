resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

variable "value" {
  type = string
}

resource "altinitycloud_env_secret" "this" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.value
}
