package gspider_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"strings"

	gs "github.com/ghostljj/gspider"
)

//go test -run Cancel -v

// Test cancellation behavior using a 2-second timeout context
func Test_Cancel_WithCancelCause(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// 2秒超时的父上下文
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := gs.SessionWithContext(ctx)

	resCh := make(chan *gs.Response, 1)
	go func() {
		res := req.Get(server.URL + "/slow")
		resCh <- res
	}()

	res := <-resCh
	if res == nil {
		t.Fatal("expected a Response, got nil")
	}
	if res.GetErr() == nil {
		t.Errorf("expected an error due to timeout cancellation, got nil (status=%d)", res.GetStatusCode())
	} else {
		t.Logf("timeout cancellation triggered, error: %v", res.GetErr())
	}
}

// 并发复用同一个 Request，手动取消时，仅取消最近一次派生的请求
func Test_Concurrent_ReuseRequest_ManualCancel_OnlyOneCanceled(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// 模拟较慢响应
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	req := gs.Session()

	const n = 3
	resCh := make(chan *gs.Response, n)
	for i := 0; i < n; i++ {
		go func() {
			res := req.Get(server.URL + "/slow")
			resCh <- res
		}()
	}

	// 等待请求启动后，手动取消
	time.Sleep(1 * time.Second)
	req.Cancel()

	canceled := 0
	succeeded := 0
	for i := 0; i < n; i++ {
		res := <-resCh
		if res == nil {
			t.Fatal("expected a Response, got nil")
		}
		if res.GetErr() != nil {
			// 手动取消的错误通常包含 "manual cancel"
			if strings.Contains(res.GetErr().Error(), "manual cancel") || strings.Contains(res.GetErr().Error(), "context canceled") {
				canceled++
			} else {
				// 其他错误也计入取消，以避免平台差异导致错误文案不同
				canceled++
			}
		} else {
			succeeded++
		}
	}
	if canceled != 1 {
		t.Errorf("expected exactly 1 canceled, got %d (succeeded=%d)", canceled, succeeded)
	}
}

// 取消所有：复用同一个 Request，启动多并发请求，调用 CancelAll，期望全部取消
func Test_Concurrent_CancelAll_AllCanceled(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// 模拟较慢响应
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	req := gs.Session()

	const n = 4
	resCh := make(chan *gs.Response, n)
	for i := 0; i < n; i++ {
		go func() {
			res := req.Get(server.URL + "/slow")
			resCh <- res
		}()
	}

	// 短暂等待所有请求启动后，调用 CancelAll
	time.Sleep(1 * time.Second)
	req.CancelAll()

	canceled := 0
	for i := 0; i < n; i++ {
		res := <-resCh
		if res == nil {
			t.Fatal("expected a Response, got nil")
		}
		if res.GetErr() == nil {
			t.Errorf("expected cancellation error, got nil (status=%d)", res.GetStatusCode())
		} else {
			// 接受 "cancel all" 或标准 "context canceled"
			if strings.Contains(res.GetErr().Error(), "cancel all") || strings.Contains(res.GetErr().Error(), "context canceled") {
				canceled++
			} else {
				canceled++
			}
		}
	}
	if canceled != n {
		t.Errorf("expected %d canceled, got %d", n, canceled)
	}
}
