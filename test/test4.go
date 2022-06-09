package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

func main() {
	var strUrl string
	strUrl = "https://www.google.com"

	req := gs.Session()
	req.HttpProxyAuto = true
	//req.Verify = true

	res := req.Get(strUrl)

	if res.GetErr() != nil {
		fmt.Println("Error=" + res.GetErr().Error())
	} else {
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(res.GetContent())
		res.PrintReqHeader("")            //打印 请求 头信息
		res.PrintReqPostData()            // 打印 请求 PostData
		res.PrintResHeader("")            //打印 响应 头信息
		res.PrintResSetCookie()           //打印 响应 头信息SetCookie
		res.PrintResUrl()                 // 打印 响应 最后的Url
		res.PrintCookies(res.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
		res.PrintStatusCode()             // 打印 响应 状态码
	}
	select {}
}
