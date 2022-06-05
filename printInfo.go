package gspider

import (
	"fmt"
	"strconv"
)

//PrintInfo
func (res *response) PrintInfo() {
	res.PrintReqHeader("")  // 打印 请求  信息
	res.PrintResHeader("")  // 打印 响应 头信息
	res.PrintResSetCookie() // 打印 响应 头信息SetCookie
	res.PrintReqUrl()       // 打印 请求 Url
	res.PrintReqPostData()  // 打印 请求 PostData
	res.PrintResUrl()       // 打印 响应 最后的Url
	res.PrintCookies()      // 获取 响应 最后的Url 的 SetCookie
	res.PrintStatusCode()   // 打印 响应 状态码
}

//PrintReqHeader 打印 请求 头信息 查看信息用
func (res *response) PrintReqHeader(key string) {
	if key == "" {
		fmt.Println("------------------  Req 请求 Header | GetReqHeader() map[string][]string")
		fmt.Println("------------------  使用 res.GetReqHeader() http.Header 例如 GetReqHeader().Get(\"User-Agent\") ")

		for k, v := range res.GetReqHeader() {
			fmt.Print("------------------          " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------------------------------------------")
	} else {
		fmt.Println("------------------  Req Header  ==> (" + res.GetReqHeader().Get(key) + ") End") // 例如 User-Agent
	}
	fmt.Println("")
}

//PrintReqPostData 打印 请求 Post内容
func (res *response) PrintReqPostData() {
	fmt.Println("------------------  Req PostData ==> ( " + res.GetReqPostData() + " )")
	fmt.Println("")
}

//PrintResHeader 打印 响应 头信息
func (res *response) PrintResHeader(key string) {
	if key == "" {
		fmt.Println("------------------  Res 响应 Header")
		fmt.Println("------------------  使用 res.GetResHeader() http.Header 例如 GetResHeader().Get(\"Content-Encoding\") ")
		for k, v := range res.GetResHeader() {
			fmt.Print("------------------          " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------------------------------------------")
	} else {
		fmt.Println("------------------  Res Header  ==> (" + res.GetResHeader().Get(key) + ") End") // 例如 Content-Encoding
	}
	fmt.Println("")
}

//PrintResSetCookie
func (res *response) PrintResSetCookie() {
	fmt.Println("------------------  S Res 响应 Set-Cookie ")
	fmt.Println("------------------  使用 res.GetResCookies() []*http.Cookie ")
	for _, itemCookie := range res.GetResCookies() {
		fmt.Println("------------------          ", itemCookie)
	}
	fmt.Println("------------------------------------------------------")
	fmt.Println()
}

//PrintReqUrl 打印 请求 URL
func (res *response) PrintReqUrl() {
	fmt.Println("------------------  Req Url 请求 URL ==> (" + res.GetReqUrl() + ") End")
	fmt.Println("")
}

//PrintResUrl 打印最后 响应 URL
func (res *response) PrintResUrl() {
	fmt.Println("------------------  Res Url 最后 响应 URL ==> (" + res.GetResUrl() + ") End") // 例如 Content-Encoding
	fmt.Println("")
}

//PrintResStatusCode 打印 响应 装态码
func (res *response) PrintStatusCode() {
	fmt.Println("------------------  Res StatusCode ==> " + strconv.Itoa(res.GetStatusCode()))
	fmt.Println("")
}

//PrintCookies 打印CookieJar
func (res *response) PrintCookies() {
	fmt.Println("------------------  S SetCookies ==> From(" + res.resUrl + ")")
	fmt.Println("------------------  使用 res.GetResCookies() map[string]string")
	defer func() {
		fmt.Println("------------------------------------------------------")
		fmt.Println("")
	}()
	if res.resCookies == nil {
		return
	}
	mapCookieHost := GetCookiesMap(res.resCookies)
	for k, v := range *mapCookieHost {
		fmt.Print("------------------          " + k)
		fmt.Print(" : ")
		fmt.Println(v)
	}
	return
}

//PrintCookies 打印CookieJar
func (req *requests) PrintCookieJar(strUrl string) {
	fmt.Println("------------------  S 全局CookieJar  ==> From(" + strUrl + ")")
	fmt.Println("------------------  使用 GetCookiesJarMap(http.CookieJar,strUrl string) map[string]string")
	defer func() {
		fmt.Println("------------------------------------------------------")
		fmt.Println("")
	}()
	if req.cookieJar == nil {
		return
	}
	mapCookieHost := GetCookiesJarMap(req.cookieJar, strUrl)
	for k, v := range *mapCookieHost {
		fmt.Print("------------------          " + k)
		fmt.Print(" : ")
		fmt.Println(v)
	}
	return
}
