package gspider

import (
	"bytes"
	"github.com/axgle/mahonia"
	"github.com/saintfish/chardet"
	"net/http"
	"strings"
)

// HttpInfo  返回信息结构
type RetHttpInfos struct {

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
	resStatusCode int
}

//NewHttpInfo  新建一个httpInfo
func newRetHttpInfos() RetHttpInfos {
	h := RetHttpInfos{}
	//清空 请求 Url
	h.reqUrl = ""
	//清空 请求 Post
	h.reqPostData = ""

	//清空 响应 头信息
	h.resHeader = nil
	//清空 请求 头信息
	h.reqHeader = nil
	//清空 响应 SetCookie
	//h.resCookies = s.resCookies[:0]
	h.resCookies = []*http.Cookie{}
	//清空 响应 后的Url
	h.resUrl = ""

	//清空
	h.resStatusCode = 0
	//清空返回内容
	h.resBytes = []byte{}
	//清空 错误信息
	h.err = nil
	return h
}

//GetReqHeader 获取 请求 头信息
func (ros *requests) GetReqHeader() http.Header {
	return ros.retHttpInfos.reqHeader
}

//GetResHeader 获取 响应 头信息
func (ros *requests) GetResHeader() http.Header {
	return ros.retHttpInfos.resHeader
}

//GetResCookies 获取 响应 Cookies
func (ros *requests) GetResCookies() []*http.Cookie {
	return ros.retHttpInfos.resCookies
}

//GetReqUrl 获取 请求 Url
func (ros *requests) GetReqUrl() string {
	return ros.retHttpInfos.reqUrl
}

//GetReqPostData 获取 请求 Post 信息
func (ros *requests) GetReqPostData() string {
	return ros.retHttpInfos.reqPostData
}

//GetResUrl 获取 响应 后的Url
func (ros *requests) GetResUrl() string {
	return ros.retHttpInfos.resUrl
}

//GetResStatusCode 获取 响应 状态码
func (ros *requests) GetResStatusCode() int {
	return ros.retHttpInfos.resStatusCode
}

//GetErr 返回错误
func (ros *requests) GetErr() error {
	return ros.retHttpInfos.err
}

//GetBytes 获取 响应 byte
func (ros *requests) GetBytes() []byte {
	return ros.retHttpInfos.resBytes
}

//GetContent 获取 响应 内容
func (ros *requests) GetContent() string {

	bodyByte := bytes.TrimPrefix(ros.retHttpInfos.resBytes, []byte("\xef\xbb\xbf")) // Or []byte{239, 187, 191}
	bodyStr := string(bodyByte)
	contentType := strings.ToLower(ros.GetResHeader().Get("Content-Type"))
	if strings.Index(contentType, "image/") <= -1 {
		//自动/手动 编码
		var charset string
		if strings.ToLower(ros.Encode) == "auto" {
			autoEncode, err := chardet.NewTextDetector().DetectBest(ros.retHttpInfos.resBytes)
			if err == nil {
				charset = autoEncode.Charset
			} else {
				ros.retHttpInfos.err = err
			}
		} else {
			charset = ros.Encode
		}
		//进行编码
		if charset != "" {
			encodeDec := mahonia.NewDecoder(charset)
			if encodeDec != nil {
				//在UTF-8字符转中，有可能会有一个BOM（字节顺序标记）这个字节顺序标记并不是必须的，有的 UTF-8 数据就是不带这个 BOM 的
				bodyStr = encodeDec.ConvertString(bodyStr) //把文本转为 srcCode 例如 GB18030
			}
		}
	}
	return bodyStr
}
