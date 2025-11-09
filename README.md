
<p align="center"> 
  <h1> 欢迎使用gspider 蜘蛛 爬虫 采集 </h1>
</p>


<p align="center">快速采集网页 </p>
 
开始
===============

## 安装
```sh
$ go get -u github.com/ghostljj/gspider
```
```azure
python 有大名鼎鼎的requests
golang 有gspider 大致使用差不多
支持http代理，Socks5代理
支持HTTP/3 (QUIC) 协议
支持浏览器指纹模拟 (Surf 模式)
```

 
## 例子

```go
package main

import (
        "fmt"
        gs "github.com/ghostljj/gspider"
)

func main() {
	var strUrl string
	strUrl = "http://2022.ip138.com/ic.asp"
	//strUrl = "http://www.baidu.com"
	//strUrl = "http://www.google.com"

	req := gs.Session()
	//ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10809))
	//ss.SetSocks5Proxy("127.0.0.1:10808", "", "")

	res := req.Get(strUrl,
		gs.OptRefererUrl("http://www.baidu.com"),
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
```

为什么打印些无用的东西给我？<br/>
因为，这就是调试信息，仔细看会发现使用函数哈。<br/>
打印后慢慢磨控制台，会有惊喜<br/>

```go

可以Post, Get, PostJson,GetJson 等  有示例2 有空可以看看

Post 时注意，送给同学们url.QueryEscape 这个函数，用于参数编码，会有用的。Post json请忽略

还有就是可以获取图像Base64字符串，使用GetBase64Image
```

设置Cookies
```go
    SetCookies(strUrl, "NewKey1=NewValue1;NewKey2=NewValue==99=2;")
```

清空Cookies
```go
     ResetCookie()
```
获取Cookies
```go
    Cookies(strUrl)
```


题外话：获取 Cookie Json
可用于Chrome的 EditThisCookie 插件
当你知道某网站的Cookie时，使用这个可以生成能用EditThisCookie导入Cookie里面。例如一些已登录的网站。
```go
    gspider.GetCookieJson(strUrl, strCookie)
```

## 浏览器指纹（Surf）集成

如果你需要更强的浏览器指纹模拟（JA3/JA4、HTTP/2/3 指纹、头顺序等），可以启用 Surf 集成：

```go
// 设置浏览器+版本指纹后即自动启用 Surf 模式
// 方式一：在 Request 上设置默认档位（影响后续请求）
req.SetSurfBrowserProfile(gs.SurfBrowserChromeStable) // 浏览器+版本（枚举）
res := req.Get("https://example.com",
)
req.SetSurfHTTP3(true)            // 是否启用 HTTP/3（QUIC）指纹
req.SetSurfOS(gs.SurfOSRandomDesktop)
req.SetHTTP3(true)                // 是否启用 HTTP/3（QUIC）指纹
// 默认 Surf 模式为短连接；如需复用可设置：
req.SetSurfClose(false)           // 关闭强制短连接以保留 Keep-Alive（默认 true）
)
```

说明：
- 调用 `req.SetSurfBrowserProfile(...)` 后，请求将通过 Surf 的 `Std()` 客户端发送并保留指纹特性。
- mTLS/证书：`SetmTLSClient(...)`/`SetmTLSClientFile(...)` 仍可工作（会尝试将证书配置应用到底层 `*http.Transport`）。
- 连接策略：Surf 模式默认短连接（`SurfClose=true`），更可控；如需贴近真实浏览器的连接复用，使用 `req.SetSurfClose(false)`。

#### 系统与浏览器版本（枚举设置，稳妥）

- 系统枚举：`req.SetSurfOS(kind)` 支持 `SurfOSWindows`、`SurfOSAndroid`、`SurfOSIOS`、`SurfOSRandomDesktop`、`SurfOSRandomMobile` 等；`SurfOSMacOS`/`SurfOSLinux` 当前映射到桌面稳定画像（库未提供专用档位）。
 - 浏览器+版本枚举：通过 `req.SetSurfBrowserProfile(profile)` 统一设置浏览器与版本；支持以下完整档位（未知档位自动回退到稳定版）：
 - Chrome 家族：`SurfBrowserChromeStable`、`SurfBrowserChrome58`、`SurfBrowserChrome62`、`SurfBrowserChrome70`、`SurfBrowserChrome72`、`SurfBrowserChrome83`、`SurfBrowserChrome87`、`SurfBrowserChrome96`、`SurfBrowserChrome100`、`SurfBrowserChrome102`、`SurfBrowserChrome106`、`SurfBrowserChrome120`、`SurfBrowserChrome120PQ`、`SurfBrowserChrome142`
 - Edge（按 Chrome 家族处理）：`SurfBrowserEdgeStable`、`SurfBrowserEdge85`、`SurfBrowserEdge106`
 - Firefox 家族：`SurfBrowserFirefoxStable`、`SurfBrowserFirefox55`、`SurfBrowserFirefox56`、`SurfBrowserFirefox63`、`SurfBrowserFirefox65`、`SurfBrowserFirefox99`、`SurfBrowserFirefox102`、`SurfBrowserFirefox105`、`SurfBrowserFirefox120`、`SurfBrowserFirefox141`、`SurfBrowserFirefox144`、`SurfBrowserFirefoxPrivate144`
 - Tor（按 Firefox 家族处理）：`SurfBrowserTor`、`SurfBrowserTorPrivate`
 - iOS/Safari 家族：`SurfBrowserSafari`、`SurfBrowserIOS`、`SurfBrowserIOS11`、`SurfBrowserIOS12`、`SurfBrowserIOS13`、`SurfBrowserIOS14`
 - Randomized：`SurfBrowserRandomized`、`SurfBrowserRandomizedALPN`、`SurfBrowserRandomizedNoALPN`
（旧写法已移除，请使用 `req.SetSurfBrowserProfile(profile)` 与 `req.SetSurfOS(kind)`）

示例：
```go
// 启用 Surf，固定 Chrome 142 指纹，复用连接，开启 HTTP/3（QUIC）并使用对应 QUIC 指纹
req.SetSurfBrowserProfile(gs.SurfBrowserChrome142)
res := req.Get("https://example.com",
req.SetSurfHTTP3(true)  // HTTP/3 使用按浏览器选择的 QUIC 指纹
req.SetSurfClose(false) // 允许连接复用（HTTP/1.1/HTTP/2 有效）
req.SetHTTP3(true)      // HTTP/3 使用按浏览器选择的 QUIC 指纹
)

// 使用较旧的 Chrome 87 指纹，短连接，HTTP/2（禁用 H3）
req.SetSurfBrowserProfile(gs.SurfBrowserChrome87)
res2 := req.Get("https://example.org",
req.SetSurfHTTP3(false) // 使用 *http.Transport，支持 LocalIP 绑定与连接控制
req.SetSurfClose(true)  // 强制短连接
req.SetHTTP3(false)     // 使用 *http.Transport，支持 LocalIP 绑定与连接控制
)

// Firefox：指定版本 token 将回退到稳定版 Firefox 指纹
req.SetSurfBrowserProfile(gs.SurfBrowserFirefoxStable)
req.SetSurfHTTP3(true)
)
req.SetHTTP3(true)
)
```

注意：
- `SetSurfBrowserProfile` 会对 TLS 指纹（JA）和 QUIC 指纹（HTTP/3Settings）进行档位选择；不存在的档位会自动回退，不会编译失败。
- 当启用 `req.SetSurfHTTP3(true)` 时，连接管理由 Surf 内部栈处理；`SurfClose` 和 `LocalIP` 绑定只在 `*http.Transport`（HTTP/1.1/HTTP/2）下生效。

### 代理与本地绑定（Surf 模式）

Surf 模式已映射现有的代理和本地绑定选项，优先级如下：
- `Socks5Address`（支持 `Socks5User/Socks5Pass`）→ `HttpProxyInfo`（自动补全协议）→ `HttpProxyAuto`（从环境 `HTTPS_PROXY`/`HTTP_PROXY` 读取）。
- 本地 IP 绑定：设置 `req.LocalIP` 时会尽可能绑定到传输层的 `LocalAddr`（适用于 HTTP/1.1/HTTP/2）。

示例：
```go
req := gs.Session()
req.Socks5Address = "127.0.0.1:1080"   // 或设置 HttpProxyInfo = "http://127.0.0.1:8888"
req.Socks5User = "user"                // 可选
req.Socks5Pass = "pass"                // 可选
req.LocalIP = "192.168.1.100"          // 绑定出站本地 IP（非 HTTP/3 场景）

// 也可以在 req 上设置默认指纹：
req.SetSurfBrowserProfile(gs.SurfBrowserChromeStable)
res := req.Get("https://example.com",
)
req.SetSurfOS(gs.SurfOSAndroid)     // 系统（枚举）
req.SetSurfHTTP3(false)             // 如需本地绑定与短连接控制，建议非 HTTP/3
)
```

- `LocalIP` 绑定和 `req.SetSurfClose(...)` 的传输层行为仅在 `*http.Transport`（HTTP/1.1/HTTP/2）下生效；HTTP/3（QUIC）由 Surf 内部栈管理，连接关闭语义不同。


### User-Agent 与指纹统一策略
- 未启用 Surf（`req.SetSurfBrowserProfile(gs.SurfBrowserDisabled)`）：请求 UA 来自 `req.UserAgent`（默认为桌面浏览器 UA），也可用 `OptHeader` 自定义覆盖。
- 启用 Surf（设置了任意非 Disabled 档位）：请求 UA 由 Surf 的浏览器+系统指纹生成；`req.UserAgent` 将被忽略以避免与 TLS/ALPN 指纹不一致。
- 如确需在 Surf 模式下自定义 UA，请显式传入 `OptHeader(map[string]string{"User-Agent": "..."})`。但这会破坏与指纹的一致性，除非你非常确定需要这样做。
- 示例建议：启用 Surf 时不要设置 `req.UserAgent`，由档位自动生成一致的 UA 与指纹。

### 在 Surf 模式下获取 UA（不发请求）
- 通过 `req.GetSurfUserAgent()` 可直接获得与当前 Surf 档位和 OS 一致的 UA 字符串，用于第三方插件或日志。
- 示例：
  ```go
  req := gspider.Session()
  req.SetSurfBrowserProfile(gspider.SurfBrowserChrome142)
  req.SetSurfOS(gspider.SurfOSWindows)
  ua := req.GetSurfUserAgent()
  fmt.Println("Fingerprint UA:", ua)
  ```

### 获取"实际发送"的 UA（轻量探针）
- 若需要确认真实发送给服务器的 UA，可在请求完成后读取：`res.GetReqHeader().Get("User-Agent")`；推荐用 `HEAD` 进行轻量探针。

## HTTP/3 (QUIC) 支持

gspider 现在支持 HTTP/3 协议，可以在 Surf 模式和非 Surf 模式下使用。

### 非 Surf 模式下使用 HTTP/3

```go
req := gs.Session()
// 启用 HTTP/3
req.SetHTTP3(true)

res := req.Get("https://cloudflare-quic.com")
if res.GetErr() != nil {
    fmt.Println("Error:", res.GetErr().Error())
} else {
    fmt.Println("Status Code:", res.GetStatusCode())
    fmt.Println("Content Length:", len(res.GetContent()))
}
```

### Surf 模式下使用 HTTP/3（带浏览器指纹）

```go
req := gs.Session()
// 设置 Surf 浏览器指纹
req.SetSurfBrowserProfile(gs.SurfBrowserChrome142)
req.SetSurfOS(gs.SurfOSWindows)
// 启用 HTTP/3
req.SetHTTP3(true)

res := req.Get("https://www.google.com")
if res.GetErr() != nil {
    fmt.Println("Error:", res.GetErr().Error())
} else {
    fmt.Println("Status Code:", res.GetStatusCode())
    fmt.Println("User-Agent:", req.UserAgent)
}
```

### HTTP/3 注意事项

1. **协议回退**：如果服务器不支持 HTTP/3，请求可能会失败或自动回退到 HTTP/2 或 HTTP/1.1（取决于服务器配置）
2. **代理支持**：HTTP/3 目前不支持代理（这是 QUIC 协议的限制）
3. **性能优势**：HTTP/3 使用 UDP 协议，在高延迟或丢包网络环境下性能更好
4. **TLS 配置**：HTTP/3 会自动使用配置的 TLS 证书（如果有设置）

### 示例代码

完整示例请查看：[_examples/testHTTP3.go](_examples/testHTTP3.go)
