package gspider

import (
	"net/http"
	"sync"
	"time"
)

//--------------------------------------------------------------------------------------------------------------

type SendHeader map[string]string
type SendFromData map[string]string
type SendJsonData string
type SendData string
type SendCookie map[string]string
type SendCookieAll map[string]string

//编译不成功。
//func Test(args ...interface {
//	SendHeader | SendFromData | SendJsonData | SendData | SendCookie | SendCookieAll
//}) {
//	for _, arg := range args {
//		switch a := any(arg).(type) {
//		case SendCookie:
//			for k, v := range a {
//				fmt.Println(k, v)
//			}
//			// arg is "GET" params
//			// ?title=website&id=1860&from=login
//		case SendJsonData:
//			fmt.Println(a, "字符串")
//		default:
//			Log.Printf("未处理 %s 参数\n", a)
//		}
//	}
//}

type request struct {
	myMutex       sync.Mutex        //本对象的同步对象
	LocalIP       string            //本地 网络 IP
	RefererUrl    string            //来源url
	isGetJson     int               //是否接收 Json  -1不发 0否 1是
	isPostJson    int               //是否提交 Json  -1不发 0否 1是
	Header        map[string]string //头参数
	Cookie        string            //cookie     单独url
	CookieAll     string            //cookieAll  根url+单独url
	RedirectCount int               //重定向次数

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
}

//defaultRequestOptions 默认配置参数
func defaultRequestOptions() *request {
	ros := request{
		isPostJson:    -1,
		isGetJson:     -1,
		Header:        make(map[string]string),
		RedirectCount: 10,

		Timeout:          30,
		ReadWriteTimeout: 30,
		KeepAliveTimeout: 30,
	}

	ros.ResetCookie()
	ros.defaultHeaderTemplate = make(map[string]string)
	ros.defaultHeaderTemplate["accept-encoding"] = "gzip, deflate"
	ros.defaultHeaderTemplate["accept-language"] = "zh-CN,zh;q=0.9"
	ros.defaultHeaderTemplate["connection"] = "keep-alive"
	ros.defaultHeaderTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"
	ros.defaultHeaderTemplate["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36"
	return &ros
}

func Session() *request {
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
	apply(*request)
}

//funcRequestOption 定义面的接口使用
type funcRequests struct {
	anyfun func(*request)
}

//apply 实现上面的接口，使用这个匿名函数，针对传入的对象，进行操作
func (fro *funcRequests) apply(ro *request) {
	fro.anyfun(ro)
}

//newFuncRequestOption 新建一个匿名函数实体。
//返回接口地址
func newFuncRequests(anonfun func(ro *request)) *funcRequests {
	return &funcRequests{
		anyfun: anonfun,
	}
}

//OptRefererUrl 设置来源地址，返回接口指针(新建一个函数，不执行的，返回他的地址而已)
func OptRefererUrl(refererUrl string) requestsInterface {
	//return &funcRequests{
	//	anyfun: func(ro *requests) {
	//		ro.RefererUrl = refererUrl
	//	},
	//}
	//下面更简洁而已，上门原理一致
	return newFuncRequests(func(ro *request) {
		ro.RefererUrl = refererUrl
	})
}

//OptHeader 设置发送头
func OptHeader(header map[string]string) requestsInterface {
	return newFuncRequests(func(ro *request) {
		ro.Header = header
	})
}

//OptRedirectCount 重定向次数
func OptRedirectCount(redirectCount int) requestsInterface {
	return newFuncRequests(func(ro *request) {
		ro.RedirectCount = redirectCount
	})
}

//OptCookie 设置当前Url cookie
func OptCookie(cookie string) requestsInterface {
	return newFuncRequests(func(ro *request) {
		ro.Cookie = cookie
	})
}

//OptCookieAll 设置当前Url+根Url cookie
func OptCookieAll(cookieAll string) requestsInterface {
	return newFuncRequests(func(ro *request) {
		ro.CookieAll = cookieAll
	})
}
