package gspider_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gs "github.com/ghostljj/gspider"
)

//go test -run Cancel -v

// 测试：使用单一分组 CancelGroup 仅取消该分组的请求，其他未分组请求不受影响
func Test_GroupCancel_OnlyGroupCanceled(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// 模拟较慢响应
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	req := gs.Session()

	// 启动两类并发：同组的2个，未分组的1个
	resCh := make(chan *gs.Response, 3)
	// 分组请求
	for i := 0; i < 2; i++ {
		go func() {
			res := req.Get(server.URL+"/slow", gs.OptCancelGroup("G1"))
			resCh <- res
		}()
	}
	// 未分组请求
	go func() {
		res := req.Get(server.URL + "/slow")
		resCh <- res
	}()

	// 等待请求启动后，取消指定分组
	time.Sleep(500 * time.Millisecond)
	req.CancelGroup("G1")

	canceled := 0
	succeeded := 0
	for i := 0; i < 3; i++ {
		res := <-resCh
		if res == nil {
			t.Fatal("expected a Response, got nil")
		}
		if res.GetErr() != nil {
			if strings.Contains(res.GetErr().Error(), "cancel group") || strings.Contains(res.GetErr().Error(), "context canceled") {
				canceled++
			} else {
				canceled++
			}
		} else {
			succeeded++
		}
	}
	if canceled != 2 || succeeded != 1 {
		t.Errorf("expected 2 canceled in group and 1 succeeded (others), got canceled=%d succeeded=%d", canceled, succeeded)
	}
}

// 测试：CancelGroupAll 取消所有分组，不影响未分组外已完成的请求（此处全部并发中）
func Test_GroupCancel_AllGroupsCanceled(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// 模拟较慢响应
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	req := gs.Session()

	resCh := make(chan *gs.Response, 4)
	// 组 G1 两个
	for i := 0; i < 2; i++ {
		go func() {
			res := req.Get(server.URL+"/slow", gs.OptCancelGroup("G1"))
			resCh <- res
		}()
	}
	// 组 G2 一个
	go func() {
		res := req.Get(server.URL+"/slow", gs.OptCancelGroup("G2"))
		resCh <- res
	}()
	// 未分组一个
	go func() {
		res := req.Get(server.URL + "/slow")
		resCh <- res
	}()

	// 取消所有分组
	time.Sleep(500 * time.Millisecond)
	req.CancelGroupAll()

	canceled := 0
	succeeded := 0
	for i := 0; i < 4; i++ {
		res := <-resCh
		if res == nil {
			t.Fatal("expected a Response, got nil")
		}
		if res.GetErr() != nil {
			if strings.Contains(res.GetErr().Error(), "cancel group all") || strings.Contains(res.GetErr().Error(), "cancel group") || strings.Contains(res.GetErr().Error(), "context canceled") {
				canceled++
			} else {
				canceled++
			}
		} else {
			succeeded++
		}
	}
	if canceled != 3 || succeeded != 1 {
		t.Errorf("expected 3 canceled in groups and 1 succeeded (ungrouped), got canceled=%d succeeded=%d", canceled, succeeded)
	}
}
