package api

// SSL represents an APISIX SSL certificate resource.
type SSL struct {
	ID           *string           `json:"id,omitempty"`
	Cert         *string           `json:"cert,omitempty"`
	Key          *string           `json:"key,omitempty"`
	Certs        []string          `json:"certs,omitempty"`
	Keys         []string          `json:"keys,omitempty"`
	SNI          *string           `json:"sni,omitempty"`
	SNIs         []string          `json:"snis,omitempty"`
	Client       *SSLClient        `json:"client,omitempty"`
	Type         *string           `json:"type,omitempty"`
	Status       *int              `json:"status,omitempty"`
	SSLProtocols []string          `json:"ssl_protocols,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	CreateTime   *int64            `json:"create_time,omitempty"`
	UpdateTime   *int64            `json:"update_time,omitempty"`
}

// SSLClient defines mTLS client verification settings.
type SSLClient struct {
	CA    *string `json:"ca,omitempty"`
	Depth *int    `json:"depth,omitempty"`
}
