package api

// GlobalRule represents an APISIX global rule resource.
type GlobalRule struct {
	ID         *string                `json:"id,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty"`
}
