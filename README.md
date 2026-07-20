![](https://domene.shop/svg/logo-no.svg)

# Terraform Provider Domeneshop

[![Tests](https://github.com/1ARdotNO/terraform-provider-domeneshop/actions/workflows/test.yml/badge.svg)](https://github.com/1ARdotNO/terraform-provider-domeneshop/actions/workflows/test.yml)
[![CodeQL](https://github.com/1ARdotNO/terraform-provider-domeneshop/actions/workflows/codeql.yml/badge.svg)](https://github.com/1ARdotNO/terraform-provider-domeneshop/actions/workflows/codeql.yml)
[![govulncheck](https://github.com/1ARdotNO/terraform-provider-domeneshop/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/1ARdotNO/terraform-provider-domeneshop/actions/workflows/govulncheck.yml)

Terraform provider for the [Domeneshop](https://domene.shop/) (domainnameshop) API:
manage DNS records and HTTP forwards, and look up domains.

This is a maintained fork of the original
[innovationnorway/terraform-provider-domeneshop](https://github.com/innovationnorway/terraform-provider-domeneshop)
(unmaintained since 2021), with support for modern Terraform versions and
Apple Silicon (`darwin/arm64`) builds.

Available in the [Terraform Registry](https://registry.terraform.io/providers/1ARdotNO/domeneshop/latest):

```hcl
terraform {
  required_providers {
    domeneshop = {
      source = "1ARdotNO/domeneshop"
    }
  }
}
```

## Requirements

-	[Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5
-	[Go](https://go.dev/doc/install) >= 1.25 (to build the provider)

## Using the provider

Credentials are an API token/secret pair, created in the
[Domeneshop admin panel](https://domene.shop/admin?view=api).

```hcl
variable "domeneshop_token" {
  type      = string
  sensitive = true
}

variable "domeneshop_secret" {
  type      = string
  sensitive = true
}

provider "domeneshop" {
  token  = var.domeneshop_token
  secret = var.domeneshop_secret
}

data "domeneshop_domains" "example" {
  domain = "example.com"
}

resource "domeneshop_record" "example" {
  domain_id = data.domeneshop_domains.example.domains[0].id
  host      = "foo"
  type      = "A"
  data      = "192.0.2.56"
  ttl       = 300
}
```

See the [provider documentation](https://registry.terraform.io/providers/1ARdotNO/domeneshop/latest/docs)
for all resources and data sources.

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:
```sh
$ go install
```

## Developing the Provider

To compile the provider, run `go install`. This will build the provider and
put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

### Testing

The acceptance tests run hermetically by default: they start an in-process
mock of the Domeneshop API that validates every request and response against
the official [API documentation](https://api.domeneshop.no/docs/) (vendored as
`internal/provider/testdata/openapi.json`). No credentials or network access
are needed:

```sh
$ TF_ACC=1 go test ./internal/provider/
```

Traffic that does not match the API documentation fails the tests with a
contract-violation error. Known divergences between the documentation and the
generated API client are reported as `[api-contract]` notes in the test
output.

To run the same suite against the real API instead, set:

- `DOMENESHOP_TOKEN`
- `DOMENESHOP_SECRET`
- `DOMENESHOP_DOMAIN` (a domain on the account the tests may write to)

*Note:* against the real API, the acceptance tests create real resources.

### CI and dependencies

Every pull request must pass the required checks (build, acceptance tests
against Terraform 1.5/1.6/1.9, `tfproviderlint`, CodeQL and `govulncheck`)
before it can merge. [Renovate](https://docs.renovatebot.com/) keeps
dependencies current and auto-merges updates that keep the checks green.

## Releasing

Releases are built and signed by the `release` workflow when a `v*` tag is
pushed, and published to the Terraform Registry.
