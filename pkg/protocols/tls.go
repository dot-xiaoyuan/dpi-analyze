package protocols

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"go.uber.org/zap"
)

type TLSHandler struct{}

func (TLSHandler) HandleData(data []byte, reader StreamReaderInterface) {
	if len(data) < 5 {
		return
	}

	// check if it's a Client Hello
	if utils.IdentifyClientHello(data) {
		// is ClientHello
		if hostname := utils.GetServerExtensionName(data[5:]); hostname != "" {
			zap.L().Debug("Client Hello", zap.String("hostname", hostname))
			reader.LockParent()
			reader.SetHostName(hostname)
			reader.UnLockParent()
		}
	}
}
