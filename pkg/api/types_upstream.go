package api

// Upstream represents an APISIX upstream resource.
type Upstream struct {
	ID            *string                `json:"id,omitempty"`
	Name          *string                `json:"name,omitempty"`
	Desc          *string                `json:"desc,omitempty"`
	Type          *string                `json:"type,omitempty"`
	Nodes         map[string]interface{} `json:"nodes,omitempty"`
	ServiceName   *string                `json:"service_name,omitempty"`
	DiscoveryType *string                `json:"discovery_type,omitempty"`
	HashOn        *string                `json:"hash_on,omitempty"`
	Key           *string                `json:"key,omitempty"`
	Checks        map[string]interface{} `json:"checks,omitempty"`
	Retries       *int                   `json:"retries,omitempty"`
	RetryTimeout  *float64               `json:"retry_timeout,omitempty"`
	Timeout       *UpstreamTimeout       `json:"timeout,omitempty"`
	PassHost      *string                `json:"pass_host,omitempty"`
	UpstreamHost  *string                `json:"upstream_host,omitempty"`
	Scheme        *string                `json:"scheme,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	KeepalivePool map[string]interface{} `json:"keepalive_pool,omitempty"`
	TLS           map[string]interface{} `json:"tls,omitempty"`
	Status        *int                   `json:"status,omitempty"`
	CreateTime    *int64                 `json:"create_time,omitempty"`
	UpdateTime    *int64                 `json:"update_time,omitempty"`
}

// UpstreamTimeout defines timeout settings for an upstream.
type UpstreamTimeout struct {
	Connect *float64 `json:"connect,omitempty"`
	Send    *float64 `json:"send,omitempty"`
	Read    *float64 `json:"read,omitempty"`
}
