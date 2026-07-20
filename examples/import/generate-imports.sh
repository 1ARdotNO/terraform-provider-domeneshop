#!/usr/bin/env bash
# Generate Terraform import blocks for every DNS record and HTTP forward in a
# Domeneshop account, for use with Terraform >= 1.5 configuration generation:
#
#   export DOMENESHOP_TOKEN=... DOMENESHOP_SECRET=...
#   ./generate-imports.sh > imports.tf
#   terraform plan -generate-config-out=generated.tf
#
# Resource names are deterministic (domain, host, type and record ID), so the
# script can be re-run without producing different names. Requires curl and jq.
set -euo pipefail

API_URL="${DOMENESHOP_HOST:-https://api.domeneshop.no/v0}"
: "${DOMENESHOP_TOKEN:?DOMENESHOP_TOKEN must be set}"
: "${DOMENESHOP_SECRET:?DOMENESHOP_SECRET must be set}"

api() {
  curl --silent --user "$DOMENESHOP_TOKEN:$DOMENESHOP_SECRET" "$API_URL$1"
}

# Keep only elements of a JSON array; error responses are objects and yield
# nothing (e.g. domains without DNS service).
array_elements() {
  jq -c 'if type == "array" then .[] else empty end'
}

# Turn a domain or host name into a valid, deterministic Terraform identifier.
ident() {
  printf '%s' "$1" | sed -e 's/@/apex/g' -e 's/\*/wildcard/g' -e 's/[^A-Za-z0-9]/_/g'
}

domains=$(api /domains)
if [ "$(jq -r 'type' <<<"$domains")" != "array" ]; then
  echo "error listing domains: $domains" >&2
  exit 1
fi

array_elements <<<"$domains" | while read -r domain; do
  domain_id=$(jq -r '.id' <<<"$domain")
  domain_ident=$(ident "$(jq -r '.domain' <<<"$domain")")

  api "/domains/$domain_id/dns" | array_elements | while read -r record; do
    record_id=$(jq -r '.id' <<<"$record")
    host_ident=$(ident "$(jq -r '.host' <<<"$record")")
    record_type=$(jq -r '.type' <<<"$record")
    printf 'import {\n  to = domeneshop_record.%s_%s_%s_%s\n  id = "%s/%s"\n}\n\n' \
      "$domain_ident" "$host_ident" "$record_type" "$record_id" "$domain_id" "$record_id"
  done

  api "/domains/$domain_id/forwards/" | array_elements | while read -r forward; do
    host=$(jq -r '.host' <<<"$forward")
    printf 'import {\n  to = domeneshop_forward.%s_%s\n  id = "%s/%s"\n}\n\n' \
      "$domain_ident" "$(ident "$host")" "$domain_id" "$host"
  done
done
