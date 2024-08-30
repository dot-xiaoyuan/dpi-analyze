package protocols

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"go.uber.org/zap"
)

type TLSHandler struct{}

func (TLSHandler) HandleData(data []byte, sr *analyze.StreamReader) {
	if len(data) < 5 {
		return
	}

	// check if it's a Client Hello
	if utils.IdentifyClientHello(data) {
		// is ClientHello
		if hostname := utils.GetServerExtensionName(data[5:]); hostname != "" {
			zap.L().Debug("Client Hello", zap.String("hostname", hostname))
			sr.Parent.Lock()
			sr.Parent.Host = hostname
			sr.Parent.Unlock()
		}
	}
}
