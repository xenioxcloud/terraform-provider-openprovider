package openprovider

import (
	"encoding/json"
	"fmt"
)

// ContactName represents name fields.
type ContactName struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Prefix     string `json:"prefix,omitempty"`
	Initials   string `json:"initials,omitempty"`
}

// ContactPhone represents a phone number.
type ContactPhone struct {
	CountryCode      string `json:"country_code"`
	AreaCode         string `json:"area_code"`
	SubscriberNumber string `json:"subscriber_number"`
}

// ContactAddress represents an address.
type ContactAddress struct {
	Street  string `json:"street"`
	Number  string `json:"number"`
	Zipcode string `json:"zipcode"`
	City    string `json:"city"`
	State   string `json:"state,omitempty"`
	Country string `json:"country"`
}

// Contact represents an Openprovider contact/customer.
type Contact struct {
	Handle      string         `json:"handle"`
	ID          int            `json:"id"`
	CompanyName string         `json:"company_name,omitempty"`
	Name        ContactName    `json:"name"`
	Phone       ContactPhone   `json:"phone"`
	Fax         *ContactPhone  `json:"fax,omitempty"`
	Address     ContactAddress `json:"address"`
	Email       string         `json:"email"`
	Locale      string         `json:"locale,omitempty"`
}

type contactGetResponse struct {
	Data Contact `json:"data"`
}

type contactCreateResponse struct {
	Data struct {
		Handle string `json:"handle"`
	} `json:"data"`
}

// GetContact retrieves a contact by handle.
func (c *Client) GetContact(handle string) (*Contact, error) {
	respBody, err := c.doRequest("GET", "/customers/"+handle, nil)
	if err != nil {
		return nil, fmt.Errorf("getting contact %s: %w", handle, err)
	}

	var result contactGetResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding contact response: %w", err)
	}

	return &result.Data, nil
}

// CreateContact creates a new contact and returns the handle.
func (c *Client) CreateContact(contact *Contact) (string, error) {
	body := map[string]any{
		"company_name": contact.CompanyName,
		"name": map[string]string{
			"first_name": contact.Name.FirstName,
			"last_name":  contact.Name.LastName,
			"prefix":     contact.Name.Prefix,
			"initials":   contact.Name.Initials,
		},
		"phone": map[string]string{
			"country_code":      contact.Phone.CountryCode,
			"area_code":         contact.Phone.AreaCode,
			"subscriber_number": contact.Phone.SubscriberNumber,
		},
		"address": map[string]string{
			"street":  contact.Address.Street,
			"number":  contact.Address.Number,
			"zipcode": contact.Address.Zipcode,
			"city":    contact.Address.City,
			"state":   contact.Address.State,
			"country": contact.Address.Country,
		},
		"email":  contact.Email,
		"locale": contact.Locale,
	}

	respBody, err := c.doRequest("POST", "/customers", body)
	if err != nil {
		return "", fmt.Errorf("creating contact: %w", err)
	}

	var result contactCreateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decoding create contact response: %w", err)
	}

	return result.Data.Handle, nil
}

// UpdateContact updates an existing contact.
func (c *Client) UpdateContact(handle string, contact *Contact) error {
	body := map[string]any{
		"company_name": contact.CompanyName,
		"name": map[string]string{
			"first_name": contact.Name.FirstName,
			"last_name":  contact.Name.LastName,
			"prefix":     contact.Name.Prefix,
			"initials":   contact.Name.Initials,
		},
		"phone": map[string]string{
			"country_code":      contact.Phone.CountryCode,
			"area_code":         contact.Phone.AreaCode,
			"subscriber_number": contact.Phone.SubscriberNumber,
		},
		"address": map[string]string{
			"street":  contact.Address.Street,
			"number":  contact.Address.Number,
			"zipcode": contact.Address.Zipcode,
			"city":    contact.Address.City,
			"state":   contact.Address.State,
			"country": contact.Address.Country,
		},
		"email":  contact.Email,
		"locale": contact.Locale,
	}

	_, err := c.doRequest("PUT", "/customers/"+handle, body)
	if err != nil {
		return fmt.Errorf("updating contact %s: %w", handle, err)
	}

	return nil
}

// DeleteContact deletes a contact by handle.
func (c *Client) DeleteContact(handle string) error {
	_, err := c.doRequest("DELETE", "/customers/"+handle, nil)
	if err != nil {
		return fmt.Errorf("deleting contact %s: %w", handle, err)
	}
	return nil
}
