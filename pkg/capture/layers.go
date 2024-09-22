package capture

// Ethernet 以太网
type Ethernet struct {
	SrcMac string `json:"src_mac"`
	DstMac string `json:"dst_mac"`
}

// Internet 网络层
type Internet struct {
	DstIP string `json:"dst_ip"`
	TTL   uint8  `json:"ttl"`
}

type Transmission struct {
	UpStream   int64 `json:"up_stream"`
	DownStream int64 `json:"down_stream"`
}

type LayerMap interface {
	Update(i interface{})
}
