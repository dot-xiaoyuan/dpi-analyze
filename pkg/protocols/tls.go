package protocols

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type TLSHandler struct{}

func (TLSHandler) HandleData(data []byte, reader StreamReaderInterface) (int, bool) {
	tls := &layers.TLS{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeTLS, tls)
	var decodedLayers []gopacket.LayerType

	err := parser.DecodeLayers(data, &decodedLayers)
	if err != nil {
		return 0, true
	}

	for _, layerType := range decodedLayers {
		if layerType == layers.LayerTypeTLS {
			for _, hs := range tls.Handshake {
				if len(data) < 6 {
					return 0, true
				}

				handShakeType := data[5]
				var sni, version, cipherSuite string
				version = hs.Version.String()

				switch handShakeType {
				case 0x01:
					sni = getSNI(data)
					//break
				case 0x02:
					cipherSuite = getCipher(data)
					//break
				}

				reader.LockParent()
				reader.SetTlsInfo(sni, version, cipherSuite)
				reader.UnLockParent()
			}
		}
	}
	return len(data), false
}

func getSNI(data []byte) string {
	return utils.GetServerExtensionName(data[5:])
}

func getCipher(data []byte) string {
	return utils.GetServerCipherSuite(data[5:])
}
