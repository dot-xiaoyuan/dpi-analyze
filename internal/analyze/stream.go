package analyze

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"go.uber.org/zap"
	"sync"
)

type Stream struct {
	sync.Mutex
	client         StreamReader
	server         StreamReader
	tcpState       *reassembly.TCPSimpleFSM
	optChecker     reassembly.TCPOptionCheck
	net, transport gopacket.Flow
	fsmErr         bool
	Urls           []string
	ident          string
}

func (s *Stream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	// FSM
	if !s.tcpState.CheckState(tcp, dir) {
		//logger.Error("FSM %s: Packet rejected by FSM (state:%s)\n", zap.String("ident", s.ident), zap.String("state", s.tcpState.String()))
		//stats.rejectFsm++
		if !s.fsmErr {
			s.fsmErr = true
			//stats.rejectConnFsm++
		}
		//if !*ignorefsmerr {
		//	return false
		//}
	}
	// Options
	err := s.optChecker.Accept(tcp, ci, dir, nextSeq, start)
	if err != nil {
		zap.L().Error(fmt.Sprintf("OptionChecker %s: Packet rejected by OptionChecker: %s", s.ident, err))
		//stats.rejectOpt++
		//if !*nooptcheck {
		//	return false
		//}
	}
	// Checksum
	accept := true
	//if *checksum {
	//	c, err := tcp.ComputeChecksum()
	//	if err != nil {
	//		Error("ChecksumCompute", "%s: Got error computing checksum: %s\n", s.ident, err)
	//		accept = false
	//	} else if c != 0x0 {
	//		Error("Checksum", "%s: Invalid checksum: 0x%x\n", s.ident, c)
	//		accept = false
	//	}
	//}
	if !accept {
		//stats.rejectOpt++
	}
	return accept
}

func (s *Stream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	dir, start, end, skip := sg.Info()
	length, saved := sg.Lengths()
	// update stats
	sgStats := sg.Stats()
	if skip > 0 {
		// stats.missedBytes += skip // 丢失字节
	}
	//stats.sz += length - saved
	//stats.pkt += sgStats.Packets
	if sgStats.Chunks > 1 {
		//stats.reassembled++ // 重组包数
	}
	//stats.outOfOrderPackets += sgStats.QueuedPackets
	//stats.outOfOrderBytes += sgStats.QueuedBytes
	//if length > stats.biggestChunkBytes {
	//	stats.biggestChunkBytes = length // 最大区块字节数
	//}
	//if sgStats.Packets > stats.biggestChunkPackets {
	//	stats.biggestChunkPackets = sgStats.Packets // 最大区块包数
	//}
	if sgStats.OverlapBytes != 0 && sgStats.OverlapPackets == 0 {
		fmt.Printf("bytes:%d, pkts:%d\n", sgStats.OverlapBytes, sgStats.OverlapPackets)
		//panic("Invalid overlap")
	}
	//stats.overlapBytes += sgStats.OverlapBytes // 重叠字节数
	//stats.overlapPackets += sgStats.OverlapPackets // 重叠包数

	var ident string
	if dir == reassembly.TCPDirClientToServer {
		ident = fmt.Sprintf("%v %v(%s): ", s.net, s.transport, dir)
	} else {
		ident = fmt.Sprintf("%v %v(%s): ", s.net.Reverse(), s.transport.Reverse(), dir)
	}
	zap.L().Debug(fmt.Sprintf("%s: SG reassembled packet with %d bytes (start:%v,end:%v,skip:%d,saved:%d,nb:%d,%d,overlap:%d,%d)\n", ident, length, start, end, skip, saved, sgStats.Packets, sgStats.Chunks, sgStats.OverlapBytes, sgStats.OverlapPackets))
	//if skip == -1 && *allowMissingInit {
	//	// this is allowed
	//} else if skip != 0 {
	//	// Missing bytes in stream: do not even try to parse it
	//	return
	//}
	data := sg.Fetch(length)
	//dns := &layers.DNS{}
	//var decoded []gopacket.LayerType
	//if len(data) < 2 {
	//	if len(data) > 0 {
	//		sg.KeepFrom(0)
	//	}
	//	return
	//}
	//dnsSize := binary.BigEndian.Uint16(data[:2])
	//missing := int(dnsSize) - len(data[2:])
	//Debug("dnsSize: %d, missing: %d\n", dnsSize, missing)
	//if missing > 0 {
	//	Info("Missing some bytes: %d\n", missing)
	//	sg.KeepFrom(0)
	//	return
	//}
	//p := gopacket.NewDecodingLayerParser(layers.LayerTypeDNS, dns)
	//err := p.DecodeLayers(data[2:], &decoded)
	//if err != nil {
	//	Error("DNS-parser", "Failed to decode DNS: %v\n", err)
	//} else {
	//	Debug("DNS: %s\n", gopacket.LayerDump(dns))
	//}
	//if len(data) > 2+int(dnsSize) {
	//	sg.KeepFrom(2 + int(dnsSize))
	//}
	if length > 0 {
		if dir == reassembly.TCPDirClientToServer {
			s.client.bytes <- data
		} else {
			s.server.bytes <- data
		}
	}
}

func (s *Stream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	zap.L().Debug(fmt.Sprintf("%s: Connection closed", s.ident))
	close(s.client.bytes)
	close(s.server.bytes)
	return false
}
