package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

func main() {

	var strUrl string
	strUrl = "https://192.168.211.211:9200/book/_search"
	strPostData := `{
						"from": 0,
						"size": 200,
						"query": {
							"match_all": {}
						}
					}`

	req := gs.Session()
	req.Verify = false

	res := req.GetJsonR(strUrl, strPostData,
		gs.OptHeader(map[string]string{"Authorization": "Basic ZWxhc3RpYzoxMjMzMjE="}))

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
