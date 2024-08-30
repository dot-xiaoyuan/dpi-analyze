package protocols

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
)

type HTTPData struct {
	Method string `bson:"method"`
	URL    string `bson:"url"`
	Host   string `bson:"host"`
}

type HTTPHandler struct{}

func (HTTPHandler) HandleData(data []byte, sr *analyze.StreamReader) {
	r := bufio.NewReader(bytes.NewReader(data))
	for {
		if sr.IsClient {
			req, err := http.ReadRequest(r)
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil {
				//flow.Debug("HTTP-request", "HTTP/%s Request error: %s (%v,%+v)\n", sr.ident, err, err, err)
				continue
			}
			body, err := ioutil.ReadAll(req.Body)
			s := len(body)
			//if err != nil {
			//Error("HTTP-request-body", "Got body err: %s\n", err)
			//}
			req.Body.Close()
			zap.L().Debug("HTTP Request", zap.String("method", req.Method), zap.String("url", req.URL.String()), zap.Int("body", s))
			//app := &HTTPData{
			//	Method: req.Method,
			//	URL:    req.URL.String(),
			//	Host:   req.Host,
			//}
			sr.Parent.Lock()
			sr.Parent.Urls = append(sr.Parent.Urls, req.URL.String())
			//sr.Parent.Protocols = append(sr.Parent.Protocols, ApplicationData{
			//	Protocol: "http",
			//	Data:     app,
			//})
			sr.Parent.Unlock()
		} else {
			res, err := http.ReadResponse(r, nil)
			var req string
			sr.Parent.Lock()
			if len(sr.Parent.Urls) == 0 {
				req = fmt.Sprintf("<no-request-seen>")
			} else {
				req, sr.Parent.Urls = sr.Parent.Urls[0], sr.Parent.Urls[1:]
			}
			//app := &HTTPData{
			//	Method: "response",
			//	URL:    res.Request.URL.String(),
			//	Host:   res.Request.Host,
			//}
			//sr.Parent.Protocols = append(sr.Parent.Protocols, ApplicationData{
			//	Protocol: "http",
			//	Data:     app,
			//})
			sr.Parent.Unlock()
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil {
				//Error("HTTP-response", "HTTP/%s Response error: %s (%v,%+v)\n", h.ident, err, err, err)
				continue
			}
			res.Body.Close()
			zap.L().Debug("HTTP Req", zap.String("req", req))
			zap.L().Debug("HTTP Response", zap.Int("status code", res.StatusCode))
		}
	}
}
