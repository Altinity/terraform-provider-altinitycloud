# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.21](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.20...v0.4.21)
### Fixed
- Do not allow empty strings on `hcloud_token_enc` property [10f1676](https://github.com/Altinity/terraform-provider-altinitycloud/commit/10f1676).
- Do not keep looping data source status when other error than `DISCONNECTED` is returned [6d81b92](https://github.com/Altinity/terraform-provider-altinitycloud/commit/6d81b92).
- Add missing validations and descriptions to `altinitycloud_env_secret` and `altinitycloud_env_certificate` resources [14caf21](https://github.com/Altinity/terraform-provider-altinitycloud/commit/14caf21).
- Add missing spec revision to `altinitycloud_env_*` data sources [c4faf74](https://github.com/Altinity/terraform-provider-altinitycloud/commit/c4faf74).


### Changed
- Improve env status data sources comment in examples [52c2b9b](https://github.com/Altinity/terraform-provider-altinitycloud/commit/52c2b9b).

## [0.4.20](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.19...v0.4.20)
### Changed
- Bump github.com/stretchr/testify to `1.11.1` [#180](https://github.com/Altinity/terraform-provider-altinitycloud/pull/180).
- Sync GraphQL schema with the latest version [9908bf2](https://github.com/Altinity/terraform-provider-altinitycloud/commit/9908bf2).

### Fixed
- Send `external_buckets` property when updating `altinitycloud_env_aws` resource [fa65c22](https://github.com/Altinity/terraform-provider-altinitycloud/commit/fa65c22).
- Fix `altinitycloud_env_aws` output peering example [d019a43](https://github.com/Altinity/terraform-provider-altinitycloud/commit/d019a43).

## [0.4.19](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.18...v0.4.19)
### Added
- Add `external_buckets` property to `altinitycloud_env_aws` resource [1611f0f](https://github.com/Altinity/terraform-provider-altinitycloud/commit/1611f0f).

## [0.4.18](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.17...v0.4.18)

### Added
- Add test cases for `altinitycloud_env_status` models [99c0206](https://github.com/Altinity/terraform-provider-altinitycloud/commit/99c0206).

### Changed
- Bump github.com/hashicorp/terraform-plugin-testing to `1.13.3` [#176](https://github.com/Altinity/terraform-provider-altinitycloud/pull/176).
- Bump github.com/hashicorp/terraform-plugin-testing to `1.15.1` [#173](https://github.com/Altinity/terraform-provider-altinitycloud/pull/173).
- Bump github.com/hashicorp/terraform-plugin-docs to `0.22.0` [#172](https://github.com/Altinity/terraform-provider-altinitycloud/pull/172).
- Bump github.com/Yamashou/gqlgenc to `0.33.0` [#171](https://github.com/Altinity/terraform-provider-altinitycloud/pull/171).
- Sync GraphQL schema with the latest version [2c8002a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/2c8002a).
- Update docs for AWS peering and VPC endpoint examples [457e209](https://github.com/Altinity/terraform-provider-altinitycloud/commit/457e209).

### Fixed
- Update docs distribution on k8s GKE example [30a094e](https://github.com/Altinity/terraform-provider-altinitycloud/commit/30a094e).
- Update zone names in `altinitycloud_gcp_env` resource examples [6217068](https://github.com/Altinity/terraform-provider-altinitycloud/commit/6217068).
- Preserve API response while reordering `node_groups` and `zones` properties [f1e9256](https://github.com/Altinity/terraform-provider-altinitycloud/commit/f1e9256).

## [0.4.17](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.16...v0.4.17)
### Fixed
- Allow `hcloud_token_enc` to be updated [4cbed8a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/4cbed8a).

## [0.4.16](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.15...v0.4.16)
### Added
- Add network peering example for GCP environments in docs [c07765e](https://github.com/Altinity/terraform-provider-altinitycloud/commit/c07765e).

## [0.4.15](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.14...v0.4.15)
### Changed
- Bump github.com/hashicorp/terraform-plugin-framework-validators to `0.21.0` [#163](https://github.com/Altinity/terraform-provider-altinitycloud/pull/163).
- Bump github.com/hashicorp/terraform-plugin-go to `0.28.0` [#167](https://github.com/Altinity/terraform-provider-altinitycloud/pull/167).
- Bump github.com/hashicorp/terraform-plugin-testing to `1.13.1` [#168](https://github.com/Altinity/terraform-provider-altinitycloud/pull/168).
- Bump github.com/hashicorp/terraform-plugin-framework to `1.15.0` [#164](https://github.com/Altinity/terraform-provider-altinitycloud/pull/164).
- Update Altinity and AWS links in docs [#159](https://github.com/Altinity/terraform-provider-altinitycloud/pull/159).

### Fixed
- Fix `node_groups` schema to use `list` instead of `set` [61c9bde](https://github.com/Altinity/terraform-provider-altinitycloud/commit/61c9bde).

## [0.4.14](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.13...v0.4.14)
### Added
- Support for AWS Permissions Boundary [d153753f](https://github.com/Altinity/terraform-provider-altinitycloud/commit/d153753f).


## [0.4.13](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.12...v0.4.13)
### Changed
- Sync GraphQL schema with the latest version [e35def2](https://github.com/Altinity/terraform-provider-altinitycloud/commit/e35def2).

### Fixed
- Keep `zones` and `locations` properties in sync with the API response order [#4820ba0](https://github.com/Altinity/terraform-provider-altinitycloud/commit/4820ba0).

## [0.4.12](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.11...v0.4.12)
### Fixed
- Handle `DISCONNECTED` status  on all env status data sources [#1e7474b](https://github.com/Altinity/terraform-provider-altinitycloud/commit/1e7474b).


## [0.4.11](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.10...v0.4.11)
### Changed
- Bump github.com/hashicorp/terraform-plugin-docs to `0.21.0` [#149](https://github.com/Altinity/terraform-provider-altinitycloud/pull/149).
- Bump golangci/golangci-lint-action to `6.5.0` and update linter settings [#145](https://github.com/Altinity/terraform-provider-altinitycloud/pull/145).
- Bump github.com/Yamashou/gqlgenc to `0.31.0` [#148](https://github.com/Altinity/terraform-provider-altinitycloud/pull/148).
- Bump github.com/hashicorp/terraform-plugin-framework-validators to `0.17.0` [#146](https://github.com/Altinity/terraform-provider-altinitycloud/pull/146).
- Bump github.com/hashicorp/terraform-plugin-framework to `0.14.1` [#147](https://github.com/Altinity/terraform-provider-altinitycloud/pull/147).

### Fixed
- Add missing `private_dns` property to `altinitycloud_env_aws` schema [#383e8f8](https://github.com/Altinity/terraform-provider-altinitycloud/commit/383e8f8).

## [0.4.10](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.9...v0.4.10)
### Added
- Support peering connection on GCP environments [#f234d07](https://github.com/Altinity/terraform-provider-altinitycloud/commit/f234d07).
- Support private service consumers on GCP environments [#34e6d19](https://github.com/Altinity/terraform-provider-altinitycloud/commit/34e6d19).

### Changed
- Bump github.com/Yamashou/gqlgenc to `0.30.3` [#139](https://github.com/Altinity/terraform-provider-altinitycloud/pull/139).
- Bump github.com/hashicorp/terraform-plugin-go to `0.26.0` [#138](https://github.com/Altinity/terraform-provider-altinitycloud/pull/138).

## [0.4.9](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.8...v0.4.9)
### Fixed
- Rollback [#0655c4a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0655c4a) which caused a inconsestency on envs creation [#fdb23d7](https://github.com/Altinity/terraform-provider-altinitycloud/commit/fdb23d7).

## [0.4.8](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.7...v0.4.8)
> ⚠️ Skip this version since it introuced a bug with inconsistency results during envs creation. Use `v0.4.9` instead.

### Changed
- Update docs to use `ccx23` as hcloud default for clickhouse nodes [#d0e0ba5](https://github.com/Altinity/terraform-provider-altinitycloud/commit/d0e0ba5).

### Fixed
- Remove `computed` from node group `zones/location` properties on environments [#0655c4a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0655c4a)

## [0.4.7](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.6...v0.4.7)
### Changed
- Update docs to use `ccx21` as hcloud default for clickhouse nodes [#e67a77f](https://github.com/Altinity/terraform-provider-altinitycloud/commit/e67a77f).

## [0.4.6](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.5...v0.4.6)
### Added
- Allow delete disconnected environments using `allow_delete_while_disconnected` [#01e5bcc](https://github.com/Altinity/terraform-provider-altinitycloud/commit/01e5bcc).

### Changed
- Bump github.com/Yamashou/gqlgenc to `0.30.2` [#135](https://github.com/Altinity/terraform-provider-altinitycloud/pull/135).

## [0.4.5](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.4...v0.4.5)
### Added
- Add `nat` property to enable AWS NAT Gateway on AWS envs [#e2ade63](https://github.com/Altinity/terraform-provider-altinitycloud/commit/e2ade63).

### Changed
- 🚨 [BREAKING CHANGE] Deprecate `number_of_zones` from all env resources [#0f07c7a](https://github.com/Altinity/terraform-provider-altinitycloud/commit/0f07c7a).
- Bump github.com/hashicorp/terraform-plugin-framework-validators to `0.16.0` [#128](https://github.com/Altinity/terraform-provider-altinitycloud/pull/128).
- Bump github.com/Yamashou/gqlgenc to `0.28.2` [#132](https://github.com/Altinity/terraform-provider-altinitycloud/pull/132).

## [0.4.4](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.3...v0.4.4)
### Fixed
- Sort k8s environments `node_groups.selector` after API responses to match script order [#cb3be84](https://github.com/Altinity/terraform-provider-altinitycloud/commit/cb3be84).
- Sort k8s environments `node_groups.tolerations` after API responses to match script order [#c7b409f](https://github.com/Altinity/terraform-provider-altinitycloud/commit/c7b409f).

## [0.4.3](https://github.com/Altinity/terraform-provider-altinitycloud/compare/v0.4.2...v0.4.3)
### Changed
- Remove `force_destroy` from hcloud examples [#ab6a199](https://github.com/Altinity/terraform-provider-altinitycloud/commit/ab6a199).
- Bump github.com/Yamashou/gqlgenc to `0.27.3` [#126](https://github.com/Altinity/terraform-provider-altinitycloud/pull/126).

### Fixed
- Make node group zones property required for k8s environments [#a3e6149](https://github.com/Altinity/terraform-provider-altinitycloud/commit/a3e6149).

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
