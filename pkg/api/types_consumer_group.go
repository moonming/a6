package api

// ConsumerGroup represents an APISIX consumer group resource.
type ConsumerGroup struct {
	ID         *string                `json:"id,omitempty"`
	Name       *string                `json:"name,omitempty"`
	Desc       *string                `json:"desc,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty"`
}
