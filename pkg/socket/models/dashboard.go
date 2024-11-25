package models

type Dashboard struct {
	Total  `json:"total"`
	Charts `json:"charts"`
}

type Total struct {
	Packets  int    `json:"packets,omitempty"`
	Traffics string `json:"traffics,omitempty"`
	Sessions int    `json:"sessions,omitempty"`
}

type Charts struct {
	ApplicationLayer interface{} `json:"application_layer,omitempty"`
	TransportLayer   interface{} `json:"transport_layer,omitempty"`
	Traffic          interface{} `json:"traffic,omitempty"`
	Application      interface{} `json:"application,omitempty"`
}
