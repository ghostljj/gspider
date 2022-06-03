package gspider

import (
	"net/http"
	"sync"
	"time"
)

//--------------------------------------------------------------------------------------------------------------

type requests struct {
	myMutex       sync.Mutex        //本对象的同步对象
	LocalIP       string            //本地 网络 IP
	RefererUrl    string            //来源url
	isGetJson     int               //是否接收 Json  -1不发 0否 1是
	isPostJson    int               //是否提交 Json  -1不发 0否 1是
	Header        map[string]string //头参数
	Cookie        string            //cookie     单独url
	CookieAll     string            //cookieAll  根url+单独url
	RedirectCount int               //重定向次数

	Encode           string        // 编码 默认 Auto 中文 GB18030  或 UTF-8
	Timeout          time.Duration // 连接超时
	ReadWriteTimeout time.Duration // 读写超时
	KeepAliveTimeout time.Duration // 保持连接超时
	HttpProxyInfo    string        // 设置Http代理 例：http://127.0.0.1:1081
	Socks5Address    string        //Socks5地址 例：127.0.0.1:7813
	Socks5User       string        //Socks5 用户名
	Socks5Pass       string        //Socks5 密码

	//Cookie
	cookieJar http.CookieJar
	//发送 请求 头
	defaultHeaderTemplate map[string]string
	//发送后接收的信息
	retHttpInfos RetHttpInfos
}

//defaultRequestOptions 默认配置参数
func defaultRequestOptions() *requests {
	ros := requests{
		isPostJson:    -1,
		isGetJson:     -1,
		Header:        make(map[string]string),
		RedirectCount: 10,

		Encode:           "Auto",
		Timeout:          30,
		ReadWriteTimeout: 30,
		KeepAliveTimeout: 30,
	}
	ros.retHttpInfos = newRetHttpInfos()

	ros.ResetCookie()
	ros.defaultHeaderTemplate = make(map[string]string)
	ros.defaultHeaderTemplate["accept-encoding"] = "gzip, deflate"
	ros.defaultHeaderTemplate["accept-language"] = "zh-CN,zh;q=0.9"
	ros.defaultHeaderTemplate["connection"] = "keep-alive"
	ros.defaultHeaderTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"
	ros.defaultHeaderTemplate["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36"
	return &ros
}

func Session() *requests {
	return defaultRequestOptions()
	//dros := defaultRequestOptions()
	//for _, opt := range opts {
	//	opt.apply(&dros) //这里是塞入实体，针对实体赋值
	//}
	//return &dros
}

//--------------------------------------------------------------------------------------------------------------

// NewRequestOptions请求参数 采集基本接口
type requestsInterface interface {
	apply(*requests)
}

//funcRequestOption 定义面的接口使用
type funcRequests struct {
	anyfun func(*requests)
}

//apply 实现上面的接口，使用这个匿名函数，针对传入的对象，进行操作
func (fro *funcRequests) apply(ro *requests) {
	fro.anyfun(ro)
}

//newFuncRequestOption 新建一个匿名函数实体。
//返回接口地址
func newFuncRequests(anonfun func(*requests)) *funcRequests {
	return &funcRequests{
		anyfun: anonfun,
	}
}

//OptRefererUrl 设置来源地址，返回接口指针(新建一个函数，不执行的，返回他的地址而已)
func OptRefererUrl(refererUrl string) requestsInterface {
	//return &newFuncRequests{
	//	anyfun: func(ro *RequestOptions) {
	//		ro.refererUrl = refererUrl
	//	},
	//}
	//下面更简洁而已，上门原理一致
	return newFuncRequests(func(ro *requests) {
		ro.RefererUrl = refererUrl
	})
}

//OptHeader 设置发送头
func OptHeader(header map[string]string) requestsInterface {
	return newFuncRequests(func(ro *requests) {
		ro.Header = header
	})
}

//OptRedirectCount 重定向次数
func OptRedirectCount(redirectCount int) requestsInterface {
	return newFuncRequests(func(ro *requests) {
		ro.RedirectCount = redirectCount
	})
}

//OptCookie 设置当前Url cookie
func OptCookie(cookie string) requestsInterface {
	return newFuncRequests(func(ro *requests) {
		ro.Cookie = cookie
	})
}

//OptCookieAll 设置当前Url+根Url cookie
func OptCookieAll(cookieAll string) requestsInterface {
	return newFuncRequests(func(ro *requests) {
		ro.CookieAll = cookieAll
	})
}
