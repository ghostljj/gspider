
<p align="center"> 
  <h1> 欢迎使用gspider 蜘蛛 爬虫 采集</h1>
</p> 


<p align="center">快速采集网页</a></p>
 
开始
===============

## 安装
```sh
$ go get -u ...
```
```go
  "github.com/axgle/mahonia"  //解决编码用
  "github.com/saintfish/chardet" //自动获取编码用 
  "golang.org/x/net/proxy" //设置代理用

  "github.com/ghostljj/gspider"//爬虫包
```
此包需要使用到上面几个库，请自行go get -u .... <br/>
支持http(s)代理，Socks5代理 <br/> 




## 例子

```go

	var strUrl string
	strUrl = "http://2018.ip138.com/ic.asp"
	ss := gspider.NewSpider()
	{ //设置代理  / Socks5
		// ss.HttpProxyInfo = "http://127.0.0.1:1081" //设置代理

		// ss.Socks5Address = "127.0.0.1:7813" //设置代理Socks5
		// ss.Socks5User = "User"
		// ss.Socks5Pass = "pass"
	}
	{ //设置Cookie
		// ss.SetCookiesAll(strUrl, "NewKey1=NewValue1;NewKey2=NewValue==99=2;")
  }

	strContent, err := ss.Send("GET", strUrl, strUrl, "", nil)
  // 或者用这个	strContent, err := ss.Get(strUrl, strUrl, nil)
	if err != nil {
		fmt.Println("Error=" + err.Error())
	} else {

		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(strContent)
		ss.PrintReqHeader("")           //打印 请求 头信息
		ss.PrintResHeader("")           //打印 响应 头信息
		ss.PrintResSetCookie()          //打印 响应 头信息SetCookie
		ss.PrintReqUrl()                // 打印 请求 Url
		ss.PrintReqPostData()           // 打印 请求 PostData
		ss.PrintResUrl()                // 打印 响应 最后的Url
		ss.PrintCookies(ss.GetResUrl()) // 获取 响应 最后的Url 的 Cookie
		ss.PrintResStatusCode()         // 打印 响应 状态码

	}
```

为什么打印些无用的东西给我？<br/>
因为，这就是调试信息，仔细看会发现使用函数哈。<br/>
打印后慢慢磨控制台，会有惊喜<br/>

```go
值得注意的是我单独写出RefererUrl,我个人认为，很多网站模拟的时候。是需要看来源的。特别是高级爬虫的时候。麻烦点是麻烦点，安全稳妥。

可以Post 和 Get 或者Send  最后有个 nil  ，这个是map[string]string 请求头的修改，不修改就是nil
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
