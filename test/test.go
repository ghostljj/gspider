package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

func main() {

	var strUrl string
	strUrl = "http://2022.ip138.com/ic.asp"
	//strUrl = "http://www.baidu.com"
	//strUrl = "http://www.google.com"

	ss := gs.Session()
	ss.Encode = "utf-8"
	ss.RefererUrl = "http://www.baidu.com"
	ss.Cookie = "aa=11;bb=22"
	ss.Header = map[string]string{"h1": "v1", "h2": "v2"}
	//ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10809))
	//ss.SetSocks5Proxy("127.0.0.1:10808", "", "")

	ss.Get(strUrl)

	if ss.GetErr() != nil {
		fmt.Println("Error=" + ss.GetErr().Error())
	} else {
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(ss.GetContent())
		ss.PrintReqHeader("")           //打印 请求 头信息
		ss.PrintReqPostData()           // 打印 请求 PostData
		ss.PrintResHeader("")           //打印 响应 头信息
		ss.PrintResSetCookie()          //打印 响应 头信息SetCookie
		ss.PrintResUrl()                // 打印 响应 最后的Url
		ss.PrintCookies(ss.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
		ss.PrintResStatusCode()         // 打印 响应 状态码

	}
}
