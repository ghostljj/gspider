package gspider

import (
	"context"
	"io"
	"sync"
	"time"
)

type UploadedProgressReader struct {
	Reader     io.Reader
	Uploaded   int64
	Total      int64
	LastTime   time.Time
	chUploaded chan *int64
	closed     bool            // 标记信道是否已关闭
	mu         sync.Mutex      // 保证并发安全（防止多协程同时操作 closed 标志）
	ctx        context.Context // 新增：监听取消信号
}

func (pr *UploadedProgressReader) Read(p []byte) (n int, err error) {
	// 优先检查 ctx 是否被取消，若取消则立即返回错误
	select {
	case <-pr.ctx.Done():
		return 0, pr.ctx.Err() // 返回 context canceled 错误
	default:
	}
	// 1. 先读取原始数据（核心逻辑不变）
	n, err = pr.Reader.Read(p)
	pr.Uploaded += int64(n)

	// 2. 加锁判断信道是否已关闭（防止 Close 与 Read 并发执行）
	pr.mu.Lock()
	isClosed := pr.closed
	pr.mu.Unlock()

	// 3. 若已关闭，直接返回，不发送数据
	if isClosed || pr.chUploaded == nil {
		return n, err
	}

	// 4. 按原逻辑判断是否需要发送进度（非阻塞发送）
	if time.Since(pr.LastTime).Milliseconds() > 500 {
		pr.LastTime = time.Now()
		select {
		case pr.chUploaded <- &pr.Uploaded: // 发送进度
		default: // 信道满/已关闭时容错（避免阻塞）
		}
	}
	return n, err
}

func (pr *UploadedProgressReader) Close() error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if !pr.closed && pr.chUploaded != nil {
		// 非阻塞发送 Total，避免信道已满或已关闭导致的问题
		select {
		case pr.chUploaded <- &pr.Total:
		default:
		}
		close(pr.chUploaded)
		pr.chUploaded = nil
		pr.closed = true
	}

	// 关闭底层 Reader（兼容处理）
	if closer, ok := pr.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
