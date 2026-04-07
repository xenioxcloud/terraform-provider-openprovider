terraform {
  required_providers {
    openprovider = {
      source = "registry.terraform.io/xenioxcloud/openprovider"
    }
  }
}

provider "openprovider" {
  # Set via OPENPROVIDER_USERNAME and OPENPROVIDER_PASSWORD env vars
  sandbox = true
}

data "openprovider_domain" "test" {
  domain = "xeniox-test.nl"
}

output "domain_info" {
  value = {
    id              = data.openprovider_domain.test.domain_id
    status          = data.openprovider_domain.test.status
    expiration_date = data.openprovider_domain.test.expiration_date
    dnssec_enabled  = data.openprovider_domain.test.is_dnssec_enabled
  }
}

resource "openprovider_domain_nameservers" "test" {
  domain = "xeniox-test.nl"

  nameservers = [
    { name = "anna.ns.cloudflare.com" },
    { name = "bob.ns.cloudflare.com" },
  ]
}

resource "openprovider_domain_dnssec" "test" {
  domain  = "xeniox-test.nl"
  enabled = false
}
