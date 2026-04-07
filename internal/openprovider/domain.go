package openprovider

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// DomainName represents a domain name split into name and extension.
type DomainName struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
}

// Nameserver represents a nameserver entry.
type Nameserver struct {
	Name string `json:"name"`
	IP4  string `json:"ip4,omitempty"`
	IP6  string `json:"ip6,omitempty"`
}

// Domain represents an Openprovider domain.
type Domain struct {
	ID              int          `json:"id"`
	Domain          DomainName   `json:"domain"`
	Status          string       `json:"status"`
	ExpirationDate  string       `json:"expiration_date"`
	Nameservers     []Nameserver `json:"name_servers"`
	IsDNSSECEnabled bool         `json:"is_dnssec_enabled"`
}

// domainListResponse wraps the list domains API response.
type domainListResponse struct {
	Data struct {
		Results []Domain `json:"results"`
		Total   int      `json:"total"`
	} `json:"data"`
}

// domainGetResponse wraps the get domain API response.
type domainGetResponse struct {
	Data Domain `json:"data"`
}

// SearchDomains searches for domains matching the given name pattern.
func (c *Client) SearchDomains(domainName, extension string) ([]Domain, error) {
	params := url.Values{}
	if domainName != "" {
		params.Set("domain_name_pattern", domainName)
	}
	if extension != "" {
		params.Set("extension", extension)
	}
	params.Set("limit", "100")

	respBody, err := c.doRequest("GET", "/domains?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("searching domains: %w", err)
	}

	var result domainListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding domain search response: %w", err)
	}

	return result.Data.Results, nil
}

// GetDomain retrieves a domain by its ID.
func (c *Client) GetDomain(id int) (*Domain, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/domains/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("getting domain %d: %w", id, err)
	}

	var result domainGetResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding domain response: %w", err)
	}

	return &result.Data, nil
}

// UpdateDomainNameservers updates the nameservers for a domain.
func (c *Client) UpdateDomainNameservers(id int, nameservers []Nameserver) error {
	body := map[string]any{
		"name_servers": nameservers,
	}

	_, err := c.doRequest("PUT", fmt.Sprintf("/domains/%d", id), body)
	if err != nil {
		return fmt.Errorf("updating nameservers for domain %d: %w", id, err)
	}

	return nil
}

// UpdateDomainDNSSEC enables or disables DNSSEC for a domain.
func (c *Client) UpdateDomainDNSSEC(id int, enabled bool) error {
	body := map[string]any{
		"is_dnssec_enabled": enabled,
	}

	_, err := c.doRequest("PUT", fmt.Sprintf("/domains/%d", id), body)
	if err != nil {
		return fmt.Errorf("updating DNSSEC for domain %d: %w", id, err)
	}

	return nil
}

// FindDomainByName finds a domain by its full name (e.g., "example.nl").
// Returns the domain or an error if not found.
func (c *Client) FindDomainByName(fullDomain string) (*Domain, error) {
	// Split domain into name and extension
	name, ext := splitDomain(fullDomain)

	domains, err := c.SearchDomains(name, ext)
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		if d.Domain.Name == name && d.Domain.Extension == ext {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("domain %s not found", fullDomain)
}

// splitDomain splits "example.nl" into ("example", "nl") and "sub.example.co.uk" into ("sub.example", "co.uk").
func splitDomain(domain string) (string, string) {
	// Handle common multi-part TLDs
	multiPartTLDs := []string{".co.uk", ".co.nz", ".com.au", ".org.uk", ".net.au"}
	for _, tld := range multiPartTLDs {
		if len(domain) > len(tld) && domain[len(domain)-len(tld):] == tld {
			return domain[:len(domain)-len(tld)], tld[1:]
		}
	}

	// Default: split at last dot
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] == '.' {
			return domain[:i], domain[i+1:]
		}
	}
	return domain, ""
}
