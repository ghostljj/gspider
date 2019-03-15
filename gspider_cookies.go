package gspider

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// Cookies 获取Cookie
// strUrl
func (s *Spider) Cookies(strUrl string) string {
	if s.cookieJar == nil {
		return ""
	}
	var str string
	mapCookieHost := s.GetCookiesMap(strUrl)
	for k, v := range mapCookieHost {
		str += k + "=" + v + ";"
	}
	return str
}

// CookiesAll 获取本url和主url的cookie
// strUrl
func (s *Spider) CookiesAll(strUrl string) string {
	if s.cookieJar == nil {
		return ""
	}
	URI, _ := url.Parse(strUrl)
	var str string
	mapCookie := s.GetCookiesMap(strUrl)
	mapCookieHost := s.GetCookiesMap(URI.Scheme + "://" + URI.Host)
	for k, v := range mapCookie {
		mapCookieHost[k] = v
	}
	for k, v := range mapCookieHost {
		str += k + "=" + v + ";"
	}
	return str
}

// ResetCookie 重置Cookie
func (s *Spider) ResetCookie() {
	s.cookieJar, _ = cookiejar.New(nil)
}

func (s *Spider) SetCookiesAll(strUrl, strCookie string) {

	URI, _ := url.Parse(strUrl)
	strHostUrl := URI.Scheme + "://" + URI.Host
	s.SetCookies(strUrl, strCookie)
	s.SetCookies(strHostUrl, strCookie)
}

// SetCookies 设置Cookie
// strUrl strCookie
func (s *Spider) SetCookies(strUrl, strCookie string) {
	if s.cookieJar == nil {
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
	s.cookieJar.SetCookies(URI, addCookies) //这里设置的cookie 会自动合并到cookieJar
}

// SetCookiesToUrl 把老Url的cookie 导入到新Url的cookie
// strUrlOld strUrlNew
func (s *Spider) SetCookiesToUrl(strUrlOld, strUrlNew string) {
	strCookieOld := s.Cookies(strUrlOld)
	s.SetCookies(strUrlNew, strCookieOld)
}
