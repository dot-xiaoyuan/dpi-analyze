package protocols

import (
	"bufio"
	"bytes"
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type HTTPData struct {
	Method string `bson:"method"`
	URL    string `bson:"url"`
	Host   string `bson:"host"`
}

type HTTPHandler struct {
}

func (HTTPHandler) HandleData(data []byte, sr StreamReaderInterface) (int, bool) {
	r := bufio.NewReader(bytes.NewReader(data))

	// 判断是否接收到完整的请求或者响应头部
	if !hasCompleteHeader(data) {
		return 0, true // 数据不完整，等待更多数据
	}

	// 检查是否是 HTTP 请求
	if CheckHttpByRequest(data[:50]) {
		req, err := http.ReadRequest(r)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			return 0, true
		} else if err != nil {
			zap.L().Debug("Error reading request", zap.Error(err))
			return 0, false
		}

		req.Body.Close()
		sr.LockParent()
		sr.SetUrls(req.RequestURI)
		sr.SetHttpInfo(req.Host, req.UserAgent(), req.Header.Get("Content-Type"), req.Header.Get("Upgrade"))
		sr.UnLockParent()

		return len(data), false
	}

	// 检查是否是 HTTP 响应
	if CheckHttpByResponse(data[:50]) {
		res, err := http.ReadResponse(r, nil)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			return 0, true
		} else if err != nil {
			return 0, false
		}

		res.Body.Close()
		sr.LockParent()
		req := "<no-request-seen>"
		if urls := sr.GetUrls(); len(urls) > 0 {
			req = urls[0]
		}
		sr.SetHttpInfo(req, "", "", "")
		sr.UnLockParent()

		return len(data), false
	}
	return 0, false // 不是 HTTP 请求或者响应
}

func hasCompleteHeader(data []byte) bool {
	headerEnd := []byte("\r\n\r\n")
	return bytes.Contains(data, headerEnd)
}
