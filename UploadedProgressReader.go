package gspider

import (
	"bytes"
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type UploadedProgressReader struct {
	Reader     io.Reader       // 底层读取器
	Uploaded   int64           // 已上传字节数
	Total      int64           // 总字节数
	LastTime   time.Time       // 上次发送进度的时间
	chUploaded chan *int64     // 进度通知通道
	closed     bool            // 标记是否已关闭
	lastSent   int64           // 上次发送的进度值
	mu         sync.Mutex      // 并发控制锁
	ctx        context.Context // 上下文，用于取消操作
	offset     int64           // 当前读取偏移量
	// 新增：缓存底层Reader的Seek接口
	seeker io.Seeker // 缓存转换后的Seeker接口
}

// 初始化seeker接口（延迟初始化）
func (pr *UploadedProgressReader) initSeeker() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if pr.seeker == nil {
		pr.seeker, _ = pr.Reader.(io.Seeker)
	}
}

// Read 实现io.Reader接口
func (pr *UploadedProgressReader) Read(p []byte) (n int, err error) {
	// 检查上下文是否已取消
	if pr.ctx != nil {
		select {
		case <-pr.ctx.Done():
			return 0, pr.ctx.Err()
		default:
		}
	}

	// 读取数据
	n, err = pr.Reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&pr.Uploaded, int64(n))
		pr.offset += int64(n)
	}

	// 检查是否已关闭
	pr.mu.Lock()
	isClosed := pr.closed
	pr.mu.Unlock()

	if isClosed || pr.chUploaded == nil {
		return n, err
	}

	// 定期发送进度更新
	pr.mu.Lock()
	if time.Since(pr.LastTime).Milliseconds() > 500 {
		pr.LastTime = time.Now()
		current := atomic.LoadInt64(&pr.Uploaded)
		lastSent := atomic.LoadInt64(&pr.lastSent)

		if current != lastSent {
			select {
			case pr.chUploaded <- &current:
				atomic.StoreInt64(&pr.lastSent, current)
			default:
			}
		}
	}
	pr.mu.Unlock()

	return n, err
}

func (pr *UploadedProgressReader) Seek(offset int64, whence int) (int64, error) {
	pr.initSeeker() // 确保底层 Seeker 已初始化
	if pr.seeker == nil {
		return 0, io.ErrNoProgress // 明确返回“不支持 Seek”错误
	}

	// 1. 执行底层 Seek 操作，获取新偏移量
	newOffset, err := pr.seeker.Seek(offset, whence)
	if err != nil {
		return 0, err // 底层 Seek 失败，直接返回
	}

	// 2. 线程安全地更新 offset 和 Uploaded（核心修复）
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.offset = newOffset // 更新偏移量

	// 根据 whence 正确计算已上传字节数（Uploaded = 新偏移量，因为上传是从当前位置读取）
	atomic.StoreInt64(&pr.Uploaded, newOffset)

	// 3. 发送 Seek 后的进度更新（可选，确保进度实时性）
	if pr.chUploaded != nil {
		current := newOffset
		lastSent := atomic.LoadInt64(&pr.lastSent)
		if current != lastSent {
			select {
			case pr.chUploaded <- &current:
				atomic.StoreInt64(&pr.lastSent, current)
			default:
			}
		}
	}

	return newOffset, nil
}

// Close 实现io.Closer接口
func (pr *UploadedProgressReader) Close() error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if pr.closed {
		return nil
	}

	// 发送最终进度
	if pr.chUploaded != nil {
		current := atomic.LoadInt64(&pr.Uploaded)
		lastSent := atomic.LoadInt64(&pr.lastSent)

		if current != lastSent {
			select {
			case pr.chUploaded <- &current:
				atomic.StoreInt64(&pr.lastSent, current)
			default:
			}
		}
	}

	pr.closed = true
	pr.chUploaded = nil

	// 关闭底层Reader（如果支持）
	if closer, ok := pr.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// 辅助函数：创建可Seek的进度阅读器
func NewUploadedProgressReader(data []byte, ctx context.Context, ch chan *int64) *UploadedProgressReader {
	reader := bytes.NewReader(data)
	return &UploadedProgressReader{
		Reader:     reader,
		seeker:     reader, // 直接设置已知的Seeker
		Total:      int64(len(data)),
		ctx:        ctx,
		chUploaded: ch,
		LastTime:   time.Now(),
	}
}
