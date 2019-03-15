
<p align="center"> 
  <h1> 欢迎使用gspider 蜘蛛 爬虫 采集</h1>
</p> 


<p align="center">快速采集网页</a></p>
 
开始
===============

## 安装

To start using GJSON, install Go and run `go get`:

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

		fmt.Println(strContent)
		ss.PrintReqHeader("")   //打印 请求 头信息
		ss.PrintReqPostData()   // 打印 请求 PostData
		ss.PrintResHeader("")   //打印 响应 头信息
		ss.PrintResSetCookie()  //打印 响应 头信息SetCookie
		ss.PrintResUrl()        // 打印 响应 最后的Url
		ss.PrintCookies(strUrl) // 获取此Url的Cookie
		ss.PrintResStatusCode() // 打印 响应 状态码
	}
```

为什么打印些无用的东西给我？<br/>
因为，这就是调试信息，仔细看会发现使用函数哈。<br/>
打印后慢慢磨控制台，会有惊喜<br/>


