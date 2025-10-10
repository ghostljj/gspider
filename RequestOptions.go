package gspider

type RequestOptions struct {
	ReadByteSize      int               // 读取缓冲区大小（每次从响应体读取的字节数）
	RefererUrl        string            // Referer 请求头（作为来源 URL）
	IsGetJson         int               // 是否偏好接收 JSON：1=设置 Accept 为 JSON；0/−1=保持默认
	IsPostJson        int               // 是否以 JSON 提交：1=Content-Type 为 application/json；0=form；−1=保持默认
	Header            map[string]string // 自定义请求头（键值对，覆盖默认头）
	Cookie            string            // 为当前 URL 设置 Cookie（写入 CookieJar）
	CookieAll         string            // 为当前 URL 及其根域设置 Cookie（两处写入 CookieJar）
	RedirectCount     int               // 最大重定向次数（>0 启用限制，否则用默认策略）
	CacheFullResponse bool              // 是否缓存完整响应体（默认 true；大文件建议关闭以节省内存）

	// —— 时间与超时设置 ——
	Timeout               int64  // 连接建立超时（TCP Dial 超时，单位：秒）。0=禁用。PostBig 中会被提升到至少 5 小时
	ReadWriteTimeout      int64  // 整体请求生命周期超时（单位：秒）。0=禁用。PostBig 中至少 60 秒
	ReadIdleTimeout       int64  // 读取阶段无数据到达的空闲超时（单位：秒）。仅在读取响应体时生效。0=禁用
	WriteIdleTimeout      int64  // 写入阶段（上传）无进展的空闲超时（单位：秒）。仅在请求体上传时生效。0=禁用
	TLSHandshakeTimeout   int64  // TLS 握手时间上限（单位：秒）。默认 10 秒
	ResponseHeaderTimeout int64  // 等待响应头时间上限（单位：秒）。默认 10 秒
	ExpectContinueTimeout int64  // Expect: 100-continue 等待时间（单位：秒）。仅在请求头包含该字段时生效。0=不设置（保持零值：不等待 100-continue，直接发送请求体）
	KeepAliveTimeout      int64  // TCP KeepAlive 探测间隔（单位：秒）。默认 30 秒
	IdleConnTimeout       int64  // 空闲连接保活时长（单位：秒）。0=不设置（保持零值：不主动关闭空闲连接）；默认 90 秒以匹配 http.DefaultTransport 行为
	TcpDelay              int64  // 连接成功后注入的延迟（单位：秒，用于限速/调试）。默认 0
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

		Timeout:               30, // TCP Dial 超时，默认 30 秒
		ReadWriteTimeout:      30, // 整体请求生命周期超时，默认 30 秒
		ReadIdleTimeout:       0,  // 读取空闲超时，默认禁用
		WriteIdleTimeout:      0,  // 写入空闲超时，默认禁用
		TLSHandshakeTimeout:   10, // TLS 握手超时，默认 10 秒
		ResponseHeaderTimeout: 10, // 响应头等待超时，默认 10 秒
		ExpectContinueTimeout: 0,  // Expect: 100-continue 等待时间，默认不设置（零值）
		KeepAliveTimeout:      30, // TCP KeepAlive 间隔，默认 30 秒
		IdleConnTimeout:       90, // 空闲连接保活时长，默认 90 秒
		TcpDelay:              0,  // 注入连接延迟（限速/调试），默认禁用
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

// OptReadIdleTimeout 设置读取阶段无数据到达的空闲超时（单位：秒）。0 表示禁用，仅在读取响应体阶段生效
func OptReadIdleTimeout(readIdleTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ReadIdleTimeout = readIdleTimeout
	})
}

// OptWriteIdleTimeout 设置写入阶段（上传）无进展的空闲超时（单位：秒）。0 表示禁用，仅在请求体上传时生效
func OptWriteIdleTimeout(writeIdleTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.WriteIdleTimeout = writeIdleTimeout
	})
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

// OptReadByteSize 设置读取缓冲区大小（每次从响应体读取的字节数）
func OptReadByteSize(readByteSize int) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ReadByteSize = readByteSize
	})
}

// OptRefererUrl 设置 Referer 请求头（来源地址）
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

// OptHeader 设置自定义请求头（键值对，覆盖默认头）
func OptHeader(header map[string]string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.Header = header
	})
}

// OptRedirectCount 设置最大重定向次数（>0 启用限制，否则使用默认策略）
func OptRedirectCount(redirectCount int) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.RedirectCount = redirectCount
	})
}

// OptCacheFullResponse 是否缓存完整响应体（默认 true；大文件建议关闭以节省内存）
func OptCacheFullResponse(cacheFullResponse bool) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.CacheFullResponse = cacheFullResponse
	})
}

// OptCookie 为当前 URL 设置 Cookie（写入 CookieJar）
func OptCookie(cookie string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.Cookie = cookie
	})
}

// OptCookieAll 为当前 URL 及其根域设置 Cookie（两处写入 CookieJar）
func OptCookieAll(cookieAll string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.CookieAll = cookieAll
	})
}

// OptTimeout 设置连接建立超时（TCP Dial 超时，单位：秒）。0=禁用。在 PostBig 中会被提升到至少 5 小时
func OptTimeout(timeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.Timeout = timeout
	})
}

// OptTcpDelay 设置连接成功后注入的延迟（单位：秒，用于限速/调试）
func OptTcpDelay(tcpDelay int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.TcpDelay = tcpDelay
	})
}

// OptReadWriteTimeout 设置整体请求生命周期超时（单位：秒）。0=禁用。在 PostBig 中至少 60 秒
func OptReadWriteTimeout(readWriteTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ReadWriteTimeout = readWriteTimeout
	})
}

// OptTLSHandshakeTimeout 设置 TLS 握手时间上限（单位：秒）
func OptTLSHandshakeTimeout(tlsHandshakeTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.TLSHandshakeTimeout = tlsHandshakeTimeout
	})
}

// OptResponseHeaderTimeout 设置等待响应头时间上限（单位：秒）
func OptResponseHeaderTimeout(responseHeaderTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ResponseHeaderTimeout = responseHeaderTimeout
	})
}

// OptExpectContinueTimeout 设置 Expect: 100-continue 等待时间（单位：秒）。
// 说明：仅在请求头包含该字段时生效；0 表示不设置（保持零值：不等待 100-continue，直接发送请求体）
func OptExpectContinueTimeout(expectContinueTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.ExpectContinueTimeout = expectContinueTimeout
	})
}

// OptKeepAliveTimeout 设置 TCP KeepAlive 探测间隔（单位：秒）
func OptKeepAliveTimeout(keepAliveTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.KeepAliveTimeout = keepAliveTimeout
	})
}

// OptIdleConnTimeout 设置空闲连接保活时长（单位：秒）。
// 说明：设为 0 表示不设置，保持 Transport 零值（不主动关闭空闲连接）；默认 90 秒以匹配 http.DefaultTransport 行为
func OptIdleConnTimeout(idleConnTimeout int64) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.IdleConnTimeout = idleConnTimeout
	})
}

// OptCancelGroup 设置取消分组标识（便于按分组取消正在进行的请求）
func OptCancelGroup(group string) requestOptionsInterface {
	return newFuncRequests(func(ro *RequestOptions) {
		ro.CancelGroup = group
	})
}
