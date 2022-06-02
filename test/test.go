package main

import (
	"fmt"
	"github.com/ghostljj/gspider"
)

func main() {

	var strUrl string
	strUrl = "http://2022.ip138.com/ic.asp"

	ss := gspider.NewSpider()
	//ps := u.GetProxy(false, 0) //可能返回nil
	//if ps != nil {             //设置代理
	//	ss.HttpProxyInfo = fmt.Sprintf("http://%s:%d", ps.IP, ps.PORT)
	//}

	// ss.HttpProxyInfo = fmt.Sprintf("http://%s:%d", "127.0.0.1", 8888)
	// strContent, err := ss.Get(strUrl, strUrl, nil)

	// strUrl = "https://xxxx.com/app/member/login.php"
	// strContent, err := ss.Post(strUrl, strUrl, `uid=&langx=zh-cn&mac=&ver=&JE=&radio=web_new&username=winner88&password=asdf1234&remember=on`, nil)

	setHeader := make(map[string]string)
	setHeader["Connection"] = ""
	strContent, err := ss.Get(strUrl, "", nil)
	if err != nil {
		fmt.Println("Error=" + err.Error())
	} else {
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(strContent)
		ss.PrintReqHeader("")           //打印 请求 头信息
		ss.PrintReqPostData()           // 打印 请求 PostData
		ss.PrintResHeader("")           //打印 响应 头信息
		ss.PrintResSetCookie()          //打印 响应 头信息SetCookie
		ss.PrintResUrl()                // 打印 响应 最后的Url
		ss.PrintCookies(ss.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
		ss.PrintResStatusCode()         // 打印 响应 状态码

	}
}
