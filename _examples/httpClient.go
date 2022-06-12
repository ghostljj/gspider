package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
	"path/filepath"
	"runtime"
)

func main() {

	var strUrl string
	//证书需要指定域名，在host里面设置好
	//证书还有过期，最好自己生成拉
	strUrl = "https://www.test.example.com:444"
	req := gs.Session()
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	req.SetmTLSClientFile(currentDir+"/x509/c.crt", currentDir+"/x509/c.key", currentDir+"/x509/s.ca")
	//SetTLSClientFile  单向用这个
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
}
