package analyze

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Collections 单向流
type Collections struct {
	Urls                []string `bson:"urls"`
	SrcIP               string   `bson:"src_ip"`
	DstIP               string   `bson:"dst_ip"`
	Host                string   `bson:"host"`
	RejectFSM           int      `bson:"reject_fsm"` // FSM (Finite State Machine)有限状态机
	RejectConnFsm       int      `bson:"reject_conn_fsm"`
	RejectOpt           int      `bson:"reject_opt"`
	MissBytes           int      `bson:"miss_bytes"`
	Sz                  int      `bson:"sz"`
	Pkt                 int      `bson:"pkt"`
	Reassembled         int      `bson:"reassembled"`
	OutOfOrderPackets   int      `bson:"out_of_order_packets"`
	OutOfOrderBytes     int      `bson:"out_of_order_bytes"`
	BiggestChunkBytes   int      `bson:"biggest_chunk_bytes"`
	BiggestChunkPackets int      `bson:"biggest_chunk_packets"`
	OverlapBytes        int      `bson:"overlap_bytes"`
	OverlapPackets      int      `bson:"overlap_packets"`
	Application         string   `bson:"application"`
}

// Stream 流
type Stream struct {
	sync.Mutex
	SessionID      string `bson:"session_id"`
	StartTime      string `bson:"start_time"`
	EndTime        string `bson:"end_time"`
	Client         StreamReader
	Server         StreamReader
	TcpState       *reassembly.TCPSimpleFSM
	OptChecker     reassembly.TCPOptionCheck
	Net, Transport gopacket.Flow
	fsmErr         bool
	Ident          string `bson:"ident"`
	Collections
	PacketCount int8  `bson:"packet_count"`
	ByteCount   int16 `bson:"byte_count"`
}

func (s *Stream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	// FSM
	if !s.TcpState.CheckState(tcp, dir) {
		s.RejectFSM++
		if !s.fsmErr {
			s.fsmErr = true
			s.RejectConnFsm++
		}
		//return false
		//if !*ignorefsmerr {
		//	return false
		//}
	}
	// Options
	err := s.OptChecker.Accept(tcp, ci, dir, nextSeq, start)
	if err != nil {
		s.RejectOpt++
		//if !*nooptcheck {
		//	return false
		//}
	}
	// Checksum
	// TODO 是否需要校验 checksum
	return true
	//accept := true
	//c, err := tcp.ComputeChecksum()
	//if err != nil {
	//	zap.L().Debug("Failed to compute checksum", zap.Error(err))
	//	accept = false
	//} else if c != 0x0 {
	//	zap.L().Debug("Checksum Invalid checksum", zap.Uint16("checksum", c))
	//	accept = false
	//}
	//if !accept {
	//	s.RejectOpt++
	//}
	//return accept
}

func (s *Stream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	dir, start, end, skip := sg.Info()
	length, saved := sg.Lengths()
	// update stats
	sgStats := sg.Stats()
	if skip > 0 {
		s.MissBytes += skip // 丢失字节
	}
	s.Sz += length - saved
	s.Pkt += sgStats.Packets
	if sgStats.Chunks > 1 {
		s.Reassembled++ // 重组包数
	}
	s.OutOfOrderPackets += sgStats.QueuedPackets
	s.OutOfOrderBytes += sgStats.QueuedBytes
	if length > s.BiggestChunkBytes {
		s.BiggestChunkBytes = length // 最大区块字节数
	}
	if sgStats.Packets > s.BiggestChunkPackets {
		s.BiggestChunkPackets = sgStats.Packets // 最大区块包数
	}
	if sgStats.OverlapBytes != 0 && sgStats.OverlapPackets == 0 {
		// fmt.Printf("bytes:%d, pkts:%d\n", sgStats.OverlapBytes, sgStats.OverlapPackets)
		// panic("Invalid overlap")
	}
	s.OverlapBytes += sgStats.OverlapBytes     // 重叠字节数
	s.OverlapPackets += sgStats.OverlapPackets // 重叠包数

	var ident string
	if dir == reassembly.TCPDirClientToServer {
		ident = fmt.Sprintf("%v %v(%s): ", s.Net, s.Transport, dir)
	} else {
		ident = fmt.Sprintf("%v %v(%s): ", s.Net.Reverse(), s.Transport.Reverse(), dir)
	}
	zap.L().Debug(fmt.Sprintf("%s: SG reassembled packet with %d bytes (start:%v,end:%v,skip:%d,saved:%d,nb:%d,%d,overlap:%d,%d)\n", ident, length, start, end, skip, saved, sgStats.Packets, sgStats.Chunks, sgStats.OverlapBytes, sgStats.OverlapPackets))
	//if skip == -1 && *allowMissingInit {
	//	// this is allowed
	//} else if skip != 0 {
	//	// Missing bytes in stream: do not even try to parse it
	//	return
	//}
	if skip != 0 {
		return
	}
	data := sg.Fetch(length)
	if length > 0 {
		if dir == reassembly.TCPDirClientToServer {
			s.Client.Bytes <- data
		} else {
			s.Server.Bytes <- data
		}
	}
}

func (s *Stream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	zap.L().Debug("Connection Closed", zap.String("ident", s.Ident))
	// 在重组结束时存储
	if config.UseMongo {
		// TODO save mongodb
		s.Lock()
		s.Save()
		s.Unlock()
	}
	close(s.Client.Bytes)
	close(s.Server.Bytes)
	return false
}

func (s *Stream) Save() {
	// TODO ignore empty host
	sessionData := Sessions{
		SessionId:   s.SessionID,
		SrcIp:       s.SrcIP,
		DstIp:       s.DstIP,
		SrcPort:     s.Client.SrcPort,
		DstPort:     s.Client.DstPort,
		Protocol:    "tcp",
		StartTime:   time.DateTime,
		EndTime:     time.DateTime,
		PacketCount: s.PacketCount,
		ByteCount:   s.ByteCount,
		ProtocolFlags: ProtocolFlags{
			TCP: TCPFlags{},
			UDP: UDPFlags{},
		},
		ApplicationProtocol: "",
		Metadata: Metadata{
			HttpInfo: HttpInfo{},
			DnsInfo:  DnsInfo{},
			RtpInfo:  RtpInfo{},
			TlsInfo:  TlsInfo{},
		},
	}
	err := mongo.InsertOne("stream", sessionData)
	if err != nil {
		panic(err)
	}
}
