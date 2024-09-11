package protocols

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type HTTPData struct {
	Method string `bson:"method"`
	URL    string `bson:"url"`
	Host   string `bson:"host"`
}

type HTTPHandler struct{}

func (HTTPHandler) HandleData(data []byte, sr StreamReaderInterface) {
	r := bufio.NewReader(bytes.NewReader(data))
	for {
		if sr.GetIdent() {
			req, err := http.ReadRequest(r)
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil {
				zap.L().Debug("Error reading request", zap.Error(err))
				continue
			}
			// body, err := ioutil.ReadAll(req.Body)
			//s := len(body)
			//if err != nil {
			//Error("HTTP-request-body", "Got body err: %s\n", err)
			//}
			req.Body.Close()
			// zap.L().Debug("HTTP Request", zap.String("method", req.Method), zap.String("url", req.URL.String()), zap.Int("body", s))
			sr.LockParent()
			urls := append(sr.GetUrls(), req.URL.String())
			sr.SetUrls(urls)
			sr.SetHttpInfo(req.Host, req.UserAgent())
			sr.UnLockParent()
		} else {
			res, err := http.ReadResponse(r, nil)
			var req string
			sr.LockParent()
			urls := sr.GetUrls()
			if len(urls) == 0 {
				req = fmt.Sprintf("<no-request-seen>")
			} else {
				req = urls[0]
				sr.SetUrls(urls[1:])
			}
			sr.SetHttpInfo(req, "")
			sr.UnLockParent()
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil {
				//Error("HTTP-response", "HTTP/%s Response error: %s (%v,%+v)\n", h.ident, err, err, err)
				continue
			}
			res.Body.Close()
			// zap.L().Debug("HTTP Req", zap.String("req", req))
			// zap.L().Debug("HTTP Response", zap.Int("status code", res.StatusCode))
		}
	}
}
