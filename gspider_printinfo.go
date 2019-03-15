package gspider

import (
	"fmt"
	"strconv"
)

//PrintReqHeader 打印 请求 头信息 查看信息用
func (s *Spider) PrintReqHeader(key string) {
	if key == "" {
		fmt.Println("------------------  S Req 请求 Header | GetReqHeader() map[string][]string")
		fmt.Println("------------------  使用 GetReqHeader() map[string][]string ")

		for k, v := range s.GetReqHeader() {
			fmt.Print("        " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------  E Req Header")
	} else {
		fmt.Println("Req Header  ==> (" + s.GetReqHeader().Get(key) + ") End") // 例如 User-Agent
	}
	fmt.Println("")
}

//PrintReqPostData 打印 请求 Post内容
func (s *Spider) PrintReqPostData() {
	fmt.Println("------------------ Req PostData ==> ( " + s.reqPostData + " )")
	fmt.Println("")
}

//PrintResHeader 打印 响应 头信息
func (s *Spider) PrintResHeader(key string) {
	if key == "" {
		fmt.Println("------------------  S Res 响应 Header")
		fmt.Println("------------------  使用 GetResHeader() map[string][]string ")
		for k, v := range s.GetResHeader() {
			fmt.Print("        " + k)
			fmt.Print(" : ")
			fmt.Println(v)
		}
		fmt.Println("------------------  E Res Header")
	} else {
		fmt.Println("Res Header  ==> (" + s.GetResHeader().Get(key) + ") End") // 例如 Content-Encoding
	}
	fmt.Println("")
}

//PrintResSetCookie 打印 响应 头的 Set-Cookie
func (s *Spider) PrintResSetCookie() {
	fmt.Println("------------------  S Res 响应 Set-Cookie ")
	fmt.Println("------------------  使用 GetResHeader()[\"Set-Cookie\"]  []string ")
	for _, itemCookie := range s.GetResHeader()["Set-Cookie"] {
		fmt.Println("        " + itemCookie)
	}
	fmt.Println("------------------  E Res 响应 Set-Cookie")
	fmt.Println()
}

//PrintResUrl 打印最后 响应 URL
func (s *Spider) PrintResUrl() {
	fmt.Println("------------------  Res Url 最后 响应 URL ==> (" + s.resUrl + ") End") // 例如 Content-Encoding
	fmt.Println("")
}

//PrintCookies 打印CookieJar
func (s *Spider) PrintCookies(strUrl string) {
	fmt.Println("------------------  S CookieJar  ==> From(" + strUrl + ")")
	fmt.Println("------------------  使用 GetCookiesMap(strUrl string) map[string]string")
	defer func() {
		fmt.Println("------------------  E CookieJar")
		fmt.Println("")
	}()
	if s.cookieJar == nil {
		return
	}
	mapCookieHost := s.GetCookiesMap(strUrl)
	for k, v := range mapCookieHost {
		fmt.Print("        " + k)
		fmt.Print(" : ")
		fmt.Println(v)
	}
	return
}

//PrintResStatusCode 打印 响应 装态码
func (s *Spider) PrintResStatusCode() {
	fmt.Println("------------------  Res StatusCode ==> " + strconv.Itoa(s.resStatusCode))
	fmt.Println("")
}
