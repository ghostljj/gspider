package gspider

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

	Timeout               int64  // TCP连接超时时间（单位：秒）
	ReadWriteTimeout      int64  // 整个请求的超时时间（单位：秒）
	TLSHandshakeTimeout   int64  // 限制执行TLS握手所花费的时间（单位：秒）
	ResponseHeaderTimeout int64  // 响应头超时时间（单位：秒）
	KeepAliveTimeout      int64  // 保持连接超时（单位：秒）
	TcpDelay              int64  // TCP 连接成功后延迟时间（单位：秒）
	CancelGroup           string // 取消分组标识，用于更灵活地按分组取消正在进行的请求
}

func (req *Request) GetRequestOptions(strUrl string, opts ...requestOptionsInterface) (ro *RequestOptions) {

	ro = &RequestOptions{
		ReadByteSize:      1024 * 4,
		IsPostJson:        -1,
		IsGetJson:         -1,
		Header:            make(map[string]string),
		RedirectCount:     30,
		CacheFullResponse: true,

		Timeout:               30,
		ReadWriteTimeout:      30,
		TLSHandshakeTimeout:   10,
		ResponseHeaderTimeout: 10,
		KeepAliveTimeout:      30,
		TcpDelay:              0,
	}
	for _, opt := range opts {
		opt.apply(ro) //这里是塞入实体，针对实体赋值
	}
	if ro.Cookie != "" {
		req.SetCookies(strUrl, ro.Cookie)
	}
	if ro.CookieAll != "" {
		req.SetCookiesAll(strUrl, ro.CookieAll)
	}
	return
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

// OptTimeout 设置 TCP连接超时时间（单位：秒）
func OptTimeout(timeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.Timeout = timeout
	})
}

// OptTcpDelay 设置 TCP 连接成功后，延迟多久（单位：秒）
func OptTcpDelay(tcpDelay int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.TcpDelay = tcpDelay
	})
}

// OptReadWriteTimeout 设置 整个请求的超时时间（单位：秒）
func OptReadWriteTimeout(readWriteTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ReadWriteTimeout = readWriteTimeout
	})
}

// OptTLSHandshakeTimeout 设置 限制执行TLS握手所花费的时间（单位：秒）
func OptTLSHandshakeTimeout(tlsHandshakeTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.TLSHandshakeTimeout = tlsHandshakeTimeout
	})
}

// OptResponseHeaderTimeout 设置 响应头超时时间（单位：秒）
func OptResponseHeaderTimeout(responseHeaderTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ResponseHeaderTimeout = responseHeaderTimeout
	})
}

// OptKeepAliveTimeout 设置 保持连接超时（单位：秒）
func OptKeepAliveTimeout(keepAliveTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.KeepAliveTimeout = keepAliveTimeout
	})
}

// OptCancelGroup 设置取消分组标识
func OptCancelGroup(group string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.CancelGroup = group
	})
}
