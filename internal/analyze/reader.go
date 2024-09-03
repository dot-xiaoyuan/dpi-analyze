package analyze

import (
	"bufio"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"go.uber.org/zap"
	"io"
	"sync"
)

// Stream Reader

type StreamReader struct {
	Ident    string
	Parent   *Stream
	IsClient bool
	Bytes    chan []byte
	data     []byte
	Protocol protocols.ProtocolType
	SrcPort  string
	DstPort  string
	Handlers map[protocols.ProtocolType]protocols.ProtocolHandler
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

func (sr *StreamReader) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	b := bufio.NewReader(sr)
	buffer := make([]byte, 0)
	var protocolIdentified bool
	var handler protocols.ProtocolHandler
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
			sr.Protocol = sr.GetIdentifier(buffer)
			if sr.Protocol != "unknown" {
				handler = sr.Handlers[sr.Protocol]
				protocolIdentified = true
				zap.L().Debug("Protocol identified", zap.String("protocols", string(sr.Protocol)))
			}
		}

		if handler != nil {
			handler.HandleData(buffer, sr)
		} else {
			zap.L().Debug("no handler for Protocol", zap.String("protocols", string(sr.Protocol)))
		}
	}
}

func (sr *StreamReader) LockParent() {
	sr.Parent.Lock()
}

func (sr *StreamReader) UnLockParent() {
	sr.Parent.Unlock()
}

func (sr *StreamReader) GetIdentifier(buffer []byte) protocols.ProtocolType {
	return protocols.IdentifyProtocol(buffer, sr.SrcPort, sr.DstPort)
}

func (sr *StreamReader) SetHostName(host string) {
	sr.Parent.Host = host
}

func (sr *StreamReader) GetIdent() bool {
	return sr.IsClient
}

func (sr *StreamReader) SetUrls(urls []string) {
	sr.Parent.Urls = urls
}

func (sr *StreamReader) GetUrls() []string {
	return sr.Parent.Urls
}
