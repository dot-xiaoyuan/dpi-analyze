package models

type IPDetail struct {
	Detail   interface{} `json:"detail"`
	History  `json:"history"`
	Features any `json:"features"`
}

type History struct {
	TTL interface{} `json:"ttl"`
	Mac interface{} `json:"mac"`
	Ua  interface{} `json:"ua"`
}
