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

	"github.com/axgle/mahonia"
	"github.com/saintfish/chardet"
	"golang.org/x/net/proxy"
)

//Get 方法
func (s *Spider) Get(strUrl, refererUrl string, header map[string]string) (string, error) {
	return s.Send("GET", strUrl, refererUrl, "", header)
}
func (s *Spider) GetJson(strUrl, refererUrl string, header map[string]string) (string, error) {
	addHeader := make(map[string]string)
	addHeader[`accept`] = `application/json, text/plain, */*`
	for k, v := range header {
		addHeader[strings.ToLower(k)] = v
	}
	return s.Send("GET", strUrl, refererUrl, "", addHeader)
}

//Post 方法
func (s *Spider) Post(strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {
	addHeader := make(map[string]string)
	addHeader[`content-type`] = `application/x-www-form-urlencoded; charset=UTF-8`
	for k, v := range header {
		addHeader[strings.ToLower(k)] = v
	}
	return s.Send("POST", strUrl, refererUrl, strPostData, addHeader)
}
func (s *Spider) PostJson(strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {
	addHeader := make(map[string]string)
	addHeader[`accept`] = `application/json, text/plain, */*`
	addHeader[`content-type`] = `application/json;charset=UTF-8`
	for k, v := range header {
		addHeader[strings.ToLower(k)] = v
	}
	return s.Send("POST", strUrl, refererUrl, strPostData, addHeader)
}

//Put Put方法
func (s *Spider) Put(strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {
	addHeader := make(map[string]string)
	addHeader[`content-type`] = `application/x-www-form-urlencoded; charset=UTF-8`
	for k, v := range header {
		addHeader[strings.ToLower(k)] = v
	}
	return s.Send("PUT", strUrl, refererUrl, strPostData, addHeader)
}
func (s *Spider) PutJson(strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {
	addHeader := make(map[string]string)
	addHeader[`accept`] = `application/json, text/plain, */*`
	addHeader[`content-type`] = `application/json;charset=UTF-8`
	for k, v := range header {
		addHeader[strings.ToLower(k)] = v
	}
	return s.Send("PUT", strUrl, refererUrl, strPostData, addHeader)
}

//获取img src 值
func (s *Spider) GetBase64ImageSrc(strUrl, refererUrl string, header map[string]string) (string, error) {
	strContent, err := s.GetBase64Image(strUrl, refererUrl, header)
	if err == nil {
		contentType := s.GetResHeader().Get("Content-Type")
		strContent = "data:" + contentType + ";base64," + strContent + ""
	}
	return strContent, err
}

//获取Base64 字符串
func (s *Spider) GetBase64Image(strUrl, refererUrl string, header map[string]string) (string, error) {
	strContent, err := s.Send("GET", strUrl, refererUrl, "", header)
	strContent = base64.StdEncoding.EncodeToString([]byte(strContent))
	return strContent, err
}

func (s *Spider) Send(strMethod, strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {
	return s.SendRedirect(strMethod, strUrl, refererUrl, strPostData, header, 10)
}

// SendRedirect 发送请求
// strMethod GET POST PUT ...
// strUrl refererUrl 网址 与 来源网址
// header  发送头信息
func (s *Spider) SendRedirect(strMethod, strUrl, refererUrl, strPostData string, header map[string]string, redirectCount int) (string, error) {

	s.ClearResReqInfo()

	strMethod = strings.ToUpper(strMethod)
	reqURI, err := url.Parse(strUrl)
	if err != nil {
		return s.resContent, err
	}
	s.reqUrl = reqURI.String()

	httpClient := &http.Client{}
	s.reqPostData = strPostData
	bytesPostData := bytes.NewBuffer([]byte(strPostData))
	httpReq, err := http.NewRequest(strMethod, s.reqUrl, bytesPostData)
	if err != nil {
		return s.resContent, err
	}

	ts := &http.Transport{}
	//超时设置  代理设置
	{
		netDialer := &net.Dialer{
			Timeout:   s.Timeout * time.Second,                          //tcp 连接时设置的连接超时
			Deadline:  time.Now().Add(s.ReadWriteTimeout * time.Second), //读写超时
			KeepAlive: s.KeepAliveTimeout * time.Second,                 //保持连接超时设置
		}

		if len(s.localIP) > 0 { //设置本地网络ip
			localAddr, err := net.ResolveIPAddr("ip", s.localIP)
			if err != nil {
				return s.resContent, err
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

		ts.Dial = (netDialer).Dial
		if len(s.HttpProxyInfo) > 0 { //http 代理设置
			proxyUrl, err := url.Parse(s.HttpProxyInfo)
			if err != nil {
				return s.resContent, err
			}
			ts.Proxy = http.ProxyURL(proxyUrl)
		}
		if len(s.Socks5Address) > 0 { //SOCKS5 代理设置
			var Socks5Auth *proxy.Auth
			if len(s.Socks5User) > 0 {
				Socks5Auth = &proxy.Auth{User: s.Socks5User, Password: s.Socks5Pass} // 没有就不设置 就是nil
			}
			netDialerNew, err := proxy.SOCKS5("tcp", s.Socks5Address,
				Socks5Auth,
				netDialer,
			)
			if err != nil {
				return s.resContent, err
			}
			ts.Dial = (netDialerNew).Dial
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
			sendHeader["Referer"] = refererUrl
		}
		for k, v := range s.headerTemplate {
			sendHeader[strings.ToLower(k)] = v
		}
		for k, v := range header {
			sendHeader[strings.ToLower(k)] = v
		}
		for k, v := range sendHeader {
			if len(v) <= 0 {
				httpReq.Header.Del(k)
			} else {
				httpReq.Header.Set(k, v)
			}
		}
	}

	httpClient.Jar = s.cookieJar

	httpRes, err := httpClient.Do(httpReq)
	if err != nil {
		return s.resContent, err
	}

	defer httpRes.Body.Close()
	var reader io.ReadCloser
	//解析gzip deflate
	{
		switch httpRes.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(httpRes.Body)
			if err != nil {
				return s.resContent, err
			}
		case "deflate":
			reader = flate.NewReader(httpRes.Body)
		default:
			reader = httpRes.Body
		}
	}

	if s.resBytes, err = pedanticReadAll(reader); err != nil {
		return s.resContent, err
	}
	//在UTF-8字符转中，有可能会有一个BOM（字节顺序标记）这个字节顺序标记并不是必须的，有的 UTF-8 数据就是不带这个 BOM 的
	bodyByte := bytes.TrimPrefix(s.resBytes, []byte("\xef\xbb\xbf")) // Or []byte{239, 187, 191}
	bodyStr := string(bodyByte)                                      //获取文本

	//返回 响应 Cookies
	s.resCookies = httpRes.Cookies()
	//设置 响应 头信息
	s.resHeader = httpRes.Header
	//设置 请求 头信息
	s.reqHeader = httpRes.Request.Header
	//设置 响应 后的Url
	s.resUrl = httpRes.Request.URL.String()
	//设置响应状态码
	s.resStatusCode = httpRes.StatusCode

	contentType := strings.ToLower(s.GetResHeader().Get("Content-Type"))
	if strings.Index(contentType, "image/") <= -1 {
		//自动/手动 编码
		var charset string
		if strings.ToLower(s.Encode) == "auto" {
			autoEncode, err := chardet.NewTextDetector().DetectBest(bodyByte)
			if err == nil {
				charset = autoEncode.Charset
			}
		} else {
			charset = s.Encode
		}
		//进行编码
		if charset != "" {
			encodeDec := mahonia.NewDecoder(charset)
			if encodeDec != nil {
				bodyStr = encodeDec.ConvertString(bodyStr) //把文本转为 srcCode 例如 GB18030
			}
		}
	}
	s.resContent = bodyStr
	return s.resContent, nil
}

//读取所有字节
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
