// Package 这是一个Get Post 强大模拟器
//
// request 是请求对象 ,申请一个新对象，gspider.Session()
//
// response 是返回对象1
//
//	文档安装调试
//	  1、 go get -u golang.org/x/pkgsite/cmd/pkgsite@latest
//	  2、 go install golang.org/x/pkgsite/cmd/pkgsite@latest
//	  3、 pkgsite -http=:6060 -list=false
//	  4、 打开 http://127.0.0.1:6060/github.com/ghostljj/gspider#pkg-overview
package gspider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"sync"
)

//--------------------------------------------------------------------------------------------------------------

// Request 这是一个请求对象
type Request struct {
	LocalIP   string // 本地 网络 IP
	UserAgent string
	cancel    context.CancelFunc
	cancelCtx context.Context
	cancelMu  sync.Mutex

	HttpProxyInfo string // 设置Http代理 例：http://127.0.0.1:1081
	HttpProxyAuto bool   // 自动获取http_proxy变量 默认不开启
	Socks5Address string // Socks5地址 例：127.0.0.1:7813
	Socks5User    string // Socks5 用户名
	Socks5Pass    string // Socks5 密码

	cookieJar       http.CookieJar // CookieJar
	Verify          bool           // https 默认不验证ssl
	tlsClientConfig *tls.Config    // 证书验证配置

	defaultHeaderTemplate map[string]string //发送 请求 头 一些默认值

	wgDone             sync.WaitGroup
	chHttpResponse     chan *http.Response
	chHttpResponseOnce sync.Once // 标记 chHttpResponse 是否已关闭
	ChUploaded         chan *int64
	chUploadedOnce     sync.Once // 标记 ChUploaded 是否已关闭
	ChContentItem      chan []byte
	chContentItemOnce  sync.Once // 标记 ChContentItem 是否已关闭
}

func (req *Request) Cancel() {
	req.cancelMu.Lock()
	defer req.cancelMu.Unlock()
	if req.cancel != nil {
		req.cancel()
	}
}

// defaultRequestOptions 默认配置参数
func defaultRequest() *Request {
	req := Request{
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
		Verify:        false,
		HttpProxyAuto: false,
	}
	req.cancelCtx = context.Background()
	req.CookieJarReset()
	req.defaultHeaderTemplate = make(map[string]string)
	req.defaultHeaderTemplate["accept-encoding"] = "gzip, deflate, br"
	req.defaultHeaderTemplate["accept-language"] = "zh-CN,zh;q=0.9"
	req.defaultHeaderTemplate["connection"] = "keep-alive"
	req.defaultHeaderTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"

	return &req
}

// Session
// 创建Request对象
func Session() *Request {
	return defaultRequest()
}

// 安全关闭 chHttpResponse
func (req *Request) safeCloseHttpResponseChan() {
	if req.chHttpResponse != nil {
		req.chHttpResponseOnce.Do(func() {
			close(req.chHttpResponse)
			req.chHttpResponse = nil
		})
	}
}

// 安全关闭 ChUploaded
func (req *Request) safeCloseUploadedChan() {
	if req.ChUploaded != nil {
		req.chUploadedOnce.Do(func() {
			close(req.ChUploaded)
			req.ChUploaded = nil
		})
	}
}

// 安全关闭 ChContentItem
func (req *Request) safeCloseContentItemChan() {
	if req.ChContentItem != nil {
		req.chContentItemOnce.Do(func() {
			close(req.ChContentItem)
			req.ChContentItem = nil
		})
	}
}

// SetTLSClientFile (server.ca)
// 单向 TLS，只验证 server.ca证书链
func (req *Request) SetTLSClientFile(serverCaFile string) {
	byteServerCa, err := os.ReadFile(serverCaFile)
	if err != nil {
		Log.Fatal("ServerCaFile:", err)
	}
	req.SetTLSClient(byteServerCa)
}

// SetTLSClient (server.ca)
// 单向 TLS，只验证 server.ca证书链
func (req *Request) SetTLSClient(serverCa []byte) {

	req.tlsClientConfig = &tls.Config{RootCAs: LoadCa(serverCa),
		Certificates: []tls.Certificate{}} //无需客户端证书
	req.Verify = true
}

// SetmTLSClientFile ("client.crt", "client.key", "server.ca")
// 双向 mTLS  客户端证书  + 服务器 server.ca证书链
func (req *Request) SetmTLSClientFile(clientCrtFile, clientKeyFile, serverCaFile string) {
	byteClientCrt, err := os.ReadFile(clientCrtFile)
	if err != nil {
		Log.Fatal("ClientCaFile:", err)
	}
	byteClientKey, err := os.ReadFile(clientKeyFile)
	if err != nil {
		Log.Fatal("ClientKeyFile:", err)
	}
	byteServerCa, err := os.ReadFile(serverCaFile)
	if err != nil {
		Log.Fatal("ServerCaFile:", err)
	}
	req.SetmTLSClient(byteClientCrt, byteClientKey, byteServerCa)
}

// SetmTLSClient ("client.crt", "client.key", "server.ca")
// 双向 mTLS  客户端证书  + 服务器 server.ca证书链  使用纯字符串可配置在应用中一起生成
func (req *Request) SetmTLSClient(clientCrt, clientKey, serverCa []byte) {
	pair, e := tls.X509KeyPair(clientCrt, clientKey)
	if e != nil {
		Log.Fatal("LoadX509KeyPair:", e)
	}
	//双向 mTLS  客户端证书  + 服务器 server.ca证书链
	req.tlsClientConfig = &tls.Config{RootCAs: LoadCa(serverCa),
		Certificates: []tls.Certificate{pair}} //还需要客户端证书
	req.Verify = true
}
