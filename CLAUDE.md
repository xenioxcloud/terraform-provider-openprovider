# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Custom Terraform provider for Openprovider domain management. Built with the Terraform Plugin Framework in Go.

## Architecture

- `internal/openprovider/` — API client (auth, domain lookup, nameserver/DNSSEC updates)
- `internal/provider/` — Terraform provider, resources, and data sources
- `main.go` — Provider entry point

## Resources

- `openprovider_domain_nameservers` — Manage nameservers on a domain
- `openprovider_domain_dnssec` — Enable/disable DNSSEC on a domain

## Data Sources

- `openprovider_domain` — Read domain info (status, expiry, DNSSEC status)

## Commands

```bash
go build -o terraform-provider-openprovider
go test ./...
```

## Publishing

Uses GoReleaser + GitHub Actions to publish to the Terraform Registry under `xenioxcloud/openprovider`.
