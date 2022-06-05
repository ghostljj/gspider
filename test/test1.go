package main

import (
	"fmt"
	"github.com/asmcos/requests"
	gs "github.com/ghostljj/gspider"
)

func main() {

	resp, err := requests.Get("http://www.zhanluejia.net.cn")
	if err != nil {
		return
	}
	println(resp.Text())

	var strUrl string
	strUrl = "http://2022.ip138.com/ic.asp"
	//strUrl = "http://www.baidu.com"
	//strUrl = "http://www.google.com"

	req := gs.Session()
	req.RefererUrl = "http://www.baidu.com"
	req.Cookie = "aa=11;bb=22"
	req.Header = map[string]string{"h1": "v1", "h2": "v2"}
	//ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10809))
	//ss.SetSocks5Proxy("127.0.0.1:10808", "", "")

	res := req.Get(strUrl)
	res.Encode = "utf-8"
	if res.GetErr() != nil {
		fmt.Println("Error=" + res.GetErr().Error())
	} else {
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(res.GetContent())
		res.PrintReqHeader("") //打印 请求 头信息
		res.GetReqHeader()
		res.PrintReqPostData()            // 打印 请求 PostData
		res.PrintResHeader("")            //打印 响应 头信息
		res.PrintResSetCookie()           //打印 响应 头信息SetCookie
		res.PrintResUrl()                 // 打印 响应 最后的Url
		res.PrintCookies(res.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
		res.PrintStatusCode()             // 打印 响应 状态码

	}
}