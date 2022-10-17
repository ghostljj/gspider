package main

import (
	"fmt"
	gs "github.com/ghostljj/gspider"
)

var msg_chan = make(chan struct{}, 2)

func main() {

	print("192.168.32.101")
	//print("192.168.32.102")
	for {
		_, ok := <-msg_chan
		if !ok {
			break
		}
	}
}

func print(ip string) {
	req := gs.Session()
	//req.LocalIP = ip
	req.SetSocks5Proxy("127.0.0.1:10808", "", "")
	//req.SetHttpProxy("http://127.0.0.1:10809")
	//res := req.Get("https://2022.ip138.com/")
	res := req.Get("https://ifconfig.me/")
	//res := req.Get("https://www.baidu.com")
	fmt.Println(res.GetContent())
	fmt.Println()
}
