package main

// CloudflareCredentials represents the credentials of type cloduflare as defined in the server config and passed to this trusted image
type CloudflareCredentials struct {
	Name                 string                                    `json:"name,omitempty"`
	Type                 string                                    `json:"type,omitempty"`
	AdditionalProperties CloudflareCredentialsAdditionalProperties `json:"additionalProperties,omitempty"`
}

// CloudflareCredentialsAdditionalProperties contains the non standard fields for this type of credentials
type CloudflareCredentialsAdditionalProperties struct {
	APIEmail string `json:"email,omitempty"`
	APIKey   string `json:"key,omitempty"`
}
