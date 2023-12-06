# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/Altinity/terraform-provider-altinitycloud/releases/tag/v0.1.1)

### Fixed:
- Make load balancers `source_ip_ranges` property optional [1d9b688](https://github.com/Altinity/terraform-provider-altinitycloud/commit/1d9b688a704c36b2e5b8a19c97820db05ce24eb3).
- Fix load balancers `internal`` mapper for k8s environemtns [0cd00fd](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0cd00fd51b462d970cd71323449c15b38c0336da).

## [0.1.0](https://github.com/Altinity/terraform-provider-altinitycloud/releases/tag/v0.1.0)

### Added

- Environment resources: `altinitycloud_env_aws`, `altinitycloud_env_gcp` and `altinitycloud_env_k8s`.
- Environment data sources: `altinitycloud_env_aws`, `altinitycloud_env_gcp` and `altinitycloud_env_k8s`.
- Environment status data source: `altinitycloud_env_aws_status`, `altinitycloud_env_gcp_status` and `altinitycloud_env_k8s_status`.
- Environment certificates resource: `altinitycloud_env_certificate`.
