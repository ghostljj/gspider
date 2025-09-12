package gspider

import "time"

type RequestOptions struct {
	ReadByteSize      int               // 读取字节大小
	RefererUrl        string            // 来源url
	IsGetJson         int               // 是否接收 Json  -1不发 0否 1是
	IsPostJson        int               // 是否提交 Json  -1不发 0否 1是
	Header            map[string]string // 头参数
	Cookie            string            // cookie     单独url
	CookieAll         string            // cookieAll  根url+单独url
	RedirectCount     int               // 重定向次数
	CacheFullResponse bool              // 是否缓存完整响应字节（默认true，超大文件建议关闭）

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

// OptReadByteSize 设置读写大小
func OptReadByteSize(readByteSize int) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ReadByteSize = readByteSize
	})
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

// OptCacheFullResponse 是否缓存完整响应字节（默认true，超大文件建议关闭）
func OptCacheFullResponse(cacheFullResponse bool) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.CacheFullResponse = cacheFullResponse
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
