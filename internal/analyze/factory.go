package analyze

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
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

	stream := &Stream{
		Net:        netFlow,
		Transport:  tcpFlow,
		TcpState:   reassembly.NewTCPSimpleFSM(fsmOptions),
		Ident:      fmt.Sprintf("%s:%s", netFlow, tcpFlow),
		OptChecker: reassembly.NewTCPOptionCheck(),
		Collections: Collections{
			SrcIP: netFlow.Src().String(),
			DstIP: netFlow.Dst().String(),
		},
	}

	stream.Client = StreamReader{
		Bytes:    make(chan []byte),
		Ident:    fmt.Sprintf("%s %s", netFlow, tcpFlow),
		Parent:   stream,
		IsClient: true,
		SrcPort:  tcpFlow.Src().String(),
		DstPort:  tcpFlow.Dst().String(),
		Handlers: map[string]ProtocolHandler{
			"http": &protocols.HTTPHandler{},
			"tls":  &protocols.TLSHandler{},
		},
	}

	stream.Server = StreamReader{
		Bytes:    make(chan []byte),
		Ident:    fmt.Sprintf("%s %s", netFlow.Reverse(), tcpFlow.Reverse()),
		Parent:   stream,
		IsClient: false,
		SrcPort:  tcpFlow.Reverse().Src().String(),
		DstPort:  tcpFlow.Reverse().Dst().String(),
		Handlers: map[string]ProtocolHandler{
			"http": &protocols.HTTPHandler{},
			"tls":  &protocols.TLSHandler{},
		},
	}

	f.wg.Add(2)
	go stream.Client.Run(&f.wg)
	go stream.Server.Run(&f.wg)
	return stream
}

func (f *Factory) WaitGoRoutines() {
	f.wg.Wait()
}
