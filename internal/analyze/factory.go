package analyze

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"sync"
)

// TCP 流重组工厂实现

type Factory struct {
	wg sync.WaitGroup
}

func (f *Factory) New(netFlow, tcpFlow gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	fsmOptions := reassembly.TCPSimpleFSMOptions{
		SupportMissingEstablishment: false, // 允许缺失 SYN、SYN+ACK、ACK
	}

	// 会话数累加
	capture.SessionCount++

	// 根据在线用户进行缓存
	srcIP, dstIP := netFlow.Src().String(), netFlow.Dst().String()
	var userIP, tranIP string
	if users.ExitsUser(srcIP) {
		userIP, tranIP = srcIP, dstIP
	} else if users.ExitsUser(dstIP) {
		userIP, tranIP = dstIP, srcIP
	}
	member.Increment(types.Feature{ // 会话数
		IP:    userIP,
		Field: types.Session,
		Value: tranIP,
	})

	stream := &Stream{
		SessionID:    protocols.GenerateSessionId(netFlow.Src().String(), netFlow.Dst().String(), tcpFlow.Src().String(), tcpFlow.Dst().String(), "tcp"),
		StartTime:    ac.GetCaptureInfo().Timestamp,
		Net:          netFlow,
		Transport:    tcpFlow,
		TcpState:     reassembly.NewTCPSimpleFSM(fsmOptions),
		Ident:        fmt.Sprintf("%s:%s", netFlow, tcpFlow),
		PacketsCount: 1,
		OptChecker:   reassembly.NewTCPOptionCheck(),
		SrcIP:        srcIP,
		DstIP:        dstIP,
		ProtocolFlags: types.ProtocolFlags{
			TCP: types.TCPFlags{
				SYN: tcp.SYN,
				ACK: tcp.ACK,
				FIN: tcp.FIN,
				RST: tcp.RST,
			},
			UDP: types.UDPFlags{
				IsDNS: false,
			},
		},
		Metadata: types.Metadata{
			HttpInfo: types.HttpInfo{},
			TlsInfo:  types.TlsInfo{},
		},
	}

	stream.Client = StreamReader{
		Bytes:    make(chan []byte),
		Ident:    fmt.Sprintf("%s %s", netFlow, tcpFlow),
		Parent:   stream,
		IsClient: true,
		SrcPort:  srcIP,
		DstPort:  dstIP,
		Handlers: map[protocols.ProtocolType]protocols.ProtocolHandler{
			protocols.HTTP: &protocols.HTTPHandler{},
			protocols.TLS:  &protocols.TLSHandler{},
		},
	}

	stream.Server = StreamReader{
		Bytes:    make(chan []byte),
		Ident:    fmt.Sprintf("%s %s", netFlow.Reverse(), tcpFlow.Reverse()),
		Parent:   stream,
		IsClient: false,
		SrcPort:  tcpFlow.Reverse().Src().String(),
		DstPort:  tcpFlow.Reverse().Dst().String(),
		Handlers: map[protocols.ProtocolType]protocols.ProtocolHandler{
			protocols.HTTP: &protocols.HTTPHandler{},
			protocols.TLS:  &protocols.TLSHandler{},
		},
	}

	f.wg.Add(2)
	go stream.Client.Run(&f.wg)
	go stream.Server.Run(&f.wg)
	return stream
}

func (f *Factory) WaitGoRoutines() {
	//time.Sleep(time.Second * 3)
	f.wg.Wait()
}
