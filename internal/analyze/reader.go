package analyze

import (
	"bufio"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"io"
	"slices"
	"strings"
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

	buffer := make([]byte, 0, 4096)
	data := make([]byte, 1024)

	var protocolIdentified bool
	var handler protocols.ProtocolHandler
	var headerBytesLimit = 512

	for {
		// read by Reader
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
		// push读取的数据
		buffer = append(buffer, data[:n]...)

		// 只使用 buffer 的前512字节进行协议判断
		if !protocolIdentified && len(buffer) > headerBytesLimit {
			sr.Protocol = sr.GetIdentifier(buffer[:headerBytesLimit])
			if sr.Protocol != "unknown" {
				handler = sr.Handlers[sr.Protocol]
				protocolIdentified = true
			}
		}

		if protocolIdentified && handler != nil {
			processedBytes, needsMoreData := handler.HandleData(buffer, sr)
			if !needsMoreData {
				buffer = buffer[processedBytes:] // 清除已处理的数据
			}
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
		_ = ants.Submit(func() { // 统计SNI
			member.Increment(member.Feature{
				IP:    sr.Parent.SrcIP,
				Field: types.SNI,
				Value: utils.FormatDomain(sni),
			})
		})
		// 如果特征库加载 进行域名分析
		if features.AhoCorasick != nil {
			if ok, feature := features.Match(sni); ok {
				sr.Parent.Metadata.ApplicationInfo.AppName = feature.Name
				sr.Parent.Metadata.ApplicationInfo.AppCategory = feature.Category
			}
			sr.Parent.Metadata.ApplicationInfo.AddUp()
		}
	}
	if version != "" {
		sr.Parent.Metadata.TlsInfo.Version = version
		_ = ants.Submit(func() {
			member.Increment(member.Feature{
				IP:    sr.Parent.SrcIP,
				Field: types.TLSVersion,
				Value: version,
			})
		})
	}
	if cipherSuite != "" {
		sr.Parent.Metadata.TlsInfo.CipherSuite = cipherSuite
		_ = ants.Submit(func() {
			member.Increment(member.Feature{
				IP:    sr.Parent.SrcIP,
				Field: types.CipherSuite,
				Value: cipherSuite,
			})
		})
	}
	sr.Parent.ApplicationProtocol = protocols.TLS
}

// GetIdent 获取流方向
func (sr *StreamReader) GetIdent() bool {
	return sr.IsClient
}

// SetUrls 设置Urls
func (sr *StreamReader) SetUrls(urls string) {
	_, existed := slices.BinarySearch(sr.Parent.Metadata.HttpInfo.Urls, urls)
	if existed {
		return
	}
	sr.Parent.Metadata.HttpInfo.Urls = append(sr.GetUrls(), urls)
}

func (sr *StreamReader) GetUrls() []string {
	return sr.Parent.Metadata.HttpInfo.Urls
}

func (sr *StreamReader) SetHttpInfo(host, userAgent, contentType, upgrade string) {
	httpInfo := types.HttpInfo{
		Host:        host,
		UserAgent:   userAgent,
		ContentType: contentType,
		Upgrade:     upgrade,
		Urls:        sr.GetUrls(),
	}
	// 如果ua有效
	if userAgent != "" {
		_ = ants.Submit(func() { // 统计UA
			member.Store(member.Hash{
				IP:    sr.Parent.SrcIP,
				Field: types.UserAgent,
				Value: userAgent,
			})
		})
	}
	// host
	if host != "" && host != "<no-request-seen>" && !strings.HasPrefix(host, "/") {
		_ = ants.Submit(func() { // 统计 http
			member.Increment(member.Feature{
				IP:    sr.Parent.SrcIP,
				Field: types.HTTP,
				Value: utils.FormatDomain(host),
			})
		})
	}
	// 如果特征库加载 进行域名分析
	if features.AhoCorasick != nil && host != "" {
		if ok, feature := features.Match(host); ok {
			sr.Parent.Metadata.ApplicationInfo.AppName = feature.Name
			sr.Parent.Metadata.ApplicationInfo.AppCategory = feature.Category
		}
		sr.Parent.Metadata.ApplicationInfo.AddUp()
	}
	sr.Parent.Metadata.HttpInfo = httpInfo
	sr.Parent.ApplicationProtocol = protocols.HTTP
}

func (sr *StreamReader) SetApplicationProtocol(applicationProtocol protocols.ProtocolType) {
	sr.Parent.ApplicationProtocol = applicationProtocol
}
