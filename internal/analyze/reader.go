package analyze

import (
	"bufio"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/resolve"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/match"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/domain"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"io"
	"slices"
	"strings"
	"sync"
	"time"
)

// Stream Reader

type StreamReader struct {
	isSaved   bool
	isUaSaved bool
	Ident     string
	Parent    *Stream
	IsClient  bool
	Bytes     chan []byte
	data      []byte
	Protocol  protocols.ProtocolType
	SrcIP     string
	DstIP     string
	SrcPort   string
	DstPort   string
	Handlers  map[protocols.ProtocolType]protocols.ProtocolHandler
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
				if !sr.isSaved {
					sr.isSaved = true
					// save 2 mongo
					if sr.Protocol == "unknown" {
						sr.Protocol = "tcp"
					}
					sessionData := types.Sessions{
						Ident:               sr.Ident,
						SessionId:           sr.Parent.SessionID,
						SrcIp:               sr.SrcIP,
						DstIp:               sr.DstIP,
						SrcPort:             sr.SrcPort,
						DstPort:             sr.DstPort,
						PacketCount:         sr.Parent.PacketsCount,
						ByteCount:           sr.Parent.BytesCount,
						Protocol:            string(sr.Protocol),
						MissBytes:           sr.Parent.MissBytes,
						OutOfOrderPackets:   sr.Parent.OutOfOrderPackets,
						OutOfOrderBytes:     sr.Parent.OutOfOrderBytes,
						OverlapBytes:        sr.Parent.OverlapBytes,
						OverlapPackets:      sr.Parent.OverlapPackets,
						StartTime:           sr.Parent.StartTime,
						EndTime:             time.Now(),
						ProtocolFlags:       sr.Parent.ProtocolFlags,
						ApplicationProtocol: sr.Parent.ApplicationProtocol,
						Metadata:            sr.Parent.Metadata,
					}
					err = mongo.Mongo.InsertOneStream("stream", sessionData)
					if err != nil {
						panic(err)
					}
					sr.Parent.Wg.Done()
				}
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
			member.Increment(types.Feature{ // SNI
				IP:    sr.Parent.SrcIP,
				Field: types.SNI,
				Value: utils.FormatDomain(sni),
			})
		})
		// 开始品牌匹配
		// step.1 先进行精确匹配
		if ok, domain := match.DomainMatch(sni, sr.Parent.SrcIP); ok {
			resolve.DeviceHandle(types.DeviceRecord{
				IP:           sr.Parent.SrcIP,
				OriginChanel: types.Device,
				OriginValue:  sni,
				Os:           "",
				Version:      "",
				Device:       "",
				Brand:        strings.ToLower(domain.BrandName),
				Model:        "",
				Description:  domain.Description,
				Icon:         domain.Icon,
				LastSeen:     time.Now(),
			})
		}
		// 如果特征库加载 进行域名分析
		if config.UseFeature && domain.AhoCorasick != nil {
			if ok, feature := domain.Match(sni); ok {
				sr.Parent.Metadata.ApplicationInfo.AppName = feature.Name
				sr.Parent.Metadata.ApplicationInfo.AppCategory = feature.Category
			} else {
				sr.Parent.Metadata.ApplicationInfo.AppName = sni
				sr.Parent.Metadata.ApplicationInfo.AppCategory = "unknown"
			}
			sr.Parent.Metadata.ApplicationInfo.AddUp()
		}
	}
	if version != "" {
		sr.Parent.Metadata.TlsInfo.Version = version
		_ = ants.Submit(func() {
			member.Increment(types.Feature{ // TLS version
				IP:    sr.Parent.SrcIP,
				Field: types.TLSVersion,
				Value: version,
			})
		})
	}
	if cipherSuite != "" {
		sr.Parent.Metadata.TlsInfo.CipherSuite = cipherSuite
		_ = ants.Submit(func() {
			member.Increment(types.Feature{ // 加密套件
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
	// 如果UserAgent不为空且开启了ua分析
	if config.UseUA && len(userAgent) > 0 && !sr.isUaSaved {
		sr.isUaSaved = true
		_ = ants.Submit(func() {
			uaStr := resolve.AnalyzeByUserAgent(sr.Parent.SrcIP, userAgent, host)
			if len(uaStr) > 0 {
				member.Store(member.Hash{
					IP:    sr.Parent.SrcIP,
					Field: types.UserAgent,
					Value: uaStr,
				})
			}
		})
	}
	// host
	if host != "" && host != "<no-request-seen>" && !strings.HasPrefix(host, "/") {
		_ = ants.Submit(func() { // 统计 http
			member.Increment(types.Feature{ // HTTP
				IP:    sr.Parent.SrcIP,
				Field: types.HTTP,
				Value: utils.FormatDomain(host),
			})
		})
	}
	// 如果特征库加载 进行域名分析
	if config.UseFeature && domain.AhoCorasick != nil && host != "" && !strings.HasPrefix(host, "/") {
		if ok, feature := domain.Match(host); ok {
			sr.Parent.Metadata.ApplicationInfo.AppName = feature.Name
			sr.Parent.Metadata.ApplicationInfo.AppCategory = feature.Category
		} else {
			sr.Parent.Metadata.ApplicationInfo.AppName = host
			sr.Parent.Metadata.ApplicationInfo.AppCategory = "unknown"
		}
		sr.Parent.Metadata.ApplicationInfo.AddUp()
	}
	sr.Parent.Metadata.HttpInfo = httpInfo
	sr.Parent.ApplicationProtocol = protocols.HTTP
}

func (sr *StreamReader) SetApplicationProtocol(applicationProtocol protocols.ProtocolType) {
	sr.Parent.ApplicationProtocol = applicationProtocol
}
