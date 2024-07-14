package protocol

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/reassemble"
	"io"
	"io/ioutil"
	"net/http"
)

type HTTPHandler struct{}

func (HTTPHandler) HandleData(data []byte, sr *reassemble.StreamReader) {
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
			fmt.Printf("HTTP/%s Request: %s %s (body:%d)\n", sr.Ident, req.Method, req.URL, s)
			sr.Parent.Lock()
			sr.Parent.Urls = append(sr.Parent.Urls, req.URL.String())
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
			sr.Parent.Unlock()
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil {
				//Error("HTTP-response", "HTTP/%s Response error: %s (%v,%+v)\n", h.ident, err, err, err)
				continue
			}
			res.Body.Close()
			fmt.Println("Response:", req, res.StatusCode)
		}
	}
}
