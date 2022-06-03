package gspider

import (
	"fmt"
	"strconv"
)

//PrintInfo
func (ros *requests) PrintInfo() {
	ros.PrintReqHeader("")            // 打印 请求  信息
	ros.PrintResHeader("")            // 打印 响应 头信息
	ros.PrintResSetCookie()           // 打印 响应 头信息SetCookie
	ros.PrintReqUrl()                 // 打印 请求 Url
	ros.PrintReqPostData()            // 打印 请求 PostData
	ros.PrintResUrl()                 // 打印 响应 最后的Url
	ros.PrintCookies(ros.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
	ros.PrintResStatusCode()          // 打印 响应 状态码
}

//PrintReqHeader 打印 请求 头信息 查看信息用
func (ros *requests) PrintReqHeader(key string) {
	if key == "" {
		fmt.Println("------------------  Req 请求 Header | GetReqHeader() map[string][]string")
		fmt.Println("------------------  使用 GetReqHeader() http.Header 例如 GetReqHeader().Get(\"User-Agent\") ")

		for k, v := range ros.GetReqHeader() {
			fmt.Print("------------------          " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------------------------------------------")
	} else {
		fmt.Println("------------------  Req Header  ==> (" + ros.GetReqHeader().Get(key) + ") End") // 例如 User-Agent
	}
	fmt.Println("")
}

//PrintReqPostData 打印 请求 Post内容
func (ros *requests) PrintReqPostData() {
	fmt.Println("------------------  Req PostData ==> ( " + ros.retHttpInfos.reqPostData + " )")
	fmt.Println("")
}

//PrintResHeader 打印 响应 头信息
func (ros *requests) PrintResHeader(key string) {
	if key == "" {
		fmt.Println("------------------  Res 响应 Header")
		fmt.Println("------------------  使用 GetResHeader() http.Header 例如 GetResHeader().Get(\"Content-Encoding\") ")
		for k, v := range ros.GetResHeader() {
			fmt.Print("------------------          " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------------------------------------------")
	} else {
		fmt.Println("------------------  Res Header  ==> (" + ros.GetResHeader().Get(key) + ") End") // 例如 Content-Encoding
	}
	fmt.Println("")
}

//PrintResSetCookie
func (ros *requests) PrintResSetCookie() {
	fmt.Println("------------------  S Res 响应 Set-Cookie ")
	fmt.Println("------------------  使用 GetResCookies() []*http.Cookie ")
	for _, itemCookie := range ros.GetResCookies() {
		fmt.Println("------------------          ", itemCookie)
	}
	fmt.Println("------------------------------------------------------")
	fmt.Println()
}

//PrintReqUrl 打印 请求 URL
func (ros *requests) PrintReqUrl() {
	fmt.Println("------------------  Req Url 请求 URL ==> (" + ros.retHttpInfos.reqUrl + ") End")
	fmt.Println("")
}

//PrintResUrl 打印最后 响应 URL
func (ros *requests) PrintResUrl() {
	fmt.Println("------------------  Res Url 最后 响应 URL ==> (" + ros.retHttpInfos.resUrl + ") End") // 例如 Content-Encoding
	fmt.Println("")
}

//PrintCookies 打印CookieJar
func (ros *requests) PrintCookies(strUrl string) {
	fmt.Println("------------------  S CookieJar  ==> From(" + strUrl + ")")
	fmt.Println("------------------  使用 GetCookiesMap(strUrl string) map[string]string")
	defer func() {
		fmt.Println("------------------------------------------------------")
		fmt.Println("")
	}()
	if ros.cookieJar == nil {
		return
	}
	mapCookieHost := ros.GetCookiesMap(strUrl)
	for k, v := range mapCookieHost {
		fmt.Print("------------------          " + k)
		fmt.Print(" : ")
		fmt.Println(v)
	}
	return
}

//PrintResStatusCode 打印 响应 装态码
func (ros *requests) PrintResStatusCode() {
	fmt.Println("------------------  Res StatusCode ==> " + strconv.Itoa(ros.retHttpInfos.resStatusCode))
	fmt.Println("")
}
