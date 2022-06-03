package gspider

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Spider  3爬虫结构体
type Spider struct {
	//本地 网络 IP
	localIP string

	encode           string        // 编码 默认 Auto 中文 GB18030  或 UTF-8
	timeout          time.Duration // 连接超时
	readWriteTimeout time.Duration // 读写超时
	keepAliveTimeout time.Duration // 保持连接超时
	httpProxyInfo    string        // 设置Http代理 例：http://127.0.0.1:1081
	socks5Address    string        //Socks5地址 例：127.0.0.1:7813
	socks5User       string        //Socks5 用户名
	socks5Pass       string        //Socks5 密码

	//Cookie
	cookieJar http.CookieJar
	//发送 请求 头
	defaultHeaderTemplate map[string]string
	//发送 请求 的Url
	reqUrl string
	//发送 请求 的内容
	reqPostData string
	//返回 请求 头信息  map[string][]string  val是[]string
	reqHeader http.Header
	//返回 响应 头信息  map[string][]string  val是[]string
	resHeader http.Header
	//返回当前Set-Cookie
	resCookies []*http.Cookie
	//返回内容[]byte
	resBytes []byte
	//返回内容
	resContent string
	//返回 错误 信息 没错返回nil
	Err error
	//返回 响应 后的Url
	resUrl string
	//返回 响应 状态码
	resStatusCode int
}

// NewSpider  初始化一个爬虫Spider
func NewSpider() Spider {
	s := Spider{}
	s.cookieJar, _ = cookiejar.New(nil)
	s.encode = "Auto"
	s.timeout = 30
	s.readWriteTimeout = 30
	s.keepAliveTimeout = 30
	s.defaultHeaderTemplate = make(map[string]string)
	s.defaultHeaderTemplate["accept-encoding"] = "gzip, deflate"
	s.defaultHeaderTemplate["accept-language"] = "zh-CN,zh;q=0.9"
	s.defaultHeaderTemplate["connection"] = "keep-alive"
	s.defaultHeaderTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"
	s.defaultHeaderTemplate["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36"
	return s
}

//// NewSpider  初始化一个爬虫Spider
//func NewSpider2(opts ...SpiderOptionInterface) Spider {
//	s := Spider{}
//	s.cookieJar, _ = cookiejar.New(nil)
//
//	s.dopts = defaultSpiderOptions()
//
//	for _, opt := range opts {
//		opt.apply(&s.dopts) //这里是塞入实体，针对实体赋值
//	}
//
//	s.headerTemplate = make(map[string]string)
//	s.headerTemplate["accept-encoding"] = "gzip, deflate"
//	s.headerTemplate["accept-language"] = "zh-CN,zh;q=0.9"
//	s.headerTemplate["connection"] = "keep-alive"
//	s.headerTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"
//	s.headerTemplate["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36"
//	return s
//}

//清空 请求 和 响应 信息
func (s *Spider) ClearResReqInfo() {
	//清空 请求 Url
	s.reqUrl = ""
	//清空 请求 Post
	s.reqPostData = ""

	//清空 响应 头信息
	s.resHeader = nil
	//清空 请求 头信息
	s.reqHeader = nil
	//清空 响应 SetCookie
	s.resCookies = s.resCookies[:0]
	s.resCookies = []*http.Cookie{}
	//清空 响应 后的Url
	s.resUrl = ""

	//清空
	s.resStatusCode = 0
	//清空返回内容
	s.resBytes = []byte{}
	//清空 内容
	s.resContent = ""
	//清空 错误信息
	s.Err = nil
}