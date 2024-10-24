package analyze

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"sync"
	"time"
)

// Stream 流
type Stream struct {
	Wg sync.WaitGroup
	sync.Mutex
	SessionID           string    `bson:"session_id"`
	StartTime           time.Time `bson:"start_time"`
	EndTime             time.Time `bson:"end_time"`
	Client              StreamReader
	Server              StreamReader
	TcpState            *reassembly.TCPSimpleFSM
	OptChecker          reassembly.TCPOptionCheck
	Net, Transport      gopacket.Flow
	fsmErr              bool
	Ident               string `bson:"ident"`
	ProtocolFlags       types.ProtocolFlags
	Metadata            types.Metadata
	SrcIP               string                 `bson:"src_ip"`
	DstIP               string                 `bson:"dst_ip"`
	RejectFSM           int                    `bson:"reject_fsm"` // FSM (Finite State Machine)有限状态机
	RejectConnFsm       int                    `bson:"reject_conn_fsm"`
	RejectOpt           int                    `bson:"reject_opt"`
	MissBytes           int                    `bson:"miss_bytes"`
	BytesCount          int                    `bson:"bytes_count"`
	PacketsCount        int                    `bson:"packets_count"`
	Reassembled         int                    `bson:"reassembled"`
	OutOfOrderPackets   int                    `bson:"out_of_order_packets"`
	OutOfOrderBytes     int                    `bson:"out_of_order_bytes"`
	BiggestChunkBytes   int                    `bson:"biggest_chunk_bytes"`
	BiggestChunkPackets int                    `bson:"biggest_chunk_packets"`
	OverlapBytes        int                    `bson:"overlap_bytes"`
	OverlapPackets      int                    `bson:"overlap_packets"`
	ApplicationProtocol protocols.ProtocolType `bson:"application_protocol"`
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
	//dir, start, end, skip := sg.Info()
	dir, _, _, skip := sg.Info()
	length, saved := sg.Lengths()
	// update stats
	sgStats := sg.Stats()
	if skip > 0 {
		s.MissBytes += skip // 丢失字节
	}
	s.BytesCount += length - saved
	s.PacketsCount += sgStats.Packets
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

	/*var ident string
	if dir == reassembly.TCPDirClientToServer {
		ident = fmt.Sprintf("%v %v(%s)", s.Net, s.Transport, dir)
	} else {
		ident = fmt.Sprintf("%v %v(%s)", s.Net.Reverse(), s.Transport.Reverse(), dir)
	}
	zap.L().Debug(i18n.TT("SG reassembled packet with bytes", map[string]interface{}{
		"count": length,
		"ident": ident,
	}), zap.Bool("start", start),
		zap.Bool("end", end),
		zap.Int("skip", skip),
		zap.Int("saved", saved),
	)*/
	if skip == -1 && config.IgnoreMissing {
		// this is allowed
	} else if skip != 0 {
		// Missing bytes in stream: do not even try to parse it
		return
	}
	_ = ants.Submit(func() {
		member.IncrementQPS(s.SrcIP)
	})
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
	//zap.L().Debug(i18n.TT("Connection Closed", map[string]interface{}{
	//	"ident": s.Ident,
	//}))

	close(s.Client.Bytes)
	close(s.Server.Bytes)
	s.Wg.Wait()

	return false
}

func (s *Stream) Save() {
	if !config.UseMongo {
		return
	}
	// TODO ignore empty host
	sessionData := types.Sessions{
		SessionId:           s.SessionID,
		SrcIp:               s.SrcIP,
		DstIp:               s.DstIP,
		SrcPort:             s.Client.SrcPort,
		DstPort:             s.Client.DstPort,
		PacketCount:         s.PacketsCount,
		ByteCount:           s.BytesCount,
		Protocol:            "tcp",
		MissBytes:           s.MissBytes,
		OutOfOrderPackets:   s.OutOfOrderPackets,
		OutOfOrderBytes:     s.OutOfOrderBytes,
		OverlapBytes:        s.OverlapBytes,
		OverlapPackets:      s.OverlapPackets,
		StartTime:           s.StartTime,
		EndTime:             time.Now(),
		ProtocolFlags:       s.ProtocolFlags,
		ApplicationProtocol: s.ApplicationProtocol,
		Metadata:            s.Metadata,
	}
	err := mongo.Mongo.InsertOne("stream", sessionData)
	if err != nil {
		panic(err)
	}
}
