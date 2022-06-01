package gspider

import (
	"fmt"
	"strconv"
)

func (s *Spider) PrintInfo() {
	s.PrintReqHeader("")          // 打印 请求  信息
	s.PrintResHeader("")          // 打印 响应 头信息
	s.PrintResSetCookie()         // 打印 响应 头信息SetCookie
	s.PrintReqUrl()               // 打印 请求 Url
	s.PrintReqPostData()          // 打印 请求 PostData
	s.PrintResUrl()               // 打印 响应 最后的Url
	s.PrintCookies(s.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
	s.PrintResStatusCode()        // 打印 响应 状态码
}

//打印 请求 头信息 查看信息用
func (s *Spider) PrintReqHeader(key string) {
	if key == "" {
		fmt.Println("------------------  Req 请求 Header | GetReqHeader() map[string][]string")
		fmt.Println("------------------  使用 GetReqHeader() http.Header 可以使用 GetReqHeader().Get(\"User-Agent\") ")

		for k, v := range s.GetReqHeader() {
			fmt.Print("------------------          " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------------------------------------------")
	} else {
		fmt.Println("------------------  Req Header  ==> (" + s.GetReqHeader().Get(key) + ") End") // 例如 User-Agent
	}
	fmt.Println("")
}

//打印 请求 Post内容
func (s *Spider) PrintReqPostData() {
	fmt.Println("------------------  Req PostData ==> ( " + s.reqPostData + " )")
	fmt.Println("")
}

//打印 响应 头信息
func (s *Spider) PrintResHeader(key string) {
	if key == "" {
		fmt.Println("------------------  Res 响应 Header")
		fmt.Println("------------------  使用 GetResHeader() http.Header 可以使用 GetResHeader().Get(\"Content-Encoding\") ")
		for k, v := range s.GetResHeader() {
			fmt.Print("------------------          " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------------------------------------------")
	} else {
		fmt.Println("------------------  Res Header  ==> (" + s.GetResHeader().Get(key) + ") End") // 例如 Content-Encoding
	}
	fmt.Println("")
}

func (s *Spider) PrintResSetCookie() {
	fmt.Println("------------------  S Res 响应 Set-Cookie ")
	fmt.Println("------------------  使用 GetResCookies() []*http.Cookie ")
	for _, itemCookie := range s.GetResCookies() {
		fmt.Println("------------------          ", itemCookie)
	}
	fmt.Println("------------------------------------------------------")
	fmt.Println()
}

//打印 请求 URL
func (s *Spider) PrintReqUrl() {
	fmt.Println("------------------  Req Url 请求 URL ==> (" + s.reqUrl + ") End")
	fmt.Println("")
}

//打印最后 响应 URL
func (s *Spider) PrintResUrl() {
	fmt.Println("------------------  Res Url 最后 响应 URL ==> (" + s.resUrl + ") End") // 例如 Content-Encoding
	fmt.Println("")
}

//打印CookieJar
func (s *Spider) PrintCookies(strUrl string) {
	fmt.Println("------------------  S CookieJar  ==> From(" + strUrl + ")")
	fmt.Println("------------------  使用 GetCookiesMap(strUrl string) map[string]string")
	defer func() {
		fmt.Println("------------------------------------------------------")
		fmt.Println("")
	}()
	if s.cookieJar == nil {
		return
	}
	mapCookieHost := s.GetCookiesMap(strUrl)
	for k, v := range mapCookieHost {
		fmt.Print("------------------          " + k)
		fmt.Print(" : ")
		fmt.Println(v)
	}
	return
}

//打印 响应 装态码
func (s *Spider) PrintResStatusCode() {
	fmt.Println("------------------  Res StatusCode ==> " + strconv.Itoa(s.resStatusCode))
	fmt.Println("")
}
