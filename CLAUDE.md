# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Custom Terraform provider for Openprovider domain management. Built with the Terraform Plugin Framework in Go.

## Architecture

- `internal/openprovider/` — API client (auth, domain lookup, nameserver/DNSSEC/settings updates, contact CRUD)
- `internal/provider/` — Terraform provider, resources, and data sources
- `main.go` — Provider entry point

## Resources

- `openprovider_domain` — Combined resource: nameservers, DNSSEC, lock, auto-renewal (preferred)
- `openprovider_contact` — Contact handle CRUD
- `openprovider_domain_nameservers` — Nameservers only (legacy, use `openprovider_domain`)
- `openprovider_domain_dnssec` — DNSSEC only (legacy, use `openprovider_domain`)
- `openprovider_domain_settings` — Lock + auto-renewal only (legacy, use `openprovider_domain`)

## Data Sources

- `openprovider_domain` — Read domain info (status, expiry, DNSSEC status)

## Commands

```bash
go build -o terraform-provider-openprovider
go test ./...
```

## Publishing

Uses GoReleaser + GitHub Actions. Tag a version to trigger a release:
```bash
git tag v0.x.0 && git push origin v0.x.0
```

## Sandbox

Set `sandbox = true` in provider config to use `api.sandbox.openprovider.nl`. The `.nl` TLD does not support domain locking — leave `is_locked` unset for `.nl` domains.
