package gspider

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

//--------------------------------------------------------------------------------------------------------------

func (ros *requests) setRequestOptions(strUrl string, opts ...requestsInterface) (refererUrl string, header map[string]string, redirectCount int) {
	ros.myMutex.Lock()
	defer ros.myMutex.Unlock()

	for _, opt := range opts {
		opt.apply(ros) //这里是塞入实体，针对实体赋值
	}
	if ros.Cookie != "" {
		ros.SetCookies(strUrl, ros.Cookie)
	}
	if ros.CookieAll != "" {
		ros.SetCookiesAll(strUrl, ros.CookieAll)
	}
	refererUrl, header, redirectCount = ros.RefererUrl, ros.Header, ros.RedirectCount
	return
}

func (ros *requests) request(strMethod, strUrl, strPostData string, opts ...requestsInterface) *response {
	refererUrl, header, redirectCount := ros.setRequestOptions(strUrl, opts...)
	return ros.send(strMethod, strUrl, strPostData, refererUrl, header, redirectCount)
}

func (ros *requests) Get(strUrl string, opts ...requestsInterface) *response {
	return ros.request("GET", strUrl, "", opts...)
}

func (ros *requests) GetJson(strUrl string, opts ...requestsInterface) *response {
	ros.isGetJson = 1
	return ros.request("GET", strUrl, "", opts...)
}
func (ros *requests) GetJsonR(strUrl, strPostData string, opts ...requestsInterface) *response {
	ros.isGetJson = 1
	if strPostData != "" {
		ros.isPostJson = 1
	}
	return ros.request("GET", strUrl, strPostData, opts...)
}

func (ros *requests) DeleteJson(strUrl string, opts ...requestsInterface) *response {
	ros.isGetJson = 1
	return ros.request("DELETE", strUrl, "", opts...)
}

//Post 方法
func (ros *requests) Post(strUrl, strPostData string, opts ...requestsInterface) *response {
	return ros.request("POST", strUrl, strPostData, opts...)
}
func (ros *requests) PostJson(strUrl, strPostData string, opts ...requestsInterface) *response {
	ros.isPostJson = 1
	ros.isGetJson = 1
	return ros.request("POST", strUrl, strPostData, opts...)
}

//Put Put方法
func (ros *requests) Put(strUrl, strPostData string, opts ...requestsInterface) *response {
	return ros.request("PUT", strUrl, strPostData, opts...)
}
func (ros *requests) PutJson(strUrl, strPostData string, opts ...requestsInterface) *response {
	ros.isGetJson = 1
	ros.isPostJson = 1
	return ros.request("PUT", strUrl, strPostData, opts...)
}

//获取img src 值
func (ros *requests) GetBase64ImageSrc(strUrl string, opts ...requestsInterface) (*response, string) {
	res, strContent := ros.GetBase64Image(strUrl, opts...)
	if res.GetErr() == nil {
		contentType := res.GetResHeader().Get("Content-Type")
		strContent = "data:" + contentType + ";base64," + strContent + ""
	}
	return res, strContent
}

//获取Base64 字符串
func (ros *requests) GetBase64Image(strUrl string, opts ...requestsInterface) (*response, string) {
	refererUrl, header, redirectCount := ros.setRequestOptions(strUrl, opts...)
	res := ros.send("GET", strUrl, "", refererUrl, header, redirectCount)
	return res, base64.StdEncoding.EncodeToString(res.GetBytes())
}

// SendRedirect 发送请求
// strMethod GET POST PUT ...
func (ros *requests) send(strMethod, strUrl, strPostData, refererUrl string, header map[string]string, redirectCount int) *response {

	res := newResponse()

	strMethod = strings.ToUpper(strMethod)
	reqURI, err := url.Parse(strUrl)
	if err != nil {
		res.err = err
		return res
	}
	res.reqUrl = reqURI.String()

	httpClient := &http.Client{}
	res.reqPostData = strPostData
	bytesPostData := bytes.NewBuffer([]byte(strPostData))
	httpReq, err := http.NewRequest(strMethod, strUrl, bytesPostData)
	if err != nil {
		res.err = err
		return res
	}

	ts := &http.Transport{}
	//超时设置  代理设置
	{
		netDialer := &net.Dialer{
			Timeout:   ros.Timeout * time.Second,                          //tcp 连接时设置的连接超时
			Deadline:  time.Now().Add(ros.ReadWriteTimeout * time.Second), //读写超时
			KeepAlive: ros.KeepAliveTimeout * time.Second,                 //保持连接超时设置
		}

		if len(ros.LocalIP) > 0 { //设置本地网络ip
			localAddr, err := net.ResolveIPAddr("ip", ros.LocalIP)
			if err != nil {
				res.err = err
				return res
			}
			localTCPAddr := net.TCPAddr{
				IP: localAddr.IP,
			}
			netDialer.LocalAddr = &localTCPAddr
		}

		ts.TLSHandshakeTimeout = 10 * time.Second   //限制执行TLS握手所花费的时间
		ts.ResponseHeaderTimeout = 10 * time.Second //限制读取response header的时间
		// ts.ExpectContinueTimeout = 1 * time.Second  //限制client在发送包含 Expect: 100-continue 的header到收到继续发送body的response之间的时间等待 POST才可能需要

		ts.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //跳过证书验证

		//ts.Dial = (netDialer).Dial //弃用，使用DialContext
		if len(ros.HttpProxyInfo) > 0 { //http 代理设置
			proxyUrl, err := url.Parse(ros.HttpProxyInfo)
			if err != nil {
				res.err = err
				return res
			}
			ts.Proxy = http.ProxyURL(proxyUrl)
			ts.DialContext = (netDialer).DialContext
		}
		if len(ros.Socks5Address) > 0 { //SOCKS5 代理设置
			var Socks5Auth *proxy.Auth
			if len(ros.Socks5User) > 0 {
				Socks5Auth = &proxy.Auth{User: ros.Socks5User, Password: ros.Socks5Pass} // 没有就不设置 就是nil
			}
			netDialerNew, err := proxy.SOCKS5("tcp", ros.Socks5Address,
				Socks5Auth,
				netDialer,
			)
			if err != nil {
				res.err = err
				return res
			}
			ts.Dial = (netDialerNew).Dial //SOCKS5 必须使用这个。。。
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

		for k, v := range ros.defaultHeaderTemplate {
			sendHeader[strings.ToLower(k)] = v
		}

		if ros.isGetJson == 1 { //接收json
			sendHeader[strings.ToLower(`accept`)] = `application/json, text/plain, */*`
		}
		if ros.isPostJson == 1 { //发送json
			sendHeader[strings.ToLower(`content-type`)] = `application/json;charset=UTF-8`
		} else if ros.isPostJson == 0 { //发送from
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

	httpClient.Jar = ros.cookieJar

	httpRes, err := httpClient.Do(httpReq)
	if err != nil {
		res.err = err
		return res
	}

	defer httpRes.Body.Close()
	var reader io.ReadCloser
	//解析gzip deflate
	{
		switch httpRes.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(httpRes.Body)
			if err != nil {
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
