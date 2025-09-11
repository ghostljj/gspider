package gspider

import (
	"bytes"
	"github.com/axgle/mahonia"
	"github.com/saintfish/chardet"
	"net/http"
	"strings"
)

// HttpInfo  返回信息结构
type Response struct {
	Encode string // 编码 默认 Auto 中文 GB18030  或 UTF-8
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
	//返回 错误 信息 没错返回nil
	err error
	//返回 响应 后的Url
	resUrl string
	//返回 响应 状态码
	statusCode int
	//请求对象
	req *Request
}

// NewHttpInfo  新建一个httpInfo
func newResponse(req *Request) *Response {
	res := Response{}
	res.Encode = "Auto"
	//清空 请求 Url
	res.reqUrl = ""
	//清空 请求 Post
	res.reqPostData = ""

	//清空 响应 头信息
	res.resHeader = nil
	//清空 请求 头信息
	res.reqHeader = nil
	//清空 响应 SetCookie
	//h.resCookies = s.resCookies[:0]
	res.resCookies = []*http.Cookie{}
	//清空 响应 后的Url
	res.resUrl = ""

	//清空
	res.statusCode = 0
	//清空返回内容
	res.resBytes = []byte{}
	//清空 错误信息
	res.err = nil
	//请求对象赋值
	res.req = req

	return &res
}

// GetReqHeader 获取 请求 头信息
func (res *Response) GetReqHeader() http.Header {
	return res.reqHeader
}

// GetResHeader 获取 响应 头信息
func (res *Response) GetResHeader() http.Header {
	return res.resHeader
}

// GetResCookies 获取 响应 Cookies
func (res *Response) GetResCookies() []*http.Cookie {
	return res.resCookies
}

// GetAllCookies 获取所有 Cookies 可以是 请求url \ 响应url
func (res *Response) GetAllCookies(strUrl string) *map[string]string {
	return res.req.GetCookiesJarMap(strUrl)
}

// GetReqUrl 获取 请求 Url
func (res *Response) GetReqUrl() string {
	return res.reqUrl
}

// GetReqPostData 获取 请求 Post 信息
func (res *Response) GetReqPostData() string {
	return res.reqPostData
}

// GetResUrl 获取 响应 后的Url
func (res *Response) GetResUrl() string {
	return res.resUrl
}

// GetStatusCode 获取 响应 状态码
func (res *Response) GetStatusCode() int {
	return res.statusCode
}

// GetBytes 获取 响应 byte
func (res *Response) GetBytes() []byte {
	return res.resBytes
}

// GetErr 返回错误
func (res *Response) GetErr() error {
	return res.err
}

// GetContent 获取 响应 内容
func (res *Response) GetContent() string {
	// 处理BOM
	bodyByte := res.resBytes
	if bytes.HasPrefix(bodyByte, []byte("\xef\xbb\xbf")) { // UTF-8 BOM
		bodyByte = bytes.TrimPrefix(bodyByte, []byte("\xef\xbb\xbf"))
	} else if bytes.HasPrefix(bodyByte, []byte("\xff\xfe")) { // UTF-16 LE BOM
		bodyByte = bytes.TrimPrefix(bodyByte, []byte("\xff\xfe"))
	} else if bytes.HasPrefix(bodyByte, []byte("\xfe\xff")) { // UTF-16 BE BOM
		bodyByte = bytes.TrimPrefix(bodyByte, []byte("\xfe\xff"))
	}

	bodyStr := string(bodyByte)
	contentType := strings.ToLower(res.resHeader.Get("Content-Type"))
	// 图片类型直接返回原始字节的字符串表示
	if strings.Index(contentType, "image/") != -1 {
		return bodyStr
	}
	//自动/手动 编码
	var charset string
	if strings.ToLower(res.Encode) == "auto" {
		autoEncode, err := chardet.NewTextDetector().DetectBest(res.resBytes)
		if err == nil {
			charset = autoEncode.Charset
		} else {
			res.err = err
		}
	} else {
		charset = res.Encode
	}

	// 如果没有指定编码，默认使用 UTF-8
	if charset == "" {
		charset = "UTF-8"
	}
	encodeDec := mahonia.NewDecoder(charset)
	if encodeDec != nil {
		//在UTF-8字符转中，有可能会有一个BOM（字节顺序标记）这个字节顺序标记并不是必须的，有的 UTF-8 数据就是不带这个 BOM 的
		bodyStr = encodeDec.ConvertString(bodyStr) //把文本转为 srcCode 例如 GB18030
	}

	return bodyStr
}
