package analyze

import (
	"bufio"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/logger"
	"go.uber.org/zap"
	"io"
	"sync"
)

type ProtocolHandler interface {
	HandleData(data []byte, reader *StreamReader)
}

type StreamReader struct {
	Ident    string
	Parent   *Stream
	IsClient bool
	bytes    chan []byte
	data     []byte
	Protocol string
	SrcPort  string
	DstPort  string
	Handlers map[string]ProtocolHandler
}

func (sr *StreamReader) Read(p []byte) (n int, err error) {
	ok := true
	for ok && len(sr.data) == 0 {
		sr.data, ok = <-sr.bytes
	}
	if !ok || len(sr.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, sr.data)
	sr.data = sr.data[l:]
	return l, nil
}

func (sr *StreamReader) run(wg *sync.WaitGroup) {
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
			logger.Debug("Error reading stream: %s", zap.Error(err))
			continue
		}
		buffer = append(buffer, data[:n]...)
		if !protocolIdentified {
			sr.Protocol = identifyProtocol(buffer, sr.SrcPort, sr.DstPort)
			if sr.Protocol != "unknown" {
				handler = sr.Handlers[sr.Protocol]
				protocolIdentified = true
				logger.Debug("Protocol identified: %v", zap.String("protocol", sr.Protocol))
			}
		}

		if handler != nil {
			handler.HandleData(buffer, sr)
		} else {
			logger.Debug("no handler for Protocol: %v", zap.String("protocol", sr.Protocol))
		}
	}
}

func identifyProtocol(buffer []byte, srcPort, dstPort string) string {
	if srcPort == "80" || dstPort == "80" {
		return "http"
	}
	return "unknown"
}
