package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

func main() {
	url := "https://httpbin.org/post"

	req := gs.Session()
	//ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10809))
	//ss.SetSocks5Proxy("127.0.0.1:10808", "", "")
	req.OnUploaded(func(uploaded *int64, req *gs.Request) {
		fmt.Println("已上传", *uploaded)
	})

	res := req.PostJson(url,
		`{"a":1,"b":2}`,
		gs.OptRefererUrl(url),
		gs.OptCookie("aa=11;bb=22"),
		gs.OptHeader(map[string]string{"h1": "v1", "h2": "v2"}),
	)
	if res.GetErr() != nil {
		fmt.Println("Error=" + res.GetErr().Error())
	} else {
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println(res.GetContent())

	}
}
