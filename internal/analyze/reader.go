package analyze

import (
	"bufio"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
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
	sr.Parent.Wg.Add(1)
	defer wg.Done()
	b := bufio.NewReader(sr)
	buffer := make([]byte, 0)
	var protocolIdentified bool
	var handler protocols.ProtocolHandler
	for {
		// read by Reader
		data := make([]byte, 1024)
		n, err := b.Read(data)
		if err != nil {
			if err == io.EOF {
				go func() {
					defer sr.Parent.Wg.Done()
					// zap.L().Debug("Stream EOF", zap.String("Ident", sr.Ident))
					sr.Parent.Save()
				}()
				break
			}
			continue
		}
		buffer = append(buffer, data[:n]...)
		if !protocolIdentified {
			sr.Protocol = sr.GetIdentifier(buffer)
			if sr.Protocol != "unknown" {
				handler = sr.Handlers[sr.Protocol]
				protocolIdentified = true
			}
		}

		if handler != nil {
			handler.HandleData(buffer, sr)
		}
	}
}

func (sr *StreamReader) LockParent() {
	sr.Parent.Lock()
}

func (sr *StreamReader) UnLockParent() {
	sr.Parent.Unlock()
}

// GetIdentifier 获取协议标识
func (sr *StreamReader) GetIdentifier(buffer []byte) protocols.ProtocolType {
	return protocols.IdentifyProtocol(buffer, sr.SrcPort, sr.DstPort)
}

// SetTlsInfo SetHostName
func (sr *StreamReader) SetTlsInfo(sni, version, cipherSuite string) {
	if sni != "" {
		sr.Parent.Metadata.TlsInfo.Sni = sni
	}
	if version != "" {
		sr.Parent.Metadata.TlsInfo.Version = version
	}
	if cipherSuite != "" {
		sr.Parent.Metadata.TlsInfo.CipherSuite = cipherSuite
	}
	sr.Parent.ApplicationProtocol = protocols.TLS
	// 如果特征库加载 进行域名分析
	if features.DomainAc != nil {
		sr.Parent.Metadata.ApplicationInfo.AppName = features.DomainMatch(sni)
	}
}

// GetIdent 获取流方向
func (sr *StreamReader) GetIdent() bool {
	return sr.IsClient
}

// SetUrls 设置Urls
func (sr *StreamReader) SetUrls(urls string) {
	sr.Parent.Metadata.HttpInfo.Urls = append(sr.GetUrls(), urls)
}

func (sr *StreamReader) GetUrls() []string {
	return sr.Parent.Metadata.HttpInfo.Urls
}

func (sr *StreamReader) SetHttpInfo(host, userAgent, contentType, upgrade string) {
	httpInfo := HttpInfo{
		Host:        host,
		UserAgent:   userAgent,
		ContentType: contentType,
		Upgrade:     upgrade,
		Urls:        sr.GetUrls(),
	}
	// 如果特征库加载 进行域名分析
	if features.DomainAc != nil {
		sr.Parent.Metadata.ApplicationInfo.AppName = features.DomainMatch(host)
	}
	sr.Parent.Metadata.HttpInfo = httpInfo
	sr.Parent.ApplicationProtocol = protocols.HTTP
}

func (sr *StreamReader) SetApplicationProtocol(applicationProtocol protocols.ProtocolType) {
	sr.Parent.ApplicationProtocol = applicationProtocol
}
