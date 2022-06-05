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

//GetCookieJson 一个全局方法 可以获取cookie json  可用于chrome的 EditThisCookie插件
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

//GetCookiesJarMap 获取 cookieJar 的 map[string]string
func GetCookiesJarMap(cookieJar http.CookieJar, strUrl string) *map[string]string {
	mapCookies := make(map[string]string)
	if cookieJar == nil {
		return &mapCookies
	}
	URI, _ := url.Parse(strUrl)
	return GetCookiesMap(cookieJar.Cookies(URI))
}

//GetCookiesMap 获取Cook的map[string]string
func GetCookiesMap(cookies []*http.Cookie) *map[string]string {
	mapCookies := make(map[string]string)
	if cookies == nil {
		return &mapCookies
	}
	for _, cookie := range cookies {
		mapCookies[cookie.Name] = cookie.Value
	}
	return &mapCookies
}

//Cookies 获取Cookie
func (ros *requests) Cookies(strUrl string) string {
	if ros.cookieJar == nil {
		return ""
	}
	var str string
	mapCookieHost := GetCookiesJarMap(ros.cookieJar, strUrl)
	for k, v := range *mapCookieHost {
		str += k + "=" + v + ";"
	}
	return str
}

//CookiesAll 获取本url和主url的cookie
func (ros *requests) CookiesAll(strUrl string) string {
	if ros.cookieJar == nil {
		return ""
	}
	URI, _ := url.Parse(strUrl)
	var str string
	mapCookie := GetCookiesJarMap(ros.cookieJar, strUrl)
	mapCookieHost := GetCookiesJarMap(ros.cookieJar, URI.Scheme+"://"+URI.Host)
	for k, v := range *mapCookie {
		(*mapCookieHost)[k] = v
	}
	for k, v := range *mapCookieHost {
		str += k + "=" + v + ";"
	}
	return str
}

//ResetCookie 重置Cookie
func (ros *requests) ResetCookie() {
	ros.cookieJar, _ = cookiejar.New(nil)
}

//SetCookies 设置当前url Cookie
func (ros *requests) SetCookies(strUrl, strCookie string) {
	if ros.cookieJar == nil {
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
	ros.cookieJar.SetCookies(URI, addCookies) //这里设置的cookie 会自动合并到cookieJar
}

//SetCookiesAll 设置根url和当前url cookie
func (ros *requests) SetCookiesAll(strUrl, strCookie string) {

	URI, _ := url.Parse(strUrl)
	strHostUrl := URI.Scheme + "://" + URI.Host
	ros.SetCookies(strUrl, strCookie)
	ros.SetCookies(strHostUrl, strCookie)
}

//SetCookiesToUrl 把老Url的cookie 导入到新Url的cookie
func (ros *requests) SetCookiesToUrl(strUrlOld, strUrlNew string) {
	strCookieOld := ros.Cookies(strUrlOld)
	ros.SetCookies(strUrlNew, strCookieOld)
}
