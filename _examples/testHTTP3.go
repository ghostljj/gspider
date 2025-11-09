package main

import (
	"fmt"

	gs "github.com/ghostljj/gspider"
)

func main() {
	// 测试 HTTP/3 支持的网站（Cloudflare 提供的 HTTP/3 测试站点）
	var strUrl string
	strUrl = "https://cloudflare-quic.com"
	// 或者使用 Google 的服务（支持 HTTP/3）
	// strUrl = "https://www.google.com"

	// ========================================
	// 示例 1: 非 Surf 模式下使用 HTTP/3
	// ========================================
	fmt.Println("========================================")
	fmt.Println("示例 1: 非 Surf 模式 + HTTP/3")
	fmt.Println("========================================")

	req := gs.Session()
	// 启用 HTTP/3（仅在非 Surf 模式下生效）
	req.SetHTTP3(true)

	// 可选：设置代理
	// req.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10808))

	res := req.Get(strUrl)
	if res.GetErr() != nil {
		fmt.Println("Error:", res.GetErr().Error())
	} else {
		fmt.Println("Status Code:", res.GetStatusCode())
		fmt.Println("Protocol:", res.GetResHeader().Get("alt-svc")) // HTTP/3 通常会在这里显示
		fmt.Println("Content Length:", len(res.GetContent()))
		fmt.Println()
	}

	// ========================================
	// 示例 2: Surf 模式下使用 HTTP/3
	// ========================================
	fmt.Println("========================================")
	fmt.Println("示例 2: Surf 模式 + HTTP/3 + 浏览器指纹")
	fmt.Println("========================================")

	req2 := gs.Session()
	// 设置 Surf 浏览器指纹
	req2.SetSurfBrowserProfile(gs.SurfBrowserChrome142)
	req2.SetSurfOS(gs.SurfOSWindows)
	// 在 Surf 模式下启用 HTTP/3
	req2.SetHTTP3(true)

	res2 := req2.Get(strUrl)
	if res2.GetErr() != nil {
		fmt.Println("Error:", res2.GetErr().Error())
	} else {
		fmt.Println("Status Code:", res2.GetStatusCode())
		fmt.Println("User-Agent:", req2.UserAgent)
		fmt.Println("Content Length:", len(res2.GetContent()))
		fmt.Println()
	}

	// ========================================
	// 示例 3: 对比 HTTP/1.1 vs HTTP/3
	// ========================================
	fmt.Println("========================================")
	fmt.Println("示例 3: HTTP/1.1 vs HTTP/3 性能对比")
	fmt.Println("========================================")

	// HTTP/1.1 请求
	req3 := gs.Session()
	req3.SetHTTP3(false) // 使用 HTTP/1.1
	res3 := req3.Get(strUrl)
	if res3.GetErr() != nil {
		fmt.Println("HTTP/1.1 Error:", res3.GetErr().Error())
	} else {
		fmt.Println("HTTP/1.1 - Status Code:", res3.GetStatusCode())
	}

	// HTTP/3 请求
	req4 := gs.Session()
	req4.SetHTTP3(true) // 使用 HTTP/3
	res4 := req4.Get(strUrl)
	if res4.GetErr() != nil {
		fmt.Println("HTTP/3 Error:", res4.GetErr().Error())
	} else {
		fmt.Println("HTTP/3 - Status Code:", res4.GetStatusCode())
	}

	fmt.Println()
	fmt.Println("测试完成！")
}
