package gspider

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//一个全局方法 可以获取cookie json  可用于chrome的 EditThisCookie插件
func GetCookieJson(strUrl, strCookie string) string {
	URI, _ := url.Parse(strUrl)
	var mapCookies []map[string]interface{}
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
		mapItem := make(map[string]interface{})
		mapItem["domain"] = URI.Host
		mapItem["name"] = attr
		mapItem["path"] = URI.Path
		if len(URI.Path) <= 0 {
			mapItem["path"] = "/"
		}
		mapItem["value"] = val
		mapItem["expirationDate"] = time.Now().Add(time.Hour * +2).Unix()
		mapItem["id"] = strconv.Itoa(i + 1)
		mapCookies = append(mapCookies, mapItem)
	}
	byteCookies, _ := json.Marshal(mapCookies)
	jsonCookies := string(byteCookies)
	return jsonCookies
}

//获取Cookie
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

// 获取本url和主url的cookie
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

//重置Cookie
func (s *Spider) ResetCookie() {
	s.cookieJar, _ = cookiejar.New(nil)
}

func (s *Spider) SetCookiesAll(strUrl, strCookie string) {

	URI, _ := url.Parse(strUrl)
	strHostUrl := URI.Scheme + "://" + URI.Host
	s.SetCookies(strUrl, strCookie)
	s.SetCookies(strHostUrl, strCookie)
}

//设置Cookie
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

// 把老Url的cookie 导入到新Url的cookie
func (s *Spider) SetCookiesToUrl(strUrlOld, strUrlNew string) {
	strCookieOld := s.Cookies(strUrlOld)
	s.SetCookies(strUrlNew, strCookieOld)
}
