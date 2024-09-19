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

type LayerMap interface {
	Update(i interface{})
	QueryAll() ([]byte, error)
}
