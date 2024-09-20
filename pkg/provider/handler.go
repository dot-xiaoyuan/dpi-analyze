package provider

import "encoding/json"

type Handler interface {
	Handle(data json.RawMessage) []byte
}

type Request struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
