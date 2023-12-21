# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.1.1...v0.1.2)

### Added
- Allow env cluster deletion when destroyng `force_destroy_clusters` [259ed86](#)
- Added `bump` and `sync` commands to `Makefile` [ac3545f](#)

### Changed
- Bump go to `1.20` [69f6a2e](https://github.com/Altinity/terraform-provider-altinitycloud/commit/69f6a2ea059df2bd2435c982b7ce2b2532d5e788)
- Bump github.com/hashicorp/terraform-plugin-go to `0.20.0` [d86d633](https://github.com/Altinity/terraform-provider-altinitycloud/commit/d86d6339523946655c6165a931ff43a64f1bca4b)



## [0.1.1](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.1.0...v0.1.1)

### Fixed:
- Make load balancers `source_ip_ranges` property optional [1d9b688](https://github.com/Altinity/terraform-provider-altinitycloud/commit/1d9b688a704c36b2e5b8a19c97820db05ce24eb3).
- Fix load balancers `internal` mapper for k8s env [0cd00fd](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0cd00fd51b462d970cd71323449c15b38c0336da).
- Fix load balancers `internal` mapper for aws and gcp envs [87f4303](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0cd00fd51b462d970cd71323449c15b38c0336da).

### Changed:
- Bump github.com/hashicorp/terraform-plugin-testing to `1.6.0` [4d5a72f](https://github.com/Altinity/
terraform-provider-altinitycloud/commit/4d5a72f801091a45a39f7997ddb084f379901b54)
- Bump actions/setup-go to `5.0.0` [4982b08](https://github.com/Altinity/terraform-provider-altinitycloud/commit/4982b08f5b9fcb2ecae8c7e580c93f30842264bd)

## [0.1.0](https://github.com/Altinity/terraform-provider-altinitycloud/releases/tag/v0.1.0)

### Added

- Environment resources: `altinitycloud_env_aws`, `altinitycloud_env_gcp` and `altinitycloud_env_k8s`.
- Environment data sources: `altinitycloud_env_aws`, `altinitycloud_env_gcp` and `altinitycloud_env_k8s`.
- Environment status data source: `altinitycloud_env_aws_status`, `altinitycloud_env_gcp_status` and `altinitycloud_env_k8s_status`.
- Environment certificates resource: `altinitycloud_env_certificate`.
