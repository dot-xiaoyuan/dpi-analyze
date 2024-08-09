package reassemble

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
	Client              StreamReader
	Server              StreamReader
	TcpState            *reassembly.TCPSimpleFSM
	OptChecker          reassembly.TCPOptionCheck
	Net, Transport      gopacket.Flow
	fsmErr              bool
	Urls                []string
	Ident               string
	Host                string
	RejectFSM           int // FSM (Finite State Machine)有限状态机
	RejectConnFsm       int
	RejectOpt           int
	MissBytes           int
	Sz                  int
	Pkt                 int
	Reassembled         int
	OutOfOrderPackets   int
	OutOfOrderBytes     int
	BiggestChunkBytes   int
	BiggestChunkPackets int
	OverlapBytes        int
	OverlapPackets      int
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
	close(s.Client.Bytes)
	close(s.Server.Bytes)
	return false
}
