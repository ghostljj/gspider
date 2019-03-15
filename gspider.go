package gspider

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/saintfish/chardet"
	"golang.org/x/net/proxy"
)

// var InitValue = 3
// func init() {
//     InitValue = 5
// 	fmt.Println("sdf")
// }

// Spider 爬虫结构体
type Spider struct {
	//编码 默认 Auto 中文 GB18030  或 UTF-8
	Encode string
	// AllowAutoRedirect 是否重定向
	AllowAutoRedirect bool
	// HttpProxyInfo 设置Http代理 例：http://127.0.0.1:1081
	HttpProxyInfo string
	//Socks5地址 例：127.0.0.1:7813
	Socks5Address string
	//Socks5 用户名
	Socks5User string
	//Socks5 密码
	Socks5Pass string
	//Cookie
	cookieJar http.CookieJar
	//Timeout 连接超时
	Timeout time.Duration
	//ReadWriteTimeout 读写超时
	ReadWriteTimeout time.Duration
	//KeepAliveTimeout 保持连接超时
	KeepAliveTimeout time.Duration
	//headerTemplate Req Header 发送 请求 头
	headerTemplate map[string]string
	//返回 响应 头信息  map[string][]string  val是[]string
	resHeader http.Header
	//返回 响应 后的Url
	resUrl string
	//返回 响应 状态码
	resStatusCode int
	//返回 请求 头信息  map[string][]string  val是[]string
	reqHeader http.Header
	//发送请求的内容
	reqPostData string
}

// NewSpider  初始化一个爬虫
// spider
func NewSpider() Spider {
	s := Spider{}
	s.Encode = "Auto"
	s.AllowAutoRedirect = true
	s.cookieJar, _ = cookiejar.New(nil)
	s.Timeout = 30
	s.ReadWriteTimeout = 30
	s.KeepAliveTimeout = 30

	s.headerTemplate = make(map[string]string)
	s.headerTemplate["accept-encoding"] = "gzip, deflate"
	s.headerTemplate["accept-language"] = "zh-CN,zh;q=0.9"
	s.headerTemplate["connection"] = "keep-alive"
	s.headerTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"
	s.headerTemplate["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36"
	return s
}

//Get Get方法
func (s *Spider) Get(strUrl, refererUrl string, header map[string]string) (string, error) {
	return s.Send("GET", strUrl, refererUrl, "", header)
}

//Post Post方法
func (s *Spider) Post(strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {
	return s.Send("POST", strUrl, refererUrl, strPostData, header)
}

//	contentType := strings.ToLower(s.GetResHeader().Get("Content-Type"))
func (s *Spider) GetBase64ImageSrc(strUrl, refererUrl string, header map[string]string) (string, error) {
	strContent, err := s.GetBase64Image(strUrl, refererUrl, header)
	if err == nil {
		contentType := s.GetResHeader().Get("Content-Type")
		strContent = "data:" + contentType + ";base64," + strContent + ""
	}
	return strContent, err
}
func (s *Spider) GetBase64Image(strUrl, refererUrl string, header map[string]string) (string, error) {
	strContent, err := s.Send("GET", strUrl, refererUrl, "", header)
	strContent = base64.StdEncoding.EncodeToString([]byte(strContent))
	return strContent, err
}

//Put Put方法
func (s *Spider) Put(strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {
	return s.Send("PUT", strUrl, refererUrl, strPostData, header)
}

// Send 发送请求
// strMethod GET POST PUT ...
// strUrl
// header  发送头信息
func (s *Spider) Send(strMethod, strUrl, refererUrl, strPostData string, header map[string]string) (string, error) {

	strMethod = strings.ToUpper(strMethod)
	s.reqPostData = ""
	reqURI, err := url.Parse(strUrl)
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{}
	s.reqPostData = strPostData
	bytesPostData := bytes.NewBuffer([]byte(strPostData))
	httpReq, err := http.NewRequest(strMethod, reqURI.String(), bytesPostData)
	if err != nil {
		return "", err
	}

	ts := &http.Transport{}
	{ //超时设置  代理设置
		netDialer := &net.Dialer{
			Timeout:   s.Timeout * time.Second,                          //tcp 连接时设置的连接超时
			Deadline:  time.Now().Add(s.ReadWriteTimeout * time.Second), //读写超时
			KeepAlive: s.KeepAliveTimeout * time.Second,                 //保持连接超时设置
		}
		ts.TLSHandshakeTimeout = 10 * time.Second   //限制执行TLS握手所花费的时间
		ts.ResponseHeaderTimeout = 10 * time.Second //限制读取response header的时间
		// ts.ExpectContinueTimeout = 1 * time.Second  //限制client在发送包含 Expect: 100-continue 的header到收到继续发送body的response之间的时间等待 POST才可能需要

		ts.Dial = (netDialer).Dial
		if len(s.HttpProxyInfo) > 0 { //http 代理设置
			proxyUrl, err := url.Parse(s.HttpProxyInfo)
			if err != nil {
				return "", err
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
				return "", err
			}
			ts.Dial = (netDialerNew).Dial
		}
	}
	httpClient.Transport = ts

	if s.AllowAutoRedirect == false { //禁止重定向 默认重定向10次
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	{ //合并Header
		sendHeader := make(map[string]string)
		if strMethod == "POST" || strMethod == "PUT" {
			sendHeader["content-type"] = "application/x-www-form-urlencoded; charset=UTF-8"
		}
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
			httpReq.Header.Set(k, v)
		}
	}
	httpClient.Jar = s.cookieJar

	httpRes, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}

	defer httpRes.Body.Close()
	var reader io.ReadCloser
	{ //解析gzip default
		switch httpRes.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(httpRes.Body)
			if err != nil {
				return "", err
			}
		default:
			reader = httpRes.Body
		}
	}
	bodyByte, err := pedanticReadAll(reader)
	if err != nil {
		return "", err
	}

	bodyStr := string(bodyByte) //获取文本

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
		if charset != "" { //进行编码
			encodeDec := mahonia.NewDecoder(charset)
			if encodeDec != nil {
				bodyStr = encodeDec.ConvertString(bodyStr) //把文本转为 srcCode 例如 GB18030
			}
		}
	}

	return bodyStr, nil
}

//pedanticReadAll  读取所有字节
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
