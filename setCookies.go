package gspider

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

//GetCookiesJarMap
//获取 cookieJar 的 map[string]string
func (req *Request) GetCookiesJarMap(strUrl string) *map[string]string {
	mapCookies := make(map[string]string)
	if req.cookieJar == nil {
		return &mapCookies
	}
	URI, _ := url.Parse(strUrl)
	return GetCookiesMap(req.cookieJar.Cookies(URI))
}

//Cookies 获取Cookie
func (req *Request) Cookies(strUrl string) string {
	if req.cookieJar == nil {
		return ""
	}
	var str string
	mapCookieHost := req.GetCookiesJarMap(strUrl)
	for k, v := range *mapCookieHost {
		str += k + "=" + v + ";"
	}
	return str
}

//CookiesAll 获取本url和主url的cookie
func (req *Request) CookiesAll(strUrl string) string {
	if req.cookieJar == nil {
		return ""
	}
	URI, _ := url.Parse(strUrl)
	var str string
	mapCookie := req.GetCookiesJarMap(strUrl)
	mapCookieHost := req.GetCookiesJarMap(URI.Scheme + "://" + URI.Host)
	for k, v := range *mapCookie {
		(*mapCookieHost)[k] = v
	}
	for k, v := range *mapCookieHost {
		str += k + "=" + v + ";"
	}
	return str
}

//ResetCookie 重置Cookie
func (req *Request) ResetCookie() {
	req.cookieJar, _ = cookiejar.New(nil)
}

//SetCookies 设置当前url Cookie
func (req *Request) SetCookies(strUrl, strCookie string) {
	if req.cookieJar == nil {
		return
	}
	URI, _ := url.Parse(strUrl)
	// HostURI, _ := url.Parse(URI.Scheme + "://" + URI.Host)
	var addCookies []*http.Cookie
	parts := strings.Split(strings.TrimSpace(strCookie), ";")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimSpace(parts[i])
		if len(parts[i]) == 0 {
			continue
		}
		attr, val := parts[i], ""
		if j := strings.Index(attr, "="); j >= 0 {
			attr, val = attr[:j], attr[j+1:]
		}
		cookieItem := &http.Cookie{
			Name:  attr,
			Value: val,
		}
		addCookies = append(addCookies, cookieItem)
	}
	req.cookieJar.SetCookies(URI, addCookies) //这里设置的cookie 会自动合并到cookieJar
}

//SetCookiesAll 设置根url和当前url cookie
func (req *Request) SetCookiesAll(strUrl, strCookie string) {

	URI, _ := url.Parse(strUrl)
	strHostUrl := URI.Scheme + "://" + URI.Host
	req.SetCookies(strUrl, strCookie)
	req.SetCookies(strHostUrl, strCookie)
}

//SetCookiesToUrl 把老Url的cookie 导入到新Url的cookie
func (req *Request) SetCookiesToUrl(strUrlOld, strUrlNew string) {
	strCookieOld := req.Cookies(strUrlOld)
	req.SetCookies(strUrlNew, strCookieOld)
}
