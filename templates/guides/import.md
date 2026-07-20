---
page_title: "Importing existing resources"
subcategory: ""
description: |-
  Import DNS records and HTTP forwards that already exist in your Domeneshop account.
---

# Importing existing resources

If your domains already have DNS records or HTTP forwards managed outside
Terraform, you can adopt them without recreating anything using Terraform's
native [`import` blocks](https://developer.hashicorp.com/terraform/language/import)
and configuration generation (Terraform >= 1.5).

## Import IDs

| Resource             | Import ID            | Example        |
|----------------------|----------------------|----------------|
| `domeneshop_record`  | `domain_id/record_id`| `1234567/98765`|
| `domeneshop_forward` | `domain_id/host`     | `1234567/www`  |

Domain IDs can be listed with the `domeneshop_domains` data source, or
directly from the API:

```shell
curl -s -u "$DOMENESHOP_TOKEN:$DOMENESHOP_SECRET" \
  https://api.domeneshop.no/v0/domains | jq '.[] | {id, domain}'
```

## Importing with configuration generation

Write an `import` block per existing resource:

```terraform
import {
  to = domeneshop_record.www
  id = "1234567/98765"
}
```

Then let Terraform generate matching configuration from the provider schema —
this fills in every attribute (including `priority`, `weight`, `port` and the
CAA/TLSA fields) exactly as the API reports them:

```shell
terraform plan -generate-config-out=generated.tf
terraform apply
```

Review `generated.tf`, move the blocks where you want them, and drop the
`import` blocks once applied.

## Generating import blocks for a whole account

The repository ships a helper script,
[`examples/import/generate-imports.sh`](https://github.com/1ARdotNO/terraform-provider-domeneshop/blob/main/examples/import/generate-imports.sh),
that emits deterministic import blocks for every DNS record and HTTP forward
on your account:

```shell
export DOMENESHOP_TOKEN=... DOMENESHOP_SECRET=...
./generate-imports.sh > imports.tf
terraform plan -generate-config-out=generated.tf
```

## Importing with the CLI

The classic one-at-a-time form works as well:

```shell
terraform import domeneshop_record.www 1234567/98765
terraform import domeneshop_forward.blog 1234567/blog
```

Unlike `import` blocks, `terraform import` requires the matching `resource`
block to already exist in configuration.
