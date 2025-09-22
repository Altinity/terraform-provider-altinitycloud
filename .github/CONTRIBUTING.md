# Contributing to AltinityCloud Terraform Provider

> Learn about our [Commitment to Open Source](https://altinity.com/ecosystem/).

Hi! We are really excited that you are interested in contributing to AltinityCloud Terraform Provider, and we really appreciate your commitment. Before submitting your contribution, please make sure to take a moment and read through the following guidelines:

- [Code of Conduct](./CODE_OF_CONDUCT.md)
- [Issue Reporting Guidelines](#issue-reporting-guidelines)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Development Setup](#development-setup)
- [Commands](#commands)
- [Project Structure](#project-structure)
- [Release Process](#releases-process)

## Issue Reporting Guidelines

- Always use [GitHub issues](https://github.com/altinity/terraform-provider-altinitycloud/issues/new/choose) to create new issues and select the corresponding issue template.

## Pull Request Guidelines

- Checkout a topic branch from a base branch, e.g. `main`, and merge back against that branch.

- [Make sure to tick the "Allow edits from maintainers" box](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork). This allows us to directly make minor edits / refactors and saves a lot of time.

- Add accompanying documentation, usage samples & test cases
- Add/update demo files to showcase your changes.
- Use existing resources as templates and ensure that each property has a corresponding `description` field.
- Each PR should be linked with an issue, use [GitHub keywords](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests) for that.
- Be sure to follow up project code styles (`$ make fmt`)

- If adding a new feature:
  - Provide a convincing reason to add this feature. Ideally, you should open a "feature request" issue first and have it approved before working on it (it should has the label "state: confirmed")
  - Each new feature should be linked to an existing issue

- If fixing a bug:
  - Provide a detailed description of the bug in the PR. A working demo would be great!

- It's OK to have multiple small commits as you work on the PR - GitHub can automatically squash them before merging.

- Make sure tests pass!

- Commit messages must follow the [semantic commit messages](https://gist.github.com/joshbuchea/6f47e86d2510bce28f8e7f42ae84c716) so that changelogs can be automatically generated.

## Development Setup

The development branch is `main`. This is the branch that all pull requests should be made against.

> ⚠️ Before you start, is recommended to have a good understanding on how the provider works, the resources it has and its different configurations.

### Pre-requirements
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

> 💡 We recommend to use [tfswitch](https://tfswitch.warrensbox.com/) to easily manage different Terraform versions in your local environment.

### Getting Started

After you have installed Terraform and Go, you can clone the repository and start working on the `main` branch.

1. [Fork](https://help.github.com/articles/fork-a-repo/) this repository to your own GitHub account and then [clone](https://help.github.com/articles/cloning-a-repository/) it to your local device.
  ```sh
  # If you don't need the whole git history, you can clone with depth 1 to reduce the download size:
  $ git clone --depth=1 git@github.com:altinity/terraform-provider-altinitycloud.git
  ```

2. Navigate into the project and create a new branch:
  ```sh
  cd terraform-provider-altinitycloud && git checkout -b MY_BRANCH_NAME
  ```

3. Download go packages
  ```sh
  $ go mod tidy
  ```

4. Run build process to ensure that everything is on place
  ```sh
  $ make build
  ```

**You are ready to go, take a look at the [`make local`](#make-local) command so you can easily test your changes locally.** 🙌

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Commands

In order to facilitate the development process, we have a `GNUmakefile` with a few scripts that are used to build, test and run working examples.

> Run `make help` to see all available local commands

### Available Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the provider binary |
| `make local` | Build and set up for local testing |
| `make docs` | Generate provider documentation |
| `make fmt` | Format Terraform and Go code |
| `make sdk` | Re-sync SDK client and models |
| `make gen` | Run SDK generation and docs generation |
| `make test` | Run Go unit tests with coverage |
| `make testacc` | Run acceptance tests |
| `make lint` | Run Go linter |
| `make bump` | Bump version tags (use with `type=major\|minor\|patch`) |

### `make local`
This command builds your local project and sets up the provider binary to be used in a `local` directory.
```sh
$ make local
```

After you run the `local` command, you should be able to run terraform with the `local` sample project:
  1. Navigate to the: `$ cd local`
  1. Run `terraform plan` and/or `terraform apply`

Another useful tip for local testing/development is to override the `local/versions.tf` with a configuration pointing to your development target environments (local, dev, stg):

  ```tf
  terraform {
    required_providers {
      altinitycloud = {
        source  = "local/altinity/altinitycloud"
        version = "0.0.1"
      }
    }
  }

  provider "altinitycloud" {
    # local settings
    api_token = "Get your token on ACM Dev"
    api_url   = "https://anywhere.altinity.cloud"
    # ca_crt = file("${path.module}/ca.crt")
  }
  ```

### `make sdk`
Re-sync local sdk client and models with the latest GraphQL spec (internal usage only)
```sh
$ make sdk
```

### `make fmt`
Format the code using native terraform and go formatting tools.
```sh
$ make fmt
```

### `make docs`
Use [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) to automatically generate resources documentation.
```sh
$ make docs
```

### `make test`
Run Go unit tests with coverage.
```sh
$ make test
```

### `make testacc`
Run acceptance tests. These are integration tests that test real infrastructure.
```sh
# Set up environment variables for testing
export ALTINITYCLOUD_API_TOKEN="your-token-here"
export ALTINITYCLOUD_API_URL="https://anywhere.altinity.cloud"
export ALTINITYCLOUD_TEST_ENV_PREFIX="your-prefix"

$ make testacc
```

### `make lint`
Run Go linter to check code quality.
```sh
$ make lint
```

### `make gen`
Run SDK generation and documentation generation in sequence.
```sh
$ make gen
```

## Project Structure

- `./.github`: contains github (and github actions) related files.

- `./internal`: contains all the code. This is the most important directory and where you will be working most of the time.
    - The directory `provider` contains everything related to the provider itself.
    - The `provider` file structure is organized by entities. For instance, in `env/aws` directory, you will find all the files related to AWS environments (resource, data source, schema, tests, helpers, etc.).
    - Then you will find `sdk` directory, which most of the content is automatically generated by the `make sdk` command.

> If you want to learn more about Terraform Providers and resources, [here](https://www.terraform.io/cdktf/concepts/providers-and-resources) is a good place to start.

- `./docs`: this directoy is automatically generated, please do not apply manual changes here. Run `make doc` in order to re-generate documentation.

- `./examples`: here you will find resource and provider examples (in `.tf` files). This will be used to generate docs and samples.

- `./templates`: contains the templates used to generate the documentation layouts.

- `./tools`: contains tools, scripts or utilities for development workflow.

## Release Process
The release process is automatically handled with [goreleaser](https://goreleaser.com/) and GitHub `release` action.
To trigger a new release you need to create a new git tag, using [SemVer](https://semver.org) pattern and then push it to the `main` branch.

Remember to create release candidate releases and spend some time testing in production before publishing a final version. You can also tag the release as "Pre-Release" on GitHub until you consider it mature enough.

### Creating a Release

1. **Bump Version**: Use `make bump type=[major|minor|patch]` to create a new version tag
2. **Push Tag**: Push the new tag to trigger the release workflow
   ```sh
   git push origin v1.2.3
   ```
3. **Monitor Release**: Check the GitHub Actions workflow for the release process
4. **Update Documentation**: Ensure all documentation is up to date
