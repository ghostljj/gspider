
<p align="center"> 
  <h1> 欢迎使用gspider 蜘蛛 爬虫 采集</h1>
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
	req.RefererUrl = "http://www.baidu.com"
	req.Cookie = "aa=11;bb=22"
	req.Header = map[string]string{"h1": "v1", "h2": "v2"}
	//ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10809))
	//ss.SetSocks5Proxy("127.0.0.1:10808", "", "")

	res := req.Get(strUrl)
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
