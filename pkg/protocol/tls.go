package protocol

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/reassemble"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
)

type TLSHandler struct{}

func (TLSHandler) HandleData(data []byte, sr *reassemble.StreamReader) {
	if len(data) < 5 {
		return
	}

	// check if it's a Client Hello
	if utils.IdentifyClientHello(data) {
		// is ClientHello
		if hostname := utils.GetServerExtensionName(data[5:]); hostname != "" {
			fmt.Println("is ClientHello ================ ", hostname)
			sr.Parent.Lock()
			sr.Parent.Host = hostname
			sr.Parent.Unlock()
		}
	}
}
