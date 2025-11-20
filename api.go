package gspider

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/andybalholm/brotli"
	"golang.org/x/net/proxy"
)

//--------------------------------------------------------------------------------------------------------------

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
	// 使用秒为单位的阈值（int64）
	timeOut := int64((5 * time.Hour) / time.Second) // 5小时 => 18000秒
	if ro.Timeout <= timeOut {
		ro.Timeout = timeOut
	}
	readWriteTimeout := int64(60) // 60秒
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

	// 1. 第一个 defer：关闭信道（先入栈）
	defer func() {
		req.safeCloseHttpResponseChan()
		req.safeCloseUploadedChan()
		req.safeCloseContentItemChan()
		req.wgDone.Wait()
		// 清理可取消上下文，避免后续复用导致立即取消
		req.cancelMu.Lock()
		req.cancel = nil
		req.cancelCause = nil
		// 将请求期上下文复位为父级上下文，确保后续派生以父级为根
		req.cancelCtx = req.baseCtx
		// 分组计数-自动清理：本次请求结束，减少对应分组计数，并在为0时清理分组资源
		if rp != nil && len(rp.CancelGroup) > 0 && req.groupCounts != nil {
			g := rp.CancelGroup
			if cnt, ok := req.groupCounts[g]; ok {
				if cnt <= 1 {
					delete(req.groupCounts, g)
					if req.groupCancelCauses != nil {
						delete(req.groupCancelCauses, g)
					}
					if req.groupCtxs != nil {
						delete(req.groupCtxs, g)
					}
				} else {
					req.groupCounts[g] = cnt - 1
				}
			}
		}
		req.cancelMu.Unlock()
	}()

	res := newResponse(req)
	strMethod = strings.ToUpper(strMethod)
	reqURI, err := url.Parse(strUrl)
	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}
	res.reqUrl = reqURI.String()

	// 允许整体超时为可选：当 ReadWriteTimeout<=0 时，不设置客户端超时（无限）
	//var httpClient *http.Client
	httpClient := req.getSurfHttpClient(rp, res)
	if httpClient == nil {
		return res
	}
	if rp.ReadWriteTimeout > 0 {
		httpClient.Timeout = time.Duration(rp.ReadWriteTimeout) * time.Second
	}

	res.reqPostData = ""
	if len(bytesPostData) <= 12000 {
		res.reqPostData = string(bytesPostData)
	}
	// 统一：始终派生可取消的 ctx
	var ctx context.Context
	var reqCtx context.Context
	{
		req.cancelMu.Lock()
		// 选择父级上下文：优先使用分组上下文，否则使用 baseCtx
		parent := req.baseCtx
		if parent == nil {
			parent = context.Background()
		}
		if rp.CancelGroup != "" {
			// 初始化分组容器
			if req.groupCtxs == nil {
				req.groupCtxs = make(map[string]context.Context)
			}
			if req.groupCancelCauses == nil {
				req.groupCancelCauses = make(map[string]context.CancelCauseFunc)
			}
			if req.groupCounts == nil {
				req.groupCounts = make(map[string]int)
			}
			if gctx, ok := req.groupCtxs[rp.CancelGroup]; ok && gctx != nil {
				parent = gctx
				// 已存在分组：计数+1
				req.groupCounts[rp.CancelGroup] = req.groupCounts[rp.CancelGroup] + 1
			} else {
				gctx, gcancel := context.WithCancelCause(parent)
				req.groupCtxs[rp.CancelGroup] = gctx
				req.groupCancelCauses[rp.CancelGroup] = gcancel
				// 新建分组：计数=1
				req.groupCounts[rp.CancelGroup] = 1
				parent = gctx
			}
		}
		// 为本次请求派生子上下文，便于仅取消该请求
		ctx, cancelCause := context.WithCancelCause(parent)
		req.cancelCause = cancelCause
		// 使用局部请求期上下文，避免并发复用时字段被覆盖导致语义漂移
		reqCtx = ctx
		req.cancelMu.Unlock()
	}

	progressReader := NewUploadedProgressReader(
		bytesPostData,
		reqCtx,
		req.ChUploaded,
	)
	// 写空闲超时监控：当上传在设定时长内没有任何进展，则取消请求
	var stopWriteMon chan struct{}
	if rp.WriteIdleTimeout > 0 && len(bytesPostData) > 0 {
		stopWriteMon = make(chan struct{})
		// 与读取侧保持一致：当上下文为 nil 时避免对 Done() 的访问
		var writeDoneCh <-chan struct{}
		if reqCtx != nil {
			writeDoneCh = reqCtx.Done()
		}
		go func(done <-chan struct{}) {
			idle := time.Duration(rp.WriteIdleTimeout) * time.Second
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			lastUploaded := atomic.LoadInt64(&progressReader.Uploaded)
			lastChange := time.Now()
			for {
				select {
				case <-stopWriteMon:
					return
				case <-done:
					return
				case <-ticker.C:
					cur := atomic.LoadInt64(&progressReader.Uploaded)
					if cur > lastUploaded {
						lastUploaded = cur
						lastChange = time.Now()
						continue
					}
					if time.Since(lastChange) >= idle {
						if req.cancelCause != nil {
							req.cancelCause(fmt.Errorf("write idle timeout: %ds without progress", rp.WriteIdleTimeout))
						}
						return
					}
				}
			}
		}(writeDoneCh)
	}
	defer func() {
		err := progressReader.Close()
		if err != nil {
			res.err = err
		}
		if stopWriteMon != nil {
			close(stopWriteMon)
		}
	}()

	var httpReq *http.Request
	{
		httpReq, err = http.NewRequestWithContext(reqCtx, strMethod, strUrl, progressReader)
	}
	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}
	// 为了让客户端在重定向或重试时能够重新发送请求体，
	// 你必须提供 GetBody 函数。
	if len(bytesPostData) > 0 {
		httpReq.GetBody = func() (io.ReadCloser, error) {
			newReader := NewUploadedProgressReader(
				bytesPostData,
				reqCtx,
				req.ChUploaded,
			)
			return newReader, nil
		}
	}

	// 说明：此处使用零值 Transport（&http.Transport{}）。未显式设置的字段遵循"零值语义"，
	// 与 http.DefaultTransport 的常用默认不同：例如 IdleConnTimeout=0 表示不主动回收空闲连接。
	// 若启用了 Surf，则避免覆盖其 Transport 以保留指纹配置
	// 否则，按现有逻辑构建 Transport

	// 非 Surf 模式：不再支持 HTTP/3

	ts := &http.Transport{}
	if rp.IdleConnTimeout > 0 {
		ts.IdleConnTimeout = time.Duration(rp.IdleConnTimeout) * time.Second // 空闲连接的最长保持时间（仅当 >0 时设置；=0 保持零值，不主动回收）
	}
	ts.TLSHandshakeTimeout = time.Duration(rp.TLSHandshakeTimeout) * time.Second     // TLS 握手超时
	ts.ResponseHeaderTimeout = time.Duration(rp.ResponseHeaderTimeout) * time.Second // 响应头等待超时
	// ExpectContinueTimeout：仅在请求包含 Expect: 100-continue 时生效；
	// 设为 0 表示不设置（保持零值：不等待 100-continue，直接发送请求体）。
	if rp.ExpectContinueTimeout > 0 {
		ts.ExpectContinueTimeout = time.Duration(rp.ExpectContinueTimeout) * time.Second
	}

	// 新增：禁用 HTTP/2，强制使用 HTTP/1.1
	//ts.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	//超时设置  代理设置
	{
		// Dialer 时间设置：Timeout=TCP 连接超时；KeepAlive=TCP 探测间隔
		netDialer := &net.Dialer{
			Timeout:   time.Duration(rp.Timeout) * time.Second,          // TCP 连接超时
			KeepAlive: time.Duration(rp.KeepAliveTimeout) * time.Second, // TCP KeepAlive 间隔
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
			//https://user:pass@host:port
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
				time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
			}
			return conn, nil
		}
		if len(req.Socks5Address) > 0 { //SOCKS5 代理设置
			//socks5://user:pass@host:port
			var Socks5Auth *proxy.Auth
			if len(req.Socks5User) > 0 {
				Socks5Auth = &proxy.Auth{User: req.Socks5User, Password: req.Socks5Pass}
			}

			// 创建基于 baseDialContext 的 SOCKS5 代理
			var netDialerNew proxy.Dialer
			netDialerNew, err = proxy.SOCKS5("tcp", req.Socks5Address,
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
						time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
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
						time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
					}
					return conn, nil
				}
			}
		}
	}
	// 非 Surf 模式（Request.surfBrowserProfile 为 Disabled）下使用自定义 *http.Transport
	if req.surfBrowserProfile == SurfBrowserDisabled {
		httpClient.Transport = ts
	}

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
		clean := func(k string) string {
			k = strings.TrimSpace(k)
			k = strings.TrimRight(k, ":")
			k = strings.ReplaceAll(k, "\r", "")
			k = strings.ReplaceAll(k, "\n", "")
			k = strings.ReplaceAll(k, " ", "")
			if len(k) == 0 {
				return ""
			}
			for i := 0; i < len(k); i++ {
				c := k[i]
				if !(c == '-' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
					return ""
				}
			}
			return k
		}
		sendHeader := make(map[string]string)
		if len(rp.RefererUrl) > 0 {
			sendHeader["referer"] = rp.RefererUrl
		}
		// UA 统一策略：启用 Surf 指纹时，不再用 req.UserAgent 覆盖 Surf 的 UA
		// 仅在未启用 Surf（=Disabled）时才使用 req.UserAgent
		if req.surfBrowserProfile == SurfBrowserDisabled {
			if len(req.UserAgent) > 0 {
				sendHeader["user-agent"] = req.UserAgent
			}
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
			sk := clean(k)
			if sk == "" {
				if req.debug {
					Log.Printf("debug: drop invalid header name=%q", k)
				}
				continue
			}
			sendHeader[strings.ToLower(sk)] = v
		}
		if ae, ok := sendHeader["accept-encoding"]; ok {
			if strings.Contains(strings.ToLower(ae), "zstd") {
				parts := strings.Split(ae, ",")
				filtered := make([]string, 0, len(parts))
				for _, p := range parts {
					s := strings.TrimSpace(p)
					if strings.EqualFold(s, "zstd") {
						continue
					}
					filtered = append(filtered, s)
				}
				sendHeader["accept-encoding"] = strings.Join(filtered, ", ")
			}
		}
		for k, v := range sendHeader {
			if len(v) <= 0 { //为空删除
				httpReq.Header.Del(k)
			} else {
				httpReq.Header.Set(k, v)
			}
		}
		for k, vs := range rp.HeaderAdds {
			sk := clean(k)
			if sk == "" {
				if req.debug {
					Log.Printf("debug: drop invalid header-add name=%q", k)
				}
				continue
			}
			lk := strings.ToLower(sk)
			httpReq.Header.Del(lk)
			for _, v := range vs {
				httpReq.Header.Add(lk, v)
			}
		}

		// 最终清理：删除任何包含非法字符或魔术键的头名，避免标准 net/http 拒绝
		for k := range httpReq.Header {
			kk := strings.TrimSpace(k)
			if kk == "Header-Order:" || kk == "PHeader-Order:" || strings.Contains(kk, ":") || strings.ContainsAny(kk, "\r\n\t ") {
				if req.debug {
					Log.Printf("debug: drop magic/invalid header name=%q", k)
				}
				httpReq.Header.Del(k)
			}
		}
	}

	httpClient.Jar = req.cookieJar
	// 连接关闭策略：仅在非 Surf 模式下使用请求级 Connection: close；
	// Surf 模式下改由传输层禁用 keep-alive 控制连接复用
	if req.surfBrowserProfile == SurfBrowserDisabled || req.surfClose {
		httpReq.Close = true
	}

	// 如果存在请求体，也应该设置 httpReq.ContentLength
	if len(bytesPostData) > 0 {
		httpReq.ContentLength = int64(len(bytesPostData))
	}

    httpRes, err := httpClient.Do(httpReq)
    if err != nil && req.surfBrowserProfile != SurfBrowserDisabled && (len(req.HttpProxyInfo) > 0 || req.HttpProxyAuto || len(req.Socks5Address) > 0) {
        // 回退到非 Surf 的标准 Transport 以提升代理兼容性
        tsFallback := &http.Transport{}
        tsFallback.TLSHandshakeTimeout = time.Duration(rp.TLSHandshakeTimeout) * time.Second
        tsFallback.ResponseHeaderTimeout = time.Duration(rp.ResponseHeaderTimeout) * time.Second
        if rp.ExpectContinueTimeout > 0 {
            tsFallback.ExpectContinueTimeout = time.Duration(rp.ExpectContinueTimeout) * time.Second
        }
        if req.Verify && req.tlsClientConfig != nil {
            tsFallback.TLSClientConfig = req.tlsClientConfig
        } else {
            tsFallback.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        }
        // 代理设置（HTTP 优先，其次环境，其次 SOCKS5）
        if len(req.HttpProxyInfo) > 0 {
            if purl, e := url.Parse(strings.TrimSpace(req.HttpProxyInfo)); e == nil {
                tsFallback.Proxy = http.ProxyURL(purl)
            }
        } else if req.HttpProxyAuto {
            tsFallback.Proxy = http.ProxyFromEnvironment
        }
        // SOCKS5 代理（覆盖 DialContext）
        if len(req.Socks5Address) > 0 {
            var auth *proxy.Auth
            if len(req.Socks5User) > 0 {
                auth = &proxy.Auth{User: req.Socks5User, Password: req.Socks5Pass}
            }
            base := &net.Dialer{Timeout: time.Duration(rp.Timeout) * time.Second, KeepAlive: time.Duration(rp.KeepAliveTimeout) * time.Second}
            if d, e := proxy.SOCKS5("tcp", req.Socks5Address, auth, base); e == nil {
                if cd, ok := d.(proxy.ContextDialer); ok {
                    tsFallback.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
                        c, e2 := cd.DialContext(ctx, network, addr)
                        if e2 != nil { return nil, e2 }
                        if rp.TcpDelay > 0 { time.Sleep(time.Duration(rp.TcpDelay) * time.Second) }
                        return c, nil
                    }
                } else {
                    tsFallback.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
                        c, e2 := d.Dial(network, addr)
                        if e2 != nil { return nil, e2 }
                        if rp.TcpDelay > 0 { time.Sleep(time.Duration(rp.TcpDelay) * time.Second) }
                        return c, nil
                    }
                }
            }
        }
        // 使用回退客户端重试一次
        fbClient := &http.Client{Transport: tsFallback, Timeout: httpClient.Timeout}
        reqClone := httpReq.Clone(httpReq.Context())
        for k := range reqClone.Header {
            kk := strings.TrimSpace(k)
            if kk == "Header-Order:" || kk == "PHeader-Order:" || strings.Contains(kk, ":") || strings.ContainsAny(kk, "\r\n\t ") {
                if req.debug {
                    Log.Printf("debug: fallback drop magic/invalid header name=%q", k)
                }
                reqClone.Header.Del(k)
            }
        }
        httpRes, err = fbClient.Do(reqClone)
    }

	if httpRes != nil {
		defer func() {
			io.Copy(io.Discard, httpRes.Body)
			httpRes.Body.Close()
		}()
	}

    if err != nil {
        if req.debug {
            Log.Printf("debug: request error url=%s err=%v", res.reqUrl, err)
        }
        res.resBytes = []byte(err.Error())
        res.err = err
        return res
    }

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
			var zr io.ReadCloser
			zr, err = zlib.NewReader(httpRes.Body)
			if err == nil {
				reader = zr
			} else {
				reader = flate.NewReader(httpRes.Body)
			}
		default:
			reader = httpRes.Body
		}
	}
	contentType := httpRes.Header.Get("Content-Type")
	isText := strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "application/json")

	if res.resBytes, err = pedanticReadAll(rp, reader, req, ctx, isText); err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}
	if !rp.CacheFullResponse {
		res.resBytes = []byte("response not cached (large file mode)")
	}
	return res
}

// pedanticReadAll 读取所有字节
func pedanticReadAll(rp *RequestOptions, r io.Reader, req *Request, ctx context.Context, isText bool) (b []byte, err error) {
	buf := make([]byte, rp.ReadByteSize)
	var bItem []byte // bItem 仅用于文本模式下累积数据

	if rp.CacheFullResponse {
		b = make([]byte, 0)
	}
	// 读空闲超时：当长时间无任何数据到达时主动取消
	var idleTimer *time.Timer
	var idleDuration time.Duration
	if rp.ReadIdleTimeout > 0 {
		idleDuration = time.Duration(rp.ReadIdleTimeout) * time.Second
		idleTimer = time.NewTimer(idleDuration)
		// 安全处理：当 ctx 为 nil 时，避免对 nil 上下文调用 Done()
		var doneCh <-chan struct{}
		if ctx != nil {
			doneCh = ctx.Done()
		}
		go func(done <-chan struct{}) {
			for {
				select {
				case <-idleTimer.C:
					// 触发空闲超时，取消本次请求上下文
					if req.cancelCause != nil {
						req.cancelCause(fmt.Errorf("read idle timeout: no data for %s", idleDuration))
					}
					return
				case <-done:
					// 当没有上下文（done=nil）时，此分支永远不会触发
					return
				}
			}
		}(doneCh)
	}
	// 提取公共发送函数，减少重复代码
	sendData := func(data []byte) error {
		if req.ChContentItem == nil {
			return nil
		}
		if ctx != nil {
			select {
			case req.ChContentItem <- data:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			// 无上下文时，阻塞等待发送完成
			req.ChContentItem <- data
			return nil
		}
	}
	defer func() {
		// 停止空闲计时器，避免泄露
		if idleTimer != nil {
			if !idleTimer.Stop() {
				// 清空可能残留的触发信号
				select {
				case <-idleTimer.C:
				default:
				}
			}
		}
		// 确保所有数据都被发送
		if req.ChContentItem != nil && len(bItem) > 0 {
			dataCopy := make([]byte, len(bItem))
			copy(dataCopy, bItem)
			if err := sendData(dataCopy); err != nil {
				// Log warning if sending remaining data fails
				fmt.Printf("Warning: Failed to send remaining data (size: %d): %v\n", len(dataCopy), err)
			}
		}
		req.safeCloseContentItemChan() // 替换原 ChContentItem 关闭
	}()
	for {
		if ctx != nil {
			select {
			case <-ctx.Done():
				// 处理中断信号，立即返回已读取的数据和错误
				return b, ctx.Err()
			default:
			}
		}
		n, err := r.Read(buf)
		if n == 0 && err == nil {
			return nil, fmt.Errorf("Read: n=0 with err=nil") // 出现这种情况时说明发生了未知错误
		}
		// 累积数据到结果
		if n > 0 {
			// 有数据到达，重置空闲计时器
			if idleTimer != nil {
				if !idleTimer.Stop() {
					select {
					case <-idleTimer.C:
					default:
					}
				}
				idleTimer.Reset(idleDuration)
			}
			if rp.CacheFullResponse {
				b = append(b, buf[:n]...)
			}
			// 根据内容类型选择处理方式
			if isText {
				bItem = append(bItem, buf[:n]...)
				// 文本模式按行发送
				if bytes.Contains(buf[:n], []byte("\n")) {
					dataCopy := make([]byte, len(bItem))
					copy(dataCopy, bItem)
					if err := sendData(dataCopy); err != nil {
						return b, err
					}
					bItem = bItem[:0] // 清空已发送数据
				}
			} else {
				// 二进制模式按块发送
				dataCopy := make([]byte, n)
				copy(dataCopy, buf[:n])
				if err := sendData(dataCopy); err != nil {
					return b, err
				}
			}
		}

		// 先处理错误前的残留数据，再处理错误
		if err != nil {
			if err == io.EOF && isText && len(bItem) > 0 {
				// 如果是EOF并且有残留数据，强制发送
				dataCopy := make([]byte, len(bItem))
				copy(dataCopy, bItem)
				if err := sendData(dataCopy); err != nil {
					// Log the failure, but allow function to return EOF
					fmt.Printf("Warning: Failed to send remaining text data (size: %d): %v\n", len(dataCopy), err)
				}
				bItem = bItem[:0]
			}
			if err == io.EOF {
				return b, nil // EOF 正常返回
			}
			return b, err // 其他错误返回错误，避免静默截断
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

// handleCallback 是一个泛型函数，用于处理任何 channel 和回调函数。
// T 是 channel 中传递的数据类型。
// req 是请求对象
func handleCallback[T any](ch <-chan T, f func(T, *Request), req *Request) {
	go func() {
		req.wgDone.Add(1)
		// 增加 recover 保护，防止回调函数 panic
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Callback panic: %v\n", r)
			}
			req.wgDone.Done() // 确保最终会执行 Done()
		}()

		// 捕获当前上下文快照，避免并发期间字段被后续请求覆盖
		localCtx := req.cancelCtx

		for {
			if localCtx != nil {
				select {
				case v, ok := <-ch:
					if !ok {
						// 信道正常关闭，退出循环
						return
					}
					f(v, req) // 执行回调逻辑
				case <-localCtx.Done():
					// 上下文被取消（如请求超时、完成），直接退出
					return
				}
			} else {
				v, ok := <-ch
				if !ok {
					return
				}
				f(v, req)
			}
		}
	}()
}

// OnHttpResponse 响应头回调
func (req *Request) OnHttpResponse(f func(httpRes *http.Response, req *Request)) {
	if req.chHttpResponse != nil {
		return
	}
	req.chHttpResponse = make(chan *http.Response, 1)
	handleCallback(req.chHttpResponse, f, req)
}

// OnUploaded 上传回调
func (req *Request) OnUploaded(f func(uploaded *int64, req *Request)) {
	if req.ChUploaded != nil {
		return
	}
	req.ChUploaded = make(chan *int64, 10)
	handleCallback(req.ChUploaded, f, req)
}

// OnContent 实现内容回调
func (req *Request) OnContent(f func(byteItem []byte, req *Request)) {
	if req.ChContentItem != nil {
		return
	}
	req.ChContentItem = make(chan []byte, 10)
	handleCallback(req.ChContentItem, f, req)
}
