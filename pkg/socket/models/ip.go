package models

type IPDetail struct {
	TTLHistory interface{} `json:"ttl_history"`
	MacHistory interface{} `json:"mac_history"`
	UaHistory  interface{} `json:"ua_history"`
	Detail     interface{} `json:"detail"`
}
