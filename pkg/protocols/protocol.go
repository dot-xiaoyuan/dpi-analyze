package protocols

import (
	"sync/atomic"
)

var (
	ApplicationCount int64
	HTTPCount        int64
	TLSCount         int64
)

// StreamReaderInterface 流Reader接口
type StreamReaderInterface interface {
	GetIdentifier([]byte) ProtocolType
	GetIdent() bool
	SetUrls(urls string)
	GetUrls() []string
	LockParent()
	UnLockParent()
	SetHttpInfo(host, userAgent, contentType, upgrade string)
	SetTlsInfo(sni, version, cipherSuite string)
	SetApplicationProtocol(applicationProtocol ProtocolType)
}

type ProtocolHandler interface {
	HandleData(data []byte, reader StreamReaderInterface) (int, bool)
}

// IncrementCount 应用协议增量统计
func IncrementCount(counts *int64) {
	atomic.AddInt64(counts, 1)
}

type Chart struct {
	Name  ProtocolType `json:"type"`
	Value int64        `json:"value"`
}

// GenerateChartData 生成应用层协议分布图
func GenerateChartData() []Chart {
	httpCount := atomic.LoadInt64(&HTTPCount)
	tlsCount := atomic.LoadInt64(&TLSCount)

	charts := []Chart{
		{Name: HTTP, Value: httpCount},
		{Name: TLS, Value: tlsCount},
	}
	return charts
}
