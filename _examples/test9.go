package main

import (
	"bytes"
	"fmt"
	gs "github.com/ghostljj/gspider"
	"io"
	"log"
	"mime/multipart"
	"os"
)

func main() {
	url := "https://httpbin.org/post"
	filePath := "c:\\test.pdf"

	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)
	err := multipartWriter.WriteField("content", "12334")
	if err != nil {
		log.Fatal(err)
		return
	}

	// --- 步骤 1: 打开要上传的文件 ---
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("无法打开文件: %s, 错误: %v", filePath, err)
	}
	defer file.Close()

	// --- 步骤 2: 创建一个文件表单字段 ---
	// 第一个参数是表单字段的 "name" (即 key)，第二个参数是文件名
	fileWriter, err := multipartWriter.CreateFormFile("file", "test.pdf")
	if err != nil {
		log.Fatalf("无法创建表单文件字段: %v", err)
	}

	// --- 步骤 3: 将文件内容拷贝到表单字段中 ---
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalf("无法拷贝文件内容: %v", err)
	}
	multipartWriter.Close()//重点

	fmt.Println("文件大小", len(requestBody.String()))

	req := gs.Session()
	//ss.SetHttpProxy(fmt.Sprintf("http://%s:%d", "127.0.0.1", 10809))
	//ss.SetSocks5Proxy("127.0.0.1:10808", "", "")
	req.OnUploaded(func(uploaded *int64, req *gs.Request) {
		fmt.Println("已上传", *uploaded)
	})

	fmt.Println(multipartWriter.FormDataContentType())

	res := req.PostBig(url,
		requestBody.Bytes(),
		gs.OptRefererUrl(url),
		gs.OptHeader(map[string]string{"Content-Type": multipartWriter.FormDataContentType()}),//重点
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
