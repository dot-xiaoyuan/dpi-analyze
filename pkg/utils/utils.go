package utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// 工具包

func IdentifyClientHello(data []byte) bool {
	if len(data) < 42 {
		return false
	}
	if data[0] != 0x16 {
		return false
	}
	if data[5] != 0x01 {
		return false
	}
	return true
}

// GetServerExtensionName 获取SNI server name indication
func GetServerExtensionName(data []byte) string {
	// Skip past fixed-length records:
	// 1  Handshake Type
	// 3  Length
	// 2  Version (again)
	// 32 Random
	// next Session ID Length
	pos := 38
	dataLen := len(data)

	/* session id */
	if dataLen < pos+1 {
		return ""
	}
	l := int(data[pos])
	pos += l + 1

	/* Cipher Suites */
	if dataLen < (pos + 2) {
		return ""
	}
	l = int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += l + 2

	/* Compression Methods */
	if dataLen < (pos + 1) {
		return ""
	}
	l = int(data[pos])
	pos += l + 1

	/* Extensions */
	if dataLen < (pos + 2) {
		return ""
	}
	extensionsLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2

	/* Parse extensions to get SNI */
	var extensionItemLen int

	/* Parse each 4 bytes for the extension header */
	for pos <= dataLen && pos < extensionsLen+2 {
		if pos+4 > dataLen {
			return ""
		}

		extensionType := binary.BigEndian.Uint16(data[pos : pos+2])
		extensionItemLen = int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		pos += 4

		if pos+extensionItemLen > dataLen {
			return ""
		}

		if extensionType == 0x00 { // SNI extension
			extensionEnd := pos + extensionItemLen
			for pos+3 <= extensionEnd {
				serverNameLen := int(binary.BigEndian.Uint16(data[pos+3 : pos+5]))
				if pos+3+serverNameLen > extensionEnd {
					return ""
				}

				if data[pos] == 0x00 {
					hostname := make([]byte, serverNameLen)
					copy(hostname, data[pos+5:pos+5+serverNameLen])
					return string(hostname)
				}
				// Move to next SNI item
				pos += 3 + serverNameLen
			}
		} else {
			// Move past other extension types
			pos += extensionItemLen
		}
	}
	return ""
}

func GetServerCipherSuite(data []byte) (cipherSuite string) {
	// length = Length(3) + Version(2) + Random(32) + Session ID (1)
	// Skip past fixed-length records:
	// 1  Handshake Type
	// 3  Length
	// 2  Version (again)
	// 32 Random
	// next Session ID Length
	pos := 38
	dataLen := len(data)

	/* session id */
	if dataLen < pos+1 {
		return
	}
	l := int(data[pos])
	pos += l + 1

	/* Cipher Suites */
	if dataLen < (pos + 2) {
		return
	}
	//zap.L().Info("cipherSuite", zap.ByteString("cs", data[pos+2]))
	cs := data[pos : pos+2]
	// zap.L().Info(fmt.Sprintf("0x%02x%02x", cs[0], cs[1]))
	cipherSuite = fmt.Sprintf("0x%02x%02x", cs[0], cs[1])
	return
}

func ReadByConn(conn net.Conn, bufSize int) (data []byte, err error) {
	buffer := make([]byte, 0)        // 用于存放所有数据
	tempBuf := make([]byte, bufSize) // 临时缓冲区

	for {
		n, err := conn.Read(tempBuf)
		if n > 0 {
			buffer = append(buffer, tempBuf[:n]...) // 将读取到的数据追加到总缓冲区中
		}
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed, all data read")
				break // 读取结束
			}
			return nil, fmt.Errorf("error reading: %v", err)
		}
	}
	return buffer, nil
}
