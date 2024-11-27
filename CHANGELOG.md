# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.2](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.1...v0.4.2)
### Changed
- Bump github.com/hashicorp/terraform-plugin-docs to `0.20.1` [#122](https://github.com/Altinity/terraform-provider-altinitycloud/pull/122).
- Bump github.com/Yamashou/gqlgenc to `0.26.2` [#120](https://github.com/Altinity/terraform-provider-altinitycloud/pull/120).
- Add deprecation notice to `number_of_zones` environments property (it will be removed in future versions) [#15243a4](https://github.com/Altinity/terraform-provider-altinitycloud/commit/15243a4).

### Fixed
- Sort environments `node_groups` after API responses to match script order [#318164a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/318164a).

## [0.4.1](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.0...v0.4.1)
### Changed
- 🚨 [BREAKING CHANGE] Rename `altinitycloud_secret` to `altinitycloud_env_secret` [#03c38db](https://github.com/Altinity/terraform-provider-altinitycloud/commit/03c38db).

## [0.4.0](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.3.1...v0.4.0)
### Added
- Support for HCloud environments [#117](https://github.com/Altinity/terraform-provider-altinitycloud/pull/117).
- New encryption `altinitycloud_secret` resource and crypto SDK package [#115](https://github.com/Altinity/terraform-provider-altinitycloud/pull/115).
- Documentation and examples for HCloud [ac9a6d2](https://github.com/Altinity/terraform-provider-altinitycloud/commit/ac9a6d2).

### Changed
- Bump github.com/hashicorp/terraform-plugin-testing to `0.11.1` [#118](https://github.com/Altinity/terraform-provider-altinitycloud/pull/118).
- Re-arrange k8s environment examples [a372128](https://github.com/Altinity/terraform-provider-altinitycloud/commit/a372128).
- New user-agent format [2e82254](https://github.com/Altinity/terraform-provider-altinitycloud/commit/2e82254).
- Improve `skip_deprovision_on_destroy` property description [be4e7b7](https://github.com/Altinity/terraform-provider-altinitycloud/commit/be4e7b7).

### Fixed
- Improve "disconnected" error message while deleting envs [aa38c53](https://github.com/Altinity/terraform-provider-altinitycloud/commit/aa38c53).
- Don't allow empty strings on node group names [c6c71a7](https://github.com/Altinity/terraform-provider-altinitycloud/commit/c6c71a7).

## [0.3.1](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.3.0...v0.3.1)
### Added
- Add example of BYOK with AWS EKS cluster using Altinity's Terraform module for BYOK on EKS [45ed7a5](https://github.com/Altinity/terraform-provider-altinitycloud/commit/45ed7a5)\
- Add support to `PrivateDNS` on `altinitycloud_env_aws` VPC endpoints [598bf7e](https://github.com/Altinity/terraform-provider-altinitycloud/commit/598bf7e)

### Changed
- Allow deletion of environment while disconnected if it's using `skip_deprovision_on_destroy` property [3b911de](https://github.com/Altinity/terraform-provider-altinitycloud/commit/3b911de)
- Bump github.com/Yamashou/gqlgenc to `0.26.1` [#114](https://github.com/Altinity/terraform-provider-altinitycloud/pull/114).
- Bump github.com/hashicorp/terraform-plugin-docs to `0.20.0` [#112](https://github.com/Altinity/terraform-provider-altinitycloud/pull/112).

## [0.3.0](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.8...v0.3.0)
### Added
- Upgrade to go 1.22 [4f85ceb](https://github.com/Altinity/terraform-provider-altinitycloud/commit/4f85ceb)
- Add 5m MFA timeout when deleting environments [0909241](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0909241).
- Support error messages in environment status data sources [9ec8a64](https://github.com/Altinity/terraform-provider-altinitycloud/commit/9ec8a64).
- Do not allow environment deletion when env is disconnected [db5f17e](https://github.com/Altinity/terraform-provider-altinitycloud/commit/db5f17e).

### Changed
- Add new environment base resource to make env resources more DRY [ee2639a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/ee2639a).
- Add default "state-only" values when importing environments [e4e4880](https://github.com/Altinity/terraform-provider-altinitycloud/commit/e4e4880).
- Sync GraphQL schema with the latest version [c7babdc](https://github.com/Altinity/terraform-provider-altinitycloud/commit/c7babdc).
- Bump github.com/Yamashou/gqlgenc to `0.25.4` [#109](https://github.com/Altinity/terraform-provider-altinitycloud/pull/109).
- Bump github.com/hashicorp/terraform-plugin-framework-validators to `0.15.0` [#110](https://github.com/Altinity/terraform-provider-altinitycloud/pull/110).
- Bump github.com/hashicorp/terraform-plugin-framework to `0.12.0` [#93](https://github.com/Altinity/terraform-provider-altinitycloud/pull/93).


### Fixed
- Fix `aws_account_id` regex validation [d924ab2](https://github.com/Altinity/terraform-provider-altinitycloud/commit/d924ab2).
- Set environment ID when reading them [8c3628e](https://github.com/Altinity/terraform-provider-altinitycloud/commit/8c3628e).


## [0.2.8](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.7...v0.2.8)

### Added
- Friendly error message when deleting env with active clusters [6a4f437](https://github.com/Altinity/terraform-provider-altinitycloud/commit/6a4f437).

### Changed
- Bump github.com/hashicorp/terraform-plugin-framework to `1.11.0` [#87](https://github.com/Altinity/terraform-provider-altinitycloud/pull/87).
- Bump github.com/hashicorp/terraform-plugin-testing to `1.10.0` [#84](https://github.com/Altinity/terraform-provider-altinitycloud/pull/84).
- Bump github.com/Yamashou/gqlgenc to `0.24.0` [#88](https://github.com/Altinity/terraform-provider-altinitycloud/pull/88).
- Sync GraphQL schema with the latest version [c26a7e9](https://github.com/Altinity/terraform-provider-altinitycloud/commit/c26a7e9).
- Add resource force-replacement warning to the `name` property on environment resources [4a6c731](https://github.com/Altinity/terraform-provider-altinitycloud/commit/4a6c731).
- Don't allow to set empty region on environment resources [86ae9da](https://github.com/Altinity/terraform-provider-altinitycloud/commit/86ae9da).

## [0.2.7](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.6...v0.2.7)

### Changed
- New documentation templates for environment status data sources [f611072](https://github.com/Altinity/terraform-provider-altinitycloud/commit/f611072).


## [0.2.6](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.5...v0.2.6)

### Added
- Remove resource environments from planning state when get not found [5a1d473](https://github.com/Altinity/terraform-provider-altinitycloud/commit/5a1d473).
- Add missing docs for `skip_deprovision_on_destroy` environment resource property [9915a01](https://github.com/Altinity/terraform-provider-altinitycloud/commit/9915a01).
- New examples and better descriptions for `altinitycloud_env_***_status` data sources [ff1d62a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/ff1d62a)

### Changed
- Bump github.com/hashicorp/terraform-plugin-framework-validators to `0.13.0` [#66](https://github.com/Altinity/terraform-provider-altinitycloud/pull/66).
- Bump github.com/Yamashou/gqlgenc to `0.23.2` [#62](https://github.com/Altinity/terraform-provider-altinitycloud/pull/62).
- Bump github.com/hashicorp/terraform-plugin-framework to `1.10.0` [#64](https://github.com/Altinity/terraform-provider-altinitycloud/pull/64).
- Bump github.com/hashicorp/terraform-plugin-testing to `1.9.0` [#65](https://github.com/Altinity/terraform-provider-altinitycloud/pull/65).

### Fixed
- Increase minimun `zones` value to `2` for `altinitycloud_env_aws` [dd77f53](https://github.com/Altinity/terraform-provider-altinitycloud/commit/dd77f53)
- Fix `force_destroy` description on environment resources docs [19a695d](https://github.com/Altinity/terraform-provider-altinitycloud/commit/19a695d).
- Add reference to Altinity docs in environment resources docs [55d69b5](https://github.com/Altinity/terraform-provider-altinitycloud/commit/55d69b5)


## [0.2.5](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.4...v0.2.5)

### Changed
- Increase delete timeout to 60 minutes when deleting environments[ac426d7](https://github.com/Altinity/terraform-provider-altinitycloud/commit/ac426d7)
- Bump github.com/hashicorp/terraform-plugin-docs to `0.19.4` [#56](https://github.com/Altinity/terraform-provider-altinitycloud/pull/56).

### Fixed
- Make `cloud_connect` read-only for `altinitycloud_env_aws` data source[bfdd203](https://github.com/Altinity/terraform-provider-altinitycloud/commit/bfdd203)

## [0.2.4](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.3...v0.2.4)

### Changed
- Bump github.com/hashicorp/terraform-plugin-docs to `0.19.3` [#54](https://github.com/Altinity/terraform-provider-altinitycloud/pull/54).
- Bump github.com/hashicorp/terraform-plugin-framework to `1.9.0` [#55](https://github.com/Altinity/terraform-provider-altinitycloud/pull/55).
- Bump github.com/hashicorp/terraform-plugin-testing to `1.8.0` [#53](https://github.com/Altinity/terraform-provider-altinitycloud/pull/53).

### Fixed
- Documentation error when setting up peering connections for AWS environments [3851c8d](https://github.com/Altinity/terraform-provider-altinitycloud/commit/3851c8d).

## [0.2.3](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.2...v0.2.3)

### Changed
- Bump go to `1.21` [82082e8](https://github.com/Altinity/terraform-provider-altinitycloud/commit/82082e8).
- Bump github.com/hashicorp/terraform-plugin-go to `0.23.0` [#47](https://github.com/Altinity/terraform-provider-altinitycloud/pull/47).

## [0.2.2](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.1...v0.2.2)

### Added
- Add `pendingMFA` property to environment SDK schemas [fd67661](https://github.com/Altinity/terraform-provider-altinitycloud/commit/fd67661).

### Changed
- Bump github.com/Yamashou/gqlgenc to `0.23.1` [#28](https://github.com/Altinity/terraform-provider-altinitycloud/pull/28).
- Bump github.com/hashicorp/terraform-plugin-docs to `0.19.2` [#43](https://github.com/Altinity/terraform-provider-altinitycloud/pull/43).
- Bump github.com/hashicorp/terraform-plugin-framework to `1.8.0` [#36](https://github.com/Altinity/terraform-provider-altinitycloud/pull/36).

### Fixed
- Documentation typos on K8s environment [5f96183](https://github.com/Altinity/terraform-provider-altinitycloud/commit/5f96183).

## [0.2.1](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.2.0...v0.2.1)

### Fixed
- Update Azure docs and fix descriptions [#29](https://github.com/Altinity/terraform-provider-altinitycloud/pull/29).

## [0.2.0](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.1.2...v0.2.0)

### Added
- Azure Environment resource: `altinitycloud_env_azure` [#28](https://github.com/Altinity/terraform-provider-altinitycloud/pull/28).
- Azure Environment data source: `altinitycloud_env_azure` [#28](https://github.com/Altinity/terraform-provider-altinitycloud/pull/28).
- Azure Environment status data source: `altinitycloud_env_azure_status` [#28](https://github.com/Altinity/terraform-provider-altinitycloud/pull/28).

### Changed
- Bump github.com/hashicorp/terraform-plugin-go to `0.22.1` [#25](https://github.com/Altinity/terraform-provider-altinitycloud/pull/25).
- Bump github.com/hashicorp/terraform-plugin-framework to `1.7.0` [#27](https://github.com/Altinity/terraform-provider-altinitycloud/pull/27).
- Bump github.com/hashicorp/terraform-plugin-testing to `1.7.0` [#22](https://github.com/Altinity/terraform-provider-altinitycloud/pull/22).
- Bump github.com/Yamashou/gqlgenc to `0.19.3` [#24](https://github.com/Altinity/terraform-provider-altinitycloud/pull/24).
- Bump github.com/hashicorp/terraform-plugin-docs to `0.18.0` [#12](https://github.com/Altinity/terraform-provider-altinitycloud/pull/12).

### Fixed
- Remove `v` from prefix command [91fa91b](https://github.com/Altinity/terraform-provider-altinitycloud/commit/91fa91b3026edfdb6897765de60d2a1bdfac2780).
- Allow gen to work with default graphql file and url [c8afa98](https://github.com/Altinity/terraform-provider-altinitycloud/commit/c8afa98ebf566daa5ef4719a5927aaa18cc75392).
- Fix typo on k8s env sample [a339a62](https://github.com/Altinity/terraform-provider-altinitycloud/commit/a339a62b580b1d9a4e7fb83ae6527d2a6c299230).

## [0.1.2](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.1.1...v0.1.2)

### Added
- Allow env cluster deletion when destroyng `force_destroy_clusters` [259ed86](https://github.com/Altinity/terraform-provider-altinitycloud/commit/259ed86f1d18cd6d0ce6c93b4c0f65626bb90492).
- Added `bump` and `sync` commands to `Makefile` [ac3545f](https://github.com/Altinity/terraform-provider-altinitycloud/commit/ac3545f2fdb0e970b349dabb7e9afa2524680589).

### Changed
- Bump go to `1.20` [69f6a2e](https://github.com/Altinity/terraform-provider-altinitycloud/commit/69f6a2ea059df2bd2435c982b7ce2b2532d5e788).
- Bump github.com/hashicorp/terraform-plugin-go to `0.20.0` [d86d633](https://github.com/Altinity/terraform-provider-altinitycloud/commit/d86d6339523946655c6165a931ff43a64f1bca4b).

## [0.1.1](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.1.0...v0.1.1)

### Fixed:
- Make load balancers `source_ip_ranges` property optional [1d9b688](https://github.com/Altinity/terraform-provider-altinitycloud/commit/1d9b688a704c36b2e5b8a19c97820db05ce24eb3).
- Fix load balancers `internal` mapper for k8s env [0cd00fd](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0cd00fd51b462d970cd71323449c15b38c0336da).
- Fix load balancers `internal` mapper for aws and gcp envs [87f4303](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0cd00fd51b462d970cd71323449c15b38c0336da).

### Changed:
- Bump github.com/hashicorp/terraform-plugin-testing to `1.6.0` [4d5a72f](https://github.com/Altinity/
terraform-provider-altinitycloud/commit/4d5a72f801091a45a39f7997ddb084f379901b54).
- Bump actions/setup-go to `5.0.0` [4982b08](https://github.com/Altinity/terraform-provider-altinitycloud/commit/4982b08f5b9fcb2ecae8c7e580c93f30842264bd).

## [0.1.0](https://github.com/Altinity/terraform-provider-altinitycloud/releases/tag/v0.1.0)

### Added

- Environment resources: `altinitycloud_env_aws`, `altinitycloud_env_gcp` and `altinitycloud_env_k8s`.
- Environment data sources: `altinitycloud_env_aws`, `altinitycloud_env_gcp` and `altinitycloud_env_k8s`.
- Environment status data sources: `altinitycloud_env_aws_status`, `altinitycloud_env_gcp_status` and `altinitycloud_env_k8s_status`.
- Environment certificates resource: `altinitycloud_env_certificate`.
