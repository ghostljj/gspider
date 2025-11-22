package gspider

import (
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetCookieJson 一个全局方法 可以获取cookie json  可用于chrome的 EditThisCookie 插件
//
// strUrl: 为设置cookie的url
//
// strCookie: 为浏览器cookie字段列，例如 aaa=111;bbb=222;ccc=333
//
// 返回Json 字符串，可以使用 EditThisCookie 导入到浏览器。只要记录这段数据，日后未登录的用户马上变成登录，不过需要注意超时问题
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

// GetCookiesMap 获取Cook的map[string]string
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

// LoadCaFile 加载ca文件
func LoadCaFile(caFile string) *x509.CertPool {
	byteCa, err := os.ReadFile(caFile)
	if err != nil {
		Log.Fatal("loadCaFile: ", err)
		return nil
	}
	return LoadCa(byteCa)
}

// LoadCa  加载ca字节
func LoadCa(byteCa []byte) *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(byteCa)
	return pool
}
