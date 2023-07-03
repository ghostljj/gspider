package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

func main() {

	req := gs.Session()
	req.SetCookies("https://chat-ws.baidu.com/lg/api/use_stream", "BDUSS=xxx-M0lpSnBYfjg5dU1JTDZHS2U3S2o2QkJicXVvOW9NY094R21SYkJrSVFBQUFBJCQAAAAAAAAAAAEAAABQ9nIAZ2hvc3RsamoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAKa4iGSmuIhkam")

	req.OnContent(func(byteItem []byte) {
		fmt.Println(string(byteItem))
	})
	res := req.Get("https://chat-ws.baidu.com/lg/api/use_stream?body=%7B%22app_id%22%3A%22b730178a5bc045daa0105dc9d2dd9f3e%22%2C%22input%22%3A%22%E6%A2%A6%E5%88%B0%E9%9D%92%E8%9B%99%22%7D",
		gs.OptHeader(map[string]string{"Content-Type": "text/event-stream"}))

	fmt.Println("---------------------------")
	fmt.Println(res.GetContent())
	fmt.Println()
}
