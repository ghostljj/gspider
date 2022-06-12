package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

func main() {
	req := gs.Session()

	Headers := make(map[string]string)
	Headers["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
	Headers["accept-encoding"] = "gzip, deflate, br"
	Headers["accept-language"] = "zh-CN,zh;q=0.9"
	Headers["sec-ch-ua"] = `" Not A;Brand";v="99", "Chromium";v="102", "Google Chrome";v="102"`
	Headers["sec-ch-ua-mobile"] = "?0"
	Headers["sec-ch-ua-platform"] = `"Windows"`
	Headers["sec-fetch-dest"] = `document`
	Headers["sec-fetch-mode"] = `navigate`
	Headers["sec-fetch-site"] = `same-origin`
	Headers["sec-fetch-user"] = `?1`
	Headers["pragma"] = `no-cache`
	Headers["cache-control"] = `no-cache`
	Headers["upgrade-insecure-requests"] = `1`
	Headers["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36 Edg/102.0.1245.39"
	Headers["X-Requested-With"] = "XMLHttpRequest"
	req.Header = Headers
	req.SetCookiesAll("https://www.baidu.com", "_vid_t=VR0fN+BV6bokqkm+baDsGdRHRT16LBbhu8gJ0jqfxqEvxqW7rWQByM39SewCdEmj8tdWlxtB79z9iQ==; https_waf_cookie=d43c2b78-1139-4f34acf89fbc0c6673d4ee0d594d787d70fb; acw_tc=ac11000116550135235237919e011b3689e9004a956d8a8416e8d6a90a6c77")
	res := req.Get("https://www.baidu.com")
	//res := req.Get("https://www.baidu.com")
	fmt.Println(res.GetContent())
}
