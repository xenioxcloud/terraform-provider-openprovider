# Terraform Provider for Openprovider

Manage domains registered at [Openprovider](https://www.openprovider.com) with Terraform — nameservers, DNSSEC, transfer lock, auto-renewal, and contact handles.

## Usage

```hcl
terraform {
  required_providers {
    openprovider = {
      source  = "xenioxcloud/openprovider"
      version = ">= 0.2.0"
    }
  }
}

provider "openprovider" {
  username = var.openprovider_username
  password = var.openprovider_password
  # sandbox = true  # use the sandbox environment for testing
}
```

### Manage a domain

```hcl
resource "openprovider_domain" "example" {
  domain    = "example.nl"
  dnssec    = false
  autorenew = "on"

  nameservers = [
    { name = "erin.ns.cloudflare.com" },
    { name = "hans.ns.cloudflare.com" },
  ]
}
```

### Import an existing domain

```bash
terraform import openprovider_domain.example example.nl
```

### Manage a contact handle

```hcl
resource "openprovider_contact" "example" {
  company_name            = "Acme B.V."
  first_name              = "John"
  last_name               = "Doe"
  email                   = "john@example.nl"
  phone_country_code      = "+31"
  phone_area_code         = "70"
  phone_subscriber_number = "1234567"
  street                  = "Hoofdstraat"
  number                  = "1"
  zipcode                 = "2511AA"
  city                    = "Den Haag"
  country                 = "NL"
}
```

### Read domain info

```hcl
data "openprovider_domain" "example" {
  domain = "example.nl"
}

output "expiry" {
  value = data.openprovider_domain.example.expiration_date
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `openprovider_domain` | Manage nameservers, DNSSEC, lock, and auto-renewal |
| `openprovider_contact` | Manage contact handles (CRUD + import) |
| `openprovider_domain_nameservers` | Manage nameservers only (use `openprovider_domain` instead) |
| `openprovider_domain_dnssec` | Manage DNSSEC only (use `openprovider_domain` instead) |
| `openprovider_domain_settings` | Manage lock + auto-renewal only (use `openprovider_domain` instead) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `openprovider_domain` | Read domain info (status, expiry, DNSSEC) |

## Authentication

Set credentials via provider config or environment variables:

| Provider Attribute | Environment Variable | Description |
|--------------------|---------------------|-------------|
| `username` | `OPENPROVIDER_USERNAME` | API username |
| `password` | `OPENPROVIDER_PASSWORD` | API password |
| `sandbox` | `OPENPROVIDER_SANDBOX` | Use sandbox environment (`true`/`false`) |

## Development

```bash
# Build
go build -o terraform-provider-openprovider

# Test locally with dev overrides in ~/.terraformrc
provider_installation {
  dev_overrides {
    "xenioxcloud/openprovider" = "/path/to/terraform-provider-openprovider"
  }
  direct {}
}
```

## Releasing

Tag a version and push — GitHub Actions builds and publishes via GoReleaser:

```bash
git tag v0.3.0
git push origin v0.3.0
```
