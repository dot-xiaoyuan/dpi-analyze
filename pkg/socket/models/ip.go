package models

type IPDetail struct {
	Detail      interface{} `json:"detail"`
	History     History     `json:"history"`
	Features    any         `json:"features"`
	Devices     any         `json:"devices"`
	DevicesLogs any         `json:"devices_logs"`
}

type History struct {
	TTL interface{} `json:"ttl"`
	Mac interface{} `json:"mac"`
	Ua  interface{} `json:"ua"`
}
