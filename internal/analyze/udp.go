package analyze

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/statictics"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	LayerTypeQUIC = gopacket.RegisterLayerType(147, gopacket.LayerTypeMetadata{
		Name:    "QUIC",
		Decoder: gopacket.DecodeFunc(decodeQUIC),
	})
	LayerTypeTFTP = gopacket.RegisterLayerType(148, gopacket.LayerTypeMetadata{
		Name:    "TFTP",
		Decoder: gopacket.DecodeFunc(decodeTFTP),
	})
	LayerTypeSNMP = gopacket.RegisterLayerType(149, gopacket.LayerTypeMetadata{
		Name:    "SNMP",
		Decoder: gopacket.DecodeFunc(decodeSNMP),
	})
	LayerTypeMDNS = gopacket.RegisterLayerType(150, gopacket.LayerTypeMetadata{
		Name:    "MDNS",
		Decoder: gopacket.DecodeFunc(decodeMDNS),
	})
)

func CheckUDP(userIP, tranIP string, udp *layers.UDP) gopacket.LayerType {
	if udp.DstPort == 53 || udp.SrcPort == 53 {
		pushTask(userIP, tranIP, types.DNS)
		return layers.LayerTypeDNS
	}
	if udp.SrcPort == 67 && udp.DstPort == 68 {
		pushTask(userIP, tranIP, types.DHCP)
		return layers.LayerTypeDHCPv4
	}
	if udp.SrcPort == 546 && udp.DstPort == 547 {
		pushTask(userIP, tranIP, types.DHCPv6)
		return layers.LayerTypeDHCPv6
	}
	if udp.SrcPort == 123 || udp.DstPort == 123 {
		pushTask(userIP, tranIP, types.NTP)
		return layers.LayerTypeNTP
	}
	if udp.SrcPort == 5353 || udp.DstPort == 5353 {
		pushTask(userIP, tranIP, types.MDNS)
		return LayerTypeMDNS
	}
	//if len(udp.Payload) > 5 && (udp.Payload[0]&0b11000000 == 0b11000000 || udp.Payload[0]&0b10000000 == 0) {
	//	pushTask(userIP, tranIP, types.QUIC)
	//	return LayerTypeQUIC
	//}
	if udp.SrcPort == 69 || udp.DstPort == 69 {
		pushTask(userIP, tranIP, types.TFTP)
		return LayerTypeTFTP
	}
	if udp.SrcPort == 161 || udp.DstPort == 161 {
		pushTask(userIP, tranIP, types.SNMP)
		return LayerTypeSNMP
	}
	if udp.SrcPort == 5353 || udp.DstPort == 5353 {
		pushTask(userIP, tranIP, types.MDNS)
		return LayerTypeMDNS
	}
	if udp.SrcPort == 4789 || udp.DstPort == 4789 {
		pushTask(userIP, tranIP, types.VXLAN)
		return layers.LayerTypeVXLAN
	}
	if udp.SrcPort == 5060 || udp.DstPort == 5060 {
		pushTask(userIP, tranIP, types.SIP)
		return layers.LayerTypeSIP
	}
	if udp.SrcPort == 6343 || udp.DstPort == 6343 {
		pushTask(userIP, tranIP, types.SFlow)
		return layers.LayerTypeSFlow
	}
	if udp.SrcPort == 6081 || udp.DstPort == 6081 {
		pushTask(userIP, tranIP, types.Geneve)
		return layers.LayerTypeGeneve
	}
	if udp.SrcPort == 3784 || udp.DstPort == 3784 {
		pushTask(userIP, tranIP, types.BFD)
		return layers.LayerTypeBFD
	}
	if udp.SrcPort == 2152 || udp.DstPort == 2152 {
		pushTask(userIP, tranIP, types.GTPv1U)
		return layers.LayerTypeGTPv1U
	}
	if udp.SrcPort == 623 || udp.DstPort == 623 {
		pushTask(userIP, tranIP, types.RMCP)
		return layers.LayerTypeRMCP
	}
	if udp.SrcPort == 1812 || udp.DstPort == 1812 {
		pushTask(userIP, tranIP, types.Radius)
		return layers.LayerTypeRADIUS
	}
	return 0
}

func decodeQUIC(data []byte, p gopacket.PacketBuilder) error {
	return nil
}

func decodeTFTP(data []byte, p gopacket.PacketBuilder) error {
	return nil
}

func decodeSNMP(data []byte, p gopacket.PacketBuilder) error {
	return nil
}

func decodeMDNS(data []byte, p gopacket.PacketBuilder) error {
	return nil
}

func pushTask(userIP, tranIP string, featureType types.FeatureType) {
	_ = ants.Submit(func() {
		member.Increment(types.Feature{
			IP:    userIP,
			Field: featureType,
			Value: tranIP,
		})
		statictics.ApplicationLayer.Increment(string(featureType))
	})
}
