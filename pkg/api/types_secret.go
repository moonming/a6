package api

// Secret represents an APISIX secret manager configuration.
type Secret struct {
	ID              *string                `json:"id,omitempty"`
	URI             *string                `json:"uri,omitempty"`
	Prefix          *string                `json:"prefix,omitempty"`
	Token           *string                `json:"token,omitempty"`
	Namespace       *string                `json:"namespace,omitempty"`
	AccessKeyID     *string                `json:"access_key_id,omitempty"`
	SecretAccessKey *string                `json:"secret_access_key,omitempty"`
	Region          *string                `json:"region,omitempty"`
	EndpointURL     *string                `json:"endpoint_url,omitempty"`
	AuthConfig      map[string]interface{} `json:"auth_config,omitempty"`
	AuthFile        *string                `json:"auth_file,omitempty"`
	CreateTime      *int64                 `json:"create_time,omitempty"`
	UpdateTime      *int64                 `json:"update_time,omitempty"`
}
