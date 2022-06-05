package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

func main() {

	var strUrl string
	strUrl = "https://2022.ip138.com/ic.asp"

	//strUrl = "http://www.baidu.com"
	//strUrl = "http://www.google.com"

	//ps := u.GetProxy(false, 0) //可能返回nil
	//if ps != nil {             //设置代理
	//  ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", ps.IP, ps.PORT))
	//}

	req := gs.Session()

	res := req.Get(strUrl, gs.OptRefererUrl("http://www.abc.com"),
		gs.OptHeader(map[string]string{"hkey1": "hvalue1", "hkey2": "hvalue2"}),
		gs.OptCookie("abc=123;ddd=222"))

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
}
