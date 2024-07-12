package analyze

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"sync"
)

type Factory struct {
	wg sync.WaitGroup
}

func (f *Factory) New(netFlow, tcpFlow gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	fsmOptions := reassembly.TCPSimpleFSMOptions{
		SupportMissingEstablishment: false, // 允许缺失 SYN、SYN+ACK、ACK
	}

	stream := &Stream{
		net:        netFlow,
		transport:  tcpFlow,
		tcpState:   reassembly.NewTCPSimpleFSM(fsmOptions),
		ident:      fmt.Sprintf("%s:%s", netFlow, tcpFlow),
		optChecker: reassembly.NewTCPOptionCheck(),
	}

	stream.client = StreamReader{
		bytes:    make(chan []byte),
		Ident:    fmt.Sprintf("%s %s", netFlow, tcpFlow),
		Parent:   stream,
		IsClient: true,
		SrcPort:  tcpFlow.Src().String(),
		DstPort:  tcpFlow.Dst().String(),
		Handlers: map[string]ProtocolHandler{
			"http": &HTTPHandler{},
		},
	}

	stream.server = StreamReader{
		bytes:    make(chan []byte),
		Ident:    fmt.Sprintf("%s %s", netFlow.Reverse(), tcpFlow.Reverse()),
		Parent:   stream,
		IsClient: false,
		SrcPort:  tcpFlow.Reverse().Src().String(),
		DstPort:  tcpFlow.Reverse().Dst().String(),
		Handlers: map[string]ProtocolHandler{
			"http": &HTTPHandler{},
		},
	}

	f.wg.Add(2)
	go stream.client.run(&f.wg)
	go stream.server.run(&f.wg)
	return stream
}

func (f *Factory) WaitGoRoutines() {
	f.wg.Wait()
}
