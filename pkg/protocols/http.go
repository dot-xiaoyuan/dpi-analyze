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

type HTTPHandler struct {
}

func (HTTPHandler) HandleData(data []byte, sr StreamReaderInterface) {
	r := bufio.NewReader(bytes.NewReader(data))
	//for {
	if CheckHttpByRequest(data[:50]) {
		req, err := http.ReadRequest(r)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			return
		} else if err != nil {
			zap.L().Debug("Error reading request", zap.Error(err))
			return
		}
		// body, err := ioutil.ReadAll(req.Body)
		//s := len(body)
		//if err != nil {
		//Error("HTTP-request-body", "Got body err: %s\n", err)
		//}
		req.Body.Close()
		// zap.L().Debug("HTTP Request", zap.String("method", req.Method), zap.String("url", req.URL.String()), zap.Int("body", s))
		sr.LockParent()
		sr.SetUrls(req.RequestURI)
		sr.SetHttpInfo(req.Host, req.UserAgent(), req.Header.Get("Content-Type"), req.Header.Get("Upgrade"))
		sr.UnLockParent()
	} else if CheckHttpByResponse(data[:50]) {
		res, err := http.ReadResponse(r, nil)
		if res != nil {
			contentType := res.Header.Get("Content-Type")
			zap.L().Debug("res", zap.Any("res", contentType))
		}
		var req string
		sr.LockParent()
		urls := sr.GetUrls()
		if len(urls) == 0 {
			req = fmt.Sprintf("<no-request-seen>")
		} else {
			req = urls[0]
		}
		sr.SetHttpInfo(req, "", "", "")
		sr.UnLockParent()
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			return
		} else if err != nil {
			zap.L().Debug("Error reading response", zap.Error(err))
			return
		}
		res.Body.Close()
		// zap.L().Debug("HTTP Req", zap.String("req", req))
		// zap.L().Debug("HTTP Response", zap.Int("status code", res.StatusCode))
	}
	//}
}
