---
page_title: "Provider: Domeneshop"
subcategory: ""
description: |-
  Terraform provider for Domeneshop (Domainnameshop).
---

# Domeneshop Provider

Terraform provider for Domeneshop (Domainnameshop).

## Example Usage

```terraform
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
```

## Schema

### Optional

- **host** (String) The base URL of the Domeneshop API. Defaults to `https://api.domeneshop.no/v0`. This can also be set with the `DOMENESHOP_HOST` environment variable. Mainly useful for testing.
- **secret** (String, Sensitive) A Domeneshop API secret. This can also be set with the `DOMENESHOP_SECRET` environment variable.
- **token** (String, Sensitive) A Domeneshop API token. This can also be set with the `DOMENESHOP_TOKEN` environment variable.