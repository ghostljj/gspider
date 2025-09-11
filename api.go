package gspider

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/andybalholm/brotli"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

//--------------------------------------------------------------------------------------------------------------

func (req *Request) GetRequestOptions(strUrl string, opts ...requestOptionsInterface) (ro *RequestOptions) {

	ro = &RequestOptions{
		ReadByteSize:          1024 * 4,
		IsPostJson:            -1,
		IsGetJson:             -1,
		Header:                make(map[string]string),
		RedirectCount:         30,
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

func (req *Request) request(strMethod, strUrl string, strPostData *string, ro *RequestOptions) *Response {
	var bytePostData []byte
	if strPostData != nil {
		bytePostData = []byte(*strPostData)
	}
	return req.requestByte(strMethod, strUrl, bytePostData, ro)
}

func (req *Request) requestByte(strMethod, strUrl string, bytesPostData []byte, ro *RequestOptions) *Response {
	if ro == nil {
		ro = req.GetRequestOptions(strUrl)
	}
	return req.sendByte(strMethod, strUrl, bytesPostData, ro)
}

func (req *Request) XXX(strMethod, strUrl string, bytesPostData []byte, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	return req.requestByte(strMethod, strUrl, bytesPostData, ro)
}

func (req *Request) Get(strUrl string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	return req.request("GET", strUrl, nil, ro)
}

func (req *Request) GetJson(strUrl string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsGetJson = 1
	return req.request("GET", strUrl, nil, ro)
}
func (req *Request) GetJsonR(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsGetJson = 1
	if strPostData != "" {
		ro.IsPostJson = 1
	}
	return req.request("GET", strUrl, &strPostData, ro)
}

func (req *Request) DeleteJson(strUrl string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsGetJson = 1
	return req.request("DELETE", strUrl, nil, ro)
}

func (req *Request) DeleteJsonR(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsGetJson = 1
	if strPostData != "" {
		ro.IsPostJson = 1
	}
	return req.request("DELETE", strUrl, &strPostData, ro)
}

// Post 方法
func (req *Request) Post(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsPostJson = 0
	return req.request("POST", strUrl, &strPostData, ro)
}

// Post 方法
func (req *Request) PostBig(strUrl string, bytesPostData []byte, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	timeOut := time.Duration(60 * 60 * 5)
	if ro.Timeout <= timeOut {
		ro.Timeout = timeOut
	}
	readWriteTimeout := time.Duration(60)
	if ro.ReadWriteTimeout <= readWriteTimeout {
		ro.ReadWriteTimeout = readWriteTimeout
	}
	return req.requestByte("POST", strUrl, bytesPostData, ro)
}

func (req *Request) PostJson(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsPostJson = 1
	ro.IsGetJson = 1
	return req.request("POST", strUrl, &strPostData, ro)
}

// Put Put方法
func (req *Request) Put(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsPostJson = 0
	return req.request("PUT", strUrl, &strPostData, ro)
}
func (req *Request) PutJson(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsGetJson = 1
	ro.IsPostJson = 1
	return req.request("PUT", strUrl, &strPostData, ro)
}

// PATCH PATCH方法
func (req *Request) Patch(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsPostJson = 0
	return req.request("PATCH", strUrl, &strPostData, ro)
}
func (req *Request) PatchJson(strUrl, strPostData string, opts ...requestOptionsInterface) *Response {
	ro := req.GetRequestOptions(strUrl, opts...)
	ro.IsGetJson = 1
	ro.IsPostJson = 1
	return req.request("PATCH", strUrl, &strPostData, ro)
}

// 获取img src 值
func (req *Request) GetBase64ImageSrc(strUrl string, opts ...requestOptionsInterface) (*Response, string) {
	res, strContent := req.GetBase64Image(strUrl, opts...)
	if res.GetErr() == nil {
		contentType := res.GetResHeader().Get("Content-Type")
		strContent = "data:" + contentType + ";base64," + strContent + ""
	}
	return res, strContent
}

// 获取Base64 字符串
func (req *Request) GetBase64Image(strUrl string, opts ...requestOptionsInterface) (*Response, string) {
	ro := req.GetRequestOptions(strUrl, opts...)
	res := req.send("GET", strUrl, nil, ro)
	return res, base64.StdEncoding.EncodeToString(res.GetBytes())
}

func (req *Request) send(strMethod, strUrl string, strPostData *string, rp *RequestOptions) *Response {
	var bytesPostData []byte
	if strPostData != nil {
		bytesPostData = []byte(*strPostData)
	}
	return req.sendByte(strMethod, strUrl, bytesPostData, rp)
}

// SendRedirect 发送请求
// strMethod GET POST PUT ...
func (req *Request) sendByte(strMethod, strUrl string, bytesPostData []byte, rp *RequestOptions) *Response {
	res := newResponse(req)
	strMethod = strings.ToUpper(strMethod)
	reqURI, err := url.Parse(strUrl)
	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}
	res.reqUrl = reqURI.String()

	httpClient := &http.Client{
		Timeout: rp.ReadWriteTimeout * time.Second,
	}

	res.reqPostData = ""
	if len(bytesPostData) <= 12000 {
		res.reqPostData = string(bytesPostData)
	}

	ctx, cancel := context.WithCancel(context.Background())
	progressReader := &UploadedProgressReader{
		Reader:     bytes.NewReader(bytesPostData),
		Total:      int64(len(bytesPostData)),
		chUploaded: req.ChUploaded,
		LastTime:   time.Now(),
		ctx:        ctx,
	}
	defer func() {
		err := progressReader.Close()
		if err != nil {
			res.err = err
		}
	}()

	httpReq, err := http.NewRequestWithContext(ctx, strMethod, strUrl, progressReader)
	req.cancel = cancel

	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}

	ts := &http.Transport{}
	ts.IdleConnTimeout = time.Second * 90                             // 空闲连接的最长保持时间。超过此时间后，连接会被自动关闭。默认90
	ts.TLSHandshakeTimeout = rp.TLSHandshakeTimeout * time.Second     // 限制执行TLS握手所花费的时间
	ts.ResponseHeaderTimeout = rp.ResponseHeaderTimeout * time.Second // 响应头超时时间
	// ts.ExpectContinueTimeout = 1 * time.Second  //限制client在发送包含 Expect: 100-continue 的header到收到继续发送body的response之间的时间等待 POST才可能需要

	// 新增：禁用 HTTP/2，强制使用 HTTP/1.1
	//ts.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	//超时设置  代理设置
	{
		netDialer := &net.Dialer{
			Timeout:   rp.Timeout * time.Second,          // TCP 连接超时
			KeepAlive: rp.KeepAliveTimeout * time.Second, // 连接保活时间
		}

		if len(req.LocalIP) > 0 { //设置本地网络ip

			//var localAddr *net.IPAddr
			var localTCPAddr *net.TCPAddr
			if isIPAddress(req.LocalIP) {
				// 直接解析IP
				ip := net.ParseIP(req.LocalIP)
				if ip == nil {
					res.resBytes = []byte(fmt.Sprintf("无效的IP地址: %s", req.LocalIP))
					res.err = fmt.Errorf("无效的IP地址: %s", req.LocalIP)
					return res
				}
				localTCPAddr = &net.TCPAddr{IP: ip, Port: 0}
			} else {
				// 解析域名
				addr, err := net.ResolveIPAddr("ip4", req.LocalIP) // 指定IPv4
				if err != nil {
					addr, err = net.ResolveIPAddr("ip6", req.LocalIP)
					if err != nil {
						res.resBytes = []byte(fmt.Sprintf("域名解析失败: %v", err))
						res.err = err
						return res
					}
				}
				localTCPAddr = &net.TCPAddr{IP: addr.IP, Port: 0}
			}
			netDialer.LocalAddr = localTCPAddr
		}

		if req.Verify && req.tlsClientConfig != nil {
			ts.TLSClientConfig = req.tlsClientConfig
		} else {
			ts.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			} //跳过证书验证
		}

		var httpProxyInfoOK = ""
		if req.HttpProxyAuto {
			//httpProxy := os.Getenv("http_proxy")
			//httpProxys := os.Getenv("https_proxy")
			//httpProxy = strings.ReplaceAll(httpProxy, "\n", "")
			//httpProxys = strings.ReplaceAll(httpProxys, "\n", "")
			//if len(httpProxy) > 0 {
			//	httpProxyInfoOK = httpProxy
			//	if strings.Index(httpProxyInfoOK, "http") == -1 {
			//		httpProxyInfoOK = "http://" + httpProxyInfoOK
			//	}
			//} else if len(httpProxys) > 0 {
			//	httpProxyInfoOK = httpProxys
			//	if strings.Index(httpProxyInfoOK, "http") == -1 {
			//		httpProxyInfoOK = "https://" + httpProxyInfoOK
			//	}
			//}
			ts.Proxy = http.ProxyFromEnvironment
		}
		if len(req.HttpProxyInfo) > 0 {
			httpProxyInfoOK = req.HttpProxyInfo
		}

		if len(httpProxyInfoOK) > 0 { //http 代理设置
			proxyUrl, err := url.Parse(httpProxyInfoOK)
			if err != nil {
				res.err = err
				res.resBytes = []byte(err.Error())
				return res
			}
			ts.Proxy = http.ProxyURL(proxyUrl)
		}
		ts.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 使用已配置的 netDialer 建立连接
			conn, err := netDialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}

			if rp.TcpDelay > 0 {
				time.Sleep(rp.TcpDelay * time.Millisecond)
			}
			return conn, nil
		}
		if len(req.Socks5Address) > 0 { //SOCKS5 代理设置
			var Socks5Auth *proxy.Auth
			if len(req.Socks5User) > 0 {
				Socks5Auth = &proxy.Auth{User: req.Socks5User, Password: req.Socks5Pass}
			}

			// 创建基于 baseDialContext 的 SOCKS5 代理
			netDialerNew, err := proxy.SOCKS5("tcp", req.Socks5Address,
				Socks5Auth,
				netDialer,
			)

			if err != nil {
				res.resBytes = []byte(err.Error())
				res.err = err
				return res
			}

			// 类型断言，检查是否实现了 ContextDialer 接口
			if ctxDialer, ok := netDialerNew.(proxy.ContextDialer); ok {
				// 使用支持上下文的 DialContext 方法
				ts.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					conn, err := ctxDialer.DialContext(ctx, network, addr)
					if err != nil {
						return nil, err
					}
					if rp.TcpDelay > 0 {
						time.Sleep(rp.TcpDelay * time.Millisecond)
					}
					return conn, nil
				}
			} else {
				// 回退到不支持上下文的 Dial 方法
				ts.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					conn, err := netDialerNew.Dial(network, addr)
					if err != nil {
						return nil, err
					}
					if rp.TcpDelay > 0 {
						time.Sleep(rp.TcpDelay * time.Millisecond)
					}
					return conn, nil
				}
			}
		}
	}
	httpClient.Transport = ts

	//设置重定向次数 默认重定向10次
	if rp.RedirectCount > 0 {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			// 没有重定向不会执行，len(via)==1 就是第一次跳进入。选择是否跳
			if len(via) >= rp.RedirectCount {
				return http.ErrUseLastResponse //返回err就是，不跳
			}
			return nil //返回nil就是跳，
		}
	}

	//合并Header
	{
		sendHeader := make(map[string]string)
		if len(rp.RefererUrl) > 0 {
			sendHeader["referer"] = rp.RefererUrl
		}
		if len(req.UserAgent) > 0 {
			sendHeader["user-agent"] = req.UserAgent
		}

		for k, v := range req.defaultHeaderTemplate {
			sendHeader[strings.ToLower(k)] = v
		}

		sendHeader[strings.ToLower(`accept-encoding`)] = `gzip, deflate, br`
		if rp.IsGetJson == 1 { //接收json
			sendHeader[strings.ToLower(`accept`)] = `application/json, text/plain, */*`
		}
		if rp.IsPostJson == 1 { //发送json
			sendHeader[strings.ToLower(`content-type`)] = `application/json;charset=UTF-8`
		} else if rp.IsPostJson == 0 { //发送from
			sendHeader[strings.ToLower(`content-type`)] = `application/x-www-form-urlencoded; charset=UTF-8`
		}

		for k, v := range rp.Header {
			sendHeader[strings.ToLower(k)] = v
		}
		for k, v := range sendHeader {
			if len(v) <= 0 { //为空删除
				httpReq.Header.Del(k)
			} else {
				httpReq.Header.Set(k, v)
			}
		}
	}

	httpClient.Jar = req.cookieJar
	httpReq.Close = true

	httpRes, err := httpClient.Do(httpReq)
	defer func() {
		if req.chHttpResponse != nil {
			close(req.chHttpResponse) // 关闭信道，通知接收方退出
			req.chHttpResponse = nil
		}
	}()
	if err != nil {
		//httpReq.Response.Close
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}
	defer httpRes.Body.Close()

	if req.chHttpResponse != nil {
		req.chHttpResponse <- httpRes
	}

	//返回 响应 Cookies
	res.resCookies = httpRes.Cookies()
	//设置 响应 头信息
	res.resHeader = httpRes.Header
	//设置 请求 头信息
	res.reqHeader = httpRes.Request.Header
	//设置 响应 后的Url
	res.resUrl = httpRes.Request.URL.String()
	//设置响应状态码
	res.statusCode = httpRes.StatusCode

	var reader io.Reader
	//解压流 gzip deflate
	ContentEncoding := httpRes.Header.Get("Content-Encoding")
	{
		switch ContentEncoding {
		case "br":
			reader = brotli.NewReader(httpRes.Body)
		case "gzip":
			reader, err = gzip.NewReader(httpRes.Body)
			if err != nil {
				res.resBytes = []byte(err.Error())
				res.err = err
				return res
			}
		case "deflate":
			reader = flate.NewReader(httpRes.Body)
		default:
			reader = httpRes.Body
		}
	}
	contentType := httpRes.Header.Get("Content-Type")
	isText := strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "application/json")

	if res.resBytes, err = pedanticReadAll(rp.ReadByteSize, reader, req, ctx, isText); err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}

	return res
}

// pedanticReadAll 读取所有字节
func pedanticReadAll(readByteSize int, r io.Reader, req *Request, ctx context.Context, isText bool) (b []byte, err error) {
	buf := make([]byte, readByteSize)
	var bItem []byte // bItem 仅用于文本模式下累积数据

	defer func() {
		if req.ChContentItem != nil {
			if len(bItem) > 0 {
				select {
				case req.ChContentItem <- bItem:
				case <-ctx.Done(): // 上下文已取消，忽略发送失败
				}
			}
			close(req.ChContentItem)
			req.ChContentItem = nil
		}
	}()
	for {
		// 关键：监听 ctx 取消信号，触发则立即中断下载
		select {
		case <-ctx.Done():
			return b, ctx.Err() // 返回已读取数据 + 取消错误
		default:
		}
		n, err := r.Read(buf)
		if n == 0 && err == nil {
			return nil, fmt.Errorf("Read: n=0 with err=nil")
		}

		b = append(b, buf[:n]...)

		// 根据内容类型选择处理方式
		if isText {
			// 文本内容：按行发送
			bItem = append(bItem, buf[:n]...)
			if req.ChContentItem != nil && bytes.Contains(buf[:n], []byte("\n")) {
				select {
				case req.ChContentItem <- bItem:
				case <-ctx.Done():
				}
				bItem = bItem[:0]
			}
		} else {
			// 二进制内容：按块发送
			if n > 0 && req.ChContentItem != nil {
				select {
				case req.ChContentItem <- buf[:n]:
				case <-ctx.Done():
				}
			}
		}

		// 先处理错误前的残留数据，再处理错误
		if err != nil {
			// 如果是EOF且还有未发送的数据，先发送
			if err == io.EOF && req.ChContentItem != nil && len(bItem) > 0 {
				req.ChContentItem <- bItem
				bItem = bItem[:0]
			}
			// 对于EOF，我们返回已读取的数据和nil错误
			if err == io.EOF {
				return b, nil
			}
			// 其他错误返回
			return b, err
		}
	}
}

// 判断字符串是IP地址还是域名
func isIPAddress(host string) bool {
	// 检查是否包含端口（如 "example.com:80"）
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}

	// 尝试解析为IP地址
	return net.ParseIP(host) != nil
}

// OnResHeader 响应头回调
func (req *Request) OnHttpResponse(f func(httpRes *http.Response, req *Request)) {
	req.chHttpResponse = make(chan *http.Response, 1)
	go func() {
		for {
			v, ok := <-req.chHttpResponse
			if ok {
				f(v, req)
			} else {
				break
			}
		}
	}()
}

// OnUploaded 上传回调
func (req *Request) OnUploaded(f func(uploaded *int64, req *Request)) {
	req.ChUploaded = make(chan *int64, 1)
	go func() {
		for {
			v, ok := <-req.ChUploaded
			if ok {
				f(v, req)
			} else {
				break
			}
		}
	}()
}

// OnContent 实现内容回调
func (req *Request) OnContent(f func(byteItem []byte, req *Request)) {
	req.ChContentItem = make(chan []byte, 1)
	go func() {
		for {
			v, ok := <-req.ChContentItem
			if ok {
				f(v, req)
			} else {
				break
			}
		}
	}()
}
