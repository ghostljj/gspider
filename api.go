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
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

//--------------------------------------------------------------------------------------------------------------

func (req *Request) setRequestOptions(strUrl string, opts ...requestInterface) (refererUrl string, header map[string]string, redirectCount int) {
	req.myMutex.Lock()
	defer req.myMutex.Unlock()

	for _, opt := range opts {
		opt.apply(req) //这里是塞入实体，针对实体赋值
	}
	if req.Cookie != "" {
		req.SetCookies(strUrl, req.Cookie)
	}
	if req.CookieAll != "" {
		req.SetCookiesAll(strUrl, req.CookieAll)
	}
	refererUrl, header, redirectCount = req.RefererUrl, req.Header, req.RedirectCount
	return
}

func (req *Request) request(strMethod, strUrl, strPostData string, opts ...requestInterface) *Response {
	refererUrl, header, redirectCount := req.setRequestOptions(strUrl, opts...)
	return req.send(strMethod, strUrl, strPostData, refererUrl, header, redirectCount)
}

func (req *Request) Get(strUrl string, opts ...requestInterface) *Response {
	return req.request("GET", strUrl, "", opts...)
}

func (req *Request) GetJson(strUrl string, opts ...requestInterface) *Response {
	req.isGetJson = 1
	return req.request("GET", strUrl, "", opts...)
}
func (req *Request) GetJsonR(strUrl, strPostData string, opts ...requestInterface) *Response {
	req.isGetJson = 1
	if strPostData != "" {
		req.isPostJson = 1
	}
	return req.request("GET", strUrl, strPostData, opts...)
}

func (req *Request) DeleteJson(strUrl string, opts ...requestInterface) *Response {
	req.isGetJson = 1
	return req.request("DELETE", strUrl, "", opts...)
}

//Post 方法
func (req *Request) Post(strUrl, strPostData string, opts ...requestInterface) *Response {
	return req.request("POST", strUrl, strPostData, opts...)
}
func (req *Request) PostJson(strUrl, strPostData string, opts ...requestInterface) *Response {
	req.isPostJson = 1
	req.isGetJson = 1
	return req.request("POST", strUrl, strPostData, opts...)
}

//Put Put方法
func (req *Request) Put(strUrl, strPostData string, opts ...requestInterface) *Response {
	return req.request("PUT", strUrl, strPostData, opts...)
}
func (req *Request) PutJson(strUrl, strPostData string, opts ...requestInterface) *Response {
	req.isGetJson = 1
	req.isPostJson = 1
	return req.request("PUT", strUrl, strPostData, opts...)
}

//获取img src 值
func (req *Request) GetBase64ImageSrc(strUrl string, opts ...requestInterface) (*Response, string) {
	res, strContent := req.GetBase64Image(strUrl, opts...)
	if res.GetErr() == nil {
		contentType := res.GetResHeader().Get("Content-Type")
		strContent = "data:" + contentType + ";base64," + strContent + ""
	}
	return res, strContent
}

//获取Base64 字符串
func (req *Request) GetBase64Image(strUrl string, opts ...requestInterface) (*Response, string) {
	refererUrl, header, redirectCount := req.setRequestOptions(strUrl, opts...)
	res := req.send("GET", strUrl, "", refererUrl, header, redirectCount)
	return res, base64.StdEncoding.EncodeToString(res.GetBytes())
}

// SendRedirect 发送请求
// strMethod GET POST PUT ...
func (req *Request) send(strMethod, strUrl, strPostData, refererUrl string, header map[string]string, redirectCount int) *Response {

	res := newResponse(req)

	strMethod = strings.ToUpper(strMethod)
	reqURI, err := url.Parse(strUrl)
	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}
	res.reqUrl = reqURI.String()

	httpClient := &http.Client{}
	res.reqPostData = strPostData
	bytesPostData := bytes.NewBuffer([]byte(strPostData))
	httpReq, err := http.NewRequest(strMethod, strUrl, bytesPostData)
	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}

	ts := &http.Transport{}
	//超时设置  代理设置
	{
		netDialer := &net.Dialer{
			Timeout:   req.Timeout * time.Second,                          //tcp 连接时设置的连接超时
			Deadline:  time.Now().Add(req.ReadWriteTimeout * time.Second), //读写超时
			KeepAlive: req.KeepAliveTimeout * time.Second,                 //保持连接超时设置
		}

		if len(req.LocalIP) > 0 { //设置本地网络ip
			localAddr, err := net.ResolveIPAddr("ip", req.LocalIP)
			if err != nil {
				res.resBytes = []byte(err.Error())
				res.err = err
				return res
			}
			localTCPAddr := net.TCPAddr{
				IP: localAddr.IP,
			}
			netDialer.LocalAddr = &localTCPAddr
			ts.DialContext = (netDialer).DialContext
		}

		ts.TLSHandshakeTimeout = time.Second * 10   //限制执行TLS握手所花费的时间
		ts.ResponseHeaderTimeout = time.Second * 10 //限制读取response header的时间
		// ts.ExpectContinueTimeout = 1 * time.Second  //限制client在发送包含 Expect: 100-continue 的header到收到继续发送body的response之间的时间等待 POST才可能需要

		if req.Verify && req.tlsClientConfig != nil {
			ts.TLSClientConfig = req.tlsClientConfig
		} else {
			ts.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //跳过证书验证
		}

		var httpProxyInfoOK = ""
		if req.HttpProxyAuto {
			httpProxy := os.Getenv("http_proxy")
			httpProxys := os.Getenv("https_proxy")
			if len(httpProxy) > 0 {
				httpProxyInfoOK = httpProxy
				if strings.Index(httpProxyInfoOK, "http") == -1 {
					httpProxyInfoOK = "http://" + httpProxyInfoOK
				}
			} else if len(httpProxys) > 0 {
				httpProxyInfoOK = httpProxys
				if strings.Index(httpProxyInfoOK, "http") == -1 {
					httpProxyInfoOK = "https://" + httpProxyInfoOK
				}
			}
		}
		if len(req.HttpProxyInfo) > 0 {
			httpProxyInfoOK = req.HttpProxyInfo
		}
		//ts.Dial = (netDialer).Dial //弃用，使用DialContext
		if len(httpProxyInfoOK) > 0 { //http 代理设置
			proxyUrl, err := url.Parse(httpProxyInfoOK)
			if err != nil {
				res.err = err
				res.resBytes = []byte(err.Error())
				return res
			}
			ts.Proxy = http.ProxyURL(proxyUrl)
			ts.DialContext = (netDialer).DialContext
		}
		if len(req.Socks5Address) > 0 { //SOCKS5 代理设置
			var Socks5Auth *proxy.Auth
			if len(req.Socks5User) > 0 {
				Socks5Auth = &proxy.Auth{User: req.Socks5User, Password: req.Socks5Pass} // 没有就不设置 就是nil
			}
			netDialerNew, err := proxy.SOCKS5("tcp", req.Socks5Address,
				Socks5Auth,
				netDialer,
			)
			if err != nil {
				res.resBytes = []byte(err.Error())
				res.err = err
				return res
			}

			//if contextDialer, ok := netDialerNew.(proxy.ContextDialer); ok {
			//	ts.DialContext = contextDialer.DialContext
			//}
			ts.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
				return netDialerNew.Dial(network, address)
			}
		}
	}
	httpClient.Transport = ts

	//设置重定向次数 默认重定向10次
	if redirectCount > 0 {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			// 没有重定向不会执行，len(via)==1 就是第一次跳进入。选择是否跳
			if len(via) >= redirectCount {
				return http.ErrUseLastResponse //返回err就是，不跳
			}
			return nil //返回nil就是跳，
		}
	}

	//合并Header
	{
		sendHeader := make(map[string]string)
		if len(refererUrl) > 0 {
			sendHeader["referer"] = refererUrl
		}

		for k, v := range req.defaultHeaderTemplate {
			sendHeader[strings.ToLower(k)] = v
		}

		if req.isGetJson == 1 { //接收json
			sendHeader[strings.ToLower(`accept`)] = `application/json, text/plain, */*`
		}
		if req.isPostJson == 1 { //发送json
			sendHeader[strings.ToLower(`content-type`)] = `application/json;charset=UTF-8`
		} else if req.isPostJson == 0 { //发送from
			sendHeader[strings.ToLower(`content-type`)] = `application/x-www-form-urlencoded; charset=UTF-8`
		}

		for k, v := range header {
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

	httpRes, err := httpClient.Do(httpReq)

	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
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

	if err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}

	defer httpRes.Body.Close()

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

	if res.resBytes, err = pedanticReadAll(reader); err != nil {
		res.resBytes = []byte(err.Error())
		res.err = err
		return res
	}

	return res
}

//pedanticReadAll 读取所有字节
func pedanticReadAll(r io.Reader) (b []byte, err error) {
	var bufa [64]byte
	buf := bufa[:]
	for {
		n, err := r.Read(buf)
		if n == 0 && err == nil {
			return nil, fmt.Errorf("Read: n=0 with err=nil")
		}
		b = append(b, buf[:n]...)
		if err == io.EOF {
			n, err := r.Read(buf)
			if n != 0 || err != io.EOF {
				return nil, fmt.Errorf("Read: n=%d err=%#v after EOF", n, err)
			}
			return b, nil
		}
		if err != nil {
			return b, err
		}
	}
}
