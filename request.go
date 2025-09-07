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
	"crypto/tls"
	"net/http"
	"os"
	"time"
)

//--------------------------------------------------------------------------------------------------------------

type SendHeader map[string]string
type SendFromData map[string]string
type SendJsonData string
type SendData string
type SendCookie map[string]string
type SendCookieAll map[string]string

//编译不成功。
//func Test(args ...interface {
//	SendHeader | SendFromData | SendJsonData | SendData | SendCookie | SendCookieAll
//}) {
//	for _, arg := range args {
//		switch a := any(arg).(type) {
//		case SendCookie:
//			for k, v := range a {
//				fmt.Println(k, v)
//			}
//			// arg is "GET" params
//			// ?title=website&id=1860&from=login
//		case SendJsonData:
//			fmt.Println(a, "字符串")
//		default:
//			Log.Printf("未处理 %s 参数\n", a)
//		}
//	}
//}

// Request 这是一个请求对象
type Request struct {
	LocalIP   string // 本地 网络 IP
	UserAgent string

	HttpProxyInfo string // 设置Http代理 例：http://127.0.0.1:1081
	HttpProxyAuto bool   // 自动获取http_proxy变量 默认不开启
	Socks5Address string // Socks5地址 例：127.0.0.1:7813
	Socks5User    string // Socks5 用户名
	Socks5Pass    string // Socks5 密码

	cookieJar       http.CookieJar // CookieJar
	Verify          bool           // https 默认不验证ssl
	tlsClientConfig *tls.Config    // 证书验证配置

	defaultHeaderTemplate map[string]string //发送 请求 头 一些默认值
	ChUploaded            chan *int64
	ChContentItem         chan []byte
}

// defaultRequestOptions 默认配置参数
func defaultRequest() *Request {
	req := Request{
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
		Verify:        false,
		HttpProxyAuto: false,
	}

	req.CookieJarReset()
	req.defaultHeaderTemplate = make(map[string]string)
	req.defaultHeaderTemplate["accept-encoding"] = "gzip, deflate, br"
	req.defaultHeaderTemplate["accept-language"] = "zh-CN,zh;q=0.9"
	req.defaultHeaderTemplate["connection"] = "keep-alive"
	req.defaultHeaderTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"

	return &req
}

// Session
//
// 创建Request对象
func Session() *Request {
	return defaultRequest()
	//dros := defaultRequestOptions()
	//for _, opt := range opts {
	//	opt.apply(&dros) //这里是塞入实体，针对实体赋值
	//}
	//return &dros
}

type RequestOptions struct {
	RefererUrl    string            // 来源url
	IsGetJson     int               // 是否接收 Json  -1不发 0否 1是
	IsPostJson    int               // 是否提交 Json  -1不发 0否 1是
	Header        map[string]string // 头参数
	Cookie        string            // cookie     单独url
	CookieAll     string            // cookieAll  根url+单独url
	RedirectCount int               // 重定向次数

	Timeout               time.Duration // 秒 TCP连接超时时间
	ReadWriteTimeout      time.Duration // 秒 整个请求的超时时间
	TLSHandshakeTimeout   time.Duration // 秒 限制执行TLS握手所花费的时间
	ResponseHeaderTimeout time.Duration // 秒 响应头超时时间
	KeepAliveTimeout      time.Duration // 秒 保持连接超时

	TcpDelay time.Duration // 毫秒 TCP 连接成功后，延迟多久
}

//--------------------------------------------------------------------------------------------------------------
//1、为此结构体，定义一个apply接口requestInterface
//2、funcRequests struct 实现apply接口
//3、创建一个函数newFuncRequests 参数为某结构体，返回funcRequests
//4、为此结构体创建参数修改函数返回requestInterface

// requestOptionsInterface 请求参数 采集基本接口
type requestOptionsInterface interface {
	apply(*RequestOptions)
}

// funcRequestOption 定义面的接口使用
type funcRequestOptions struct {
	anyfun func(*RequestOptions)
}

// apply 实现上面的接口，使用这个匿名函数，针对传入的对象，进行操作
func (fro *funcRequestOptions) apply(req *RequestOptions) {
	fro.anyfun(req)
}

// newFuncRequestOption 新建一个匿名函数实体。
// 返回接口地址
func newFuncRequests(anonfun func(ro *RequestOptions)) requestOptionsInterface {
	return &funcRequestOptions{
		anyfun: anonfun,
	}
}

// OptRefererUrl 设置来源地址，返回接口指针(新建一个函数，不执行的，返回他的地址而已)
func OptRefererUrl(refererUrl string) requestOptionsInterface {
	//return &funcRequests{
	//	anyfun: func(ro *requests) {
	//		ro.RefererUrl = refererUrl
	//	},
	//}
	//下面更简洁而已，上门原理一致
	return newFuncRequests(func(ro *RequestOptions) {
		ro.RefererUrl = refererUrl
	})
}

// OptHeader 设置发送头
func OptHeader(header map[string]string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.Header = header
	})
}

// OptRedirectCount 重定向次数
func OptRedirectCount(redirectCount int) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.RedirectCount = redirectCount
	})
}

// OptCookie 设置当前Url cookie
func OptCookie(cookie string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.Cookie = cookie
	})
}

// OptCookieAll 设置当前Url+根Url cookie
func OptCookieAll(cookieAll string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.CookieAll = cookieAll
	})
}

// OptTimeout 设置 秒 TCP连接超时时间
func OptTimeout(timeout time.Duration) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.Timeout = timeout
	})
}

// OptTcpDelay 毫秒  TCP 连接成功后，延迟多久
func OptTcpDelay(tcpDelay time.Duration) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.TcpDelay = tcpDelay
	})
}

// OptReadWriteTimeout 设置 秒 整个请求的超时时间
func OptReadWriteTimeout(readWriteTimeout time.Duration) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ReadWriteTimeout = readWriteTimeout
	})
}

// OptTLSHandshakeTimeout 设置 秒 限制执行TLS握手所花费的时间
func OptTLSHandshakeTimeout(tlsHandshakeTimeout time.Duration) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.TLSHandshakeTimeout = tlsHandshakeTimeout
	})
}

// OptResponseHeaderTimeout 设置 秒 响应头超时时间
func OptResponseHeaderTimeout(responseHeaderTimeout time.Duration) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ResponseHeaderTimeout = responseHeaderTimeout
	})
}

// OptKeepAliveTimeout 设置 秒 保持连接，超时
func OptKeepAliveTimeout(keepAliveTimeout time.Duration) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.KeepAliveTimeout = keepAliveTimeout
	})
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
