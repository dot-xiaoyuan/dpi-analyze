package reassemble

import (
	"bufio"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"go.uber.org/zap"
	"io"
	"sync"
)

type StreamReader struct {
	Ident    string
	Parent   *Stream
	IsClient bool
	Bytes    chan []byte
	data     []byte
	Protocol string
	SrcPort  string
	DstPort  string
	Handlers map[string]ProtocolHandler
}

func (sr *StreamReader) Read(p []byte) (n int, err error) {
	ok := true
	for ok && len(sr.data) == 0 {
		sr.data, ok = <-sr.Bytes
	}
	if !ok || len(sr.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, sr.data)
	sr.data = sr.data[l:]
	return l, nil
}

type ProtocolHandler interface {
	HandleData(data []byte, reader *StreamReader)
}

func (sr *StreamReader) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	b := bufio.NewReader(sr)
	buffer := make([]byte, 0)
	var protocolIdentified bool
	var handler ProtocolHandler
	for {
		// read by Reader
		data := make([]byte, 4096)
		n, err := b.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			zap.L().Debug("Error reading stream", zap.Error(err))
			continue
		}
		buffer = append(buffer, data[:n]...)
		if !protocolIdentified {
			sr.Protocol = utils.IdentifyProtocol(buffer, sr.SrcPort, sr.DstPort)
			if sr.Protocol != "unknown" {
				handler = sr.Handlers[sr.Protocol]
				protocolIdentified = true
				zap.L().Debug("Protocol identified", zap.String("protocol", sr.Protocol))
			}
		}

		if handler != nil {
			handler.HandleData(buffer, sr)
		} else {
			zap.L().Debug("no handler for Protocol", zap.String("protocol", sr.Protocol))
		}
	}
}
