package gspider

import (
	enetxhttp "github.com/enetx/http"
	"github.com/enetx/uquic/http3"
	"net/http"
)

// ——————————————————————————————————————————————————————————————————————————————
// HTTP/3 适配器：将 enetx/http 和 net/http 之间进行转换
// ——————————————————————————————————————————————————————————————————————————————

// http3TransportAdapter 适配 http3.RoundTripper 到 net/http.RoundTripper
type http3TransportAdapter struct {
	http3Transport *http3.RoundTripper
}

// RoundTrip 实现 net/http.RoundTripper 接口
func (a *http3TransportAdapter) RoundTrip(req *http.Request) (*http.Response, error) {
	// 将 net/http.Request 转换为 enetx/http.Request
	enetxReq := &enetxhttp.Request{
		Method:           req.Method,
		URL:              req.URL,
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           enetxhttp.Header(req.Header),
		Body:             req.Body,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Close:            req.Close,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          enetxhttp.Header(req.Trailer),
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
		TLS:              req.TLS,
		GetBody:          req.GetBody,
		Pattern:          req.Pattern,
		Cancel:           req.Cancel,
	}
	enetxReq = enetxReq.WithContext(req.Context())

	// 使用 HTTP/3 RoundTripper 发送请求
	enetxResp, err := a.http3Transport.RoundTrip(enetxReq)
	if err != nil {
		return nil, err
	}

	// 将 enetx/http.Response 转换为 net/http.Response
	resp := &http.Response{
		Status:           enetxResp.Status,
		StatusCode:       enetxResp.StatusCode,
		Proto:            enetxResp.Proto,
		ProtoMajor:       enetxResp.ProtoMajor,
		ProtoMinor:       enetxResp.ProtoMinor,
		Header:           http.Header(enetxResp.Header),
		Body:             enetxResp.Body,
		ContentLength:    enetxResp.ContentLength,
		TransferEncoding: enetxResp.TransferEncoding,
		Close:            enetxResp.Close,
		Uncompressed:     enetxResp.Uncompressed,
		Trailer:          http.Header(enetxResp.Trailer),
		Request:          req,
		TLS:              enetxResp.TLS,
	}

	return resp, nil
}

// CloseIdleConnections 关闭空闲连接
func (a *http3TransportAdapter) CloseIdleConnections() {
	if a.http3Transport != nil {
		a.http3Transport.Close()
	}
}
