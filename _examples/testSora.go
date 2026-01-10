package main

import (
	"fmt"

	gs "github.com/ghostljj/gspider"
)

func main() {

	var strUrl string
	strUrl = "https://sora.chatgpt.com/p/s_68f23567d1008191a74ea632405fa6d7"
	//strUrl = "http://www.baidu.com"
	//strUrl = "http://www.google.com"

	req := gs.Session()
	req.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10808))
	// Surf 模式下不要设置 req.UserAgent，UA 将由指纹档位生成
	req.SetSurfBrowserProfile(gs.SurfBrowserChrome143)
	req.SetSurfOS(gs.SurfOSWindows)
	// 直接读取与指纹一致的 UA（不发请求），用于插件传参或日志
	fmt.Println("Fingerprint UA:", req.UserAgent)

	res := req.Get(strUrl,
		gs.OptRefererUrl(strUrl),
		gs.OptCookie("aa=11;bb=22"),
		gs.OptHeader(map[string]string{"h1": "v1", "h2": "v2"}),
	)
	res.Encode = "utf-8"
	if res.GetErr() != nil {
		fmt.Println("Error=" + res.GetErr().Error())
	} else {

		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(res.GetContent())
		// 验证实际发送的 UA 与指纹一致性
		fmt.Println("Sent UA:", res.GetReqHeader().Get("User-Agent"))
		res.PrintReqHeader("") //打印 请求 头信息
		res.GetReqHeader()
		res.PrintReqPostData()            // 打印 请求 PostData
		res.PrintResHeader("")            //打印 响应 头信息
		res.PrintResSetCookie()           //打印 响应 头信息SetCookie
		res.PrintResUrl()                 // 打印 响应 最后的Url
		res.PrintCookies(res.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
		res.PrintStatusCode()             // 打印 响应 状态码

	}
}
