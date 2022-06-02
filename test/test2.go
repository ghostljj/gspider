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

	//ss := gs.NewSpider2(gs.Encode("utf-8"))
	ss := gs.NewSpider()

	//ps := u.GetProxy(false, 0) //可能返回nil
	//if ps != nil {             //设置代理
	//  ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", ps.IP, ps.PORT))
	//}
	//ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10809))
	//ss.SetSocks5Proxy("127.0.0.1:10808", "", "")

	//strContent, err := ss.Post(strUrl, gs.NewRequestOptions(gs.OptRefererUrl("http://www.abc.com"),
	//	gs.OptPostData(`uid=&langx=zh-cn&mac=&ver=&JE=&radio=web_new&username=winner88&password=asdf1234&remember=on`)))

	ros := gs.NewRequestOptions(gs.OptRefererUrl("http://www.abc.com"),
		gs.OptRequestHeader(map[string]string{"hkey1": "hvalue1", "hkey2": "hvalue2"}),
		gs.OptRequestCookie("abc=123"))
	//上面时糖果语法。下面是傻瓜语法。一样ok
	//ros.RefererUrl = "http://www.abc.com"
	//ros.Header = map[string]string{"hkey1": "hvalue1", "hkey2": "hvalue2"}
	//ros.Cookie = "abc=123"
	strContent, err := ss.Get(strUrl, ros)
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
