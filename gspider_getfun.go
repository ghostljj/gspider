package gspider

import (
	"net/http"
	"net/url"
)

//获取 请求 头信息
func (s *Spider) GetReqHeader() http.Header {
	return s.reqHeader
}

//获取 响应 头信息
func (s *Spider) GetResHeader() http.Header {
	return s.resHeader
}

//获取 响应 后的Url
func (s *Spider) GetResUrl() string {
	return s.resUrl
}

//获取 响应 状态码
func (s *Spider) GetResStatusCode() int {
	return s.resStatusCode
}

//获取 cookieJar 的 map[string]strin
func (s *Spider) GetCookiesMap(strUrl string) map[string]string {

	URI, _ := url.Parse(strUrl)
	gCurCookies := s.cookieJar.Cookies(URI)
	mapCookies := make(map[string]string)
	cookieNum := len(gCurCookies)
	for i := 0; i < cookieNum; i++ {
		var curCk *http.Cookie = gCurCookies[i]
		mapCookies[curCk.Name] = curCk.Value
	}
	return mapCookies
}
