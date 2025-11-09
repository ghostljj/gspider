# HTTP/3 实现文档

## 概述

本文档描述了在 gspider 中实现的 HTTP/3 (QUIC) 支持功能。该实现允许在 Surf 模式和非 Surf 模式下使用 HTTP/3 协议。

## 技术架构

### 依赖库

- `github.com/enetx/uquic`: QUIC 协议实现
- `github.com/enetx/uquic/http3`: HTTP/3 协议实现
- `github.com/enetx/utls`: TLS 指纹库
- `github.com/enetx/http`: 扩展的 HTTP 客户端库

### 核心组件

#### 1. HTTP/3 适配器 (`api.go`)

创建了 `http3TransportAdapter` 结构体，用于桥接 `enetx/http` 和标准库 `net/http` 之间的接口差异：

```go
type http3TransportAdapter struct {
    http3Transport *http3.RoundTripper
}
```

**主要功能**：
- 实现 `net/http.RoundTripper` 接口
- 将 `net/http.Request` 转换为 `enetx/http.Request`
- 将 `enetx/http.Response` 转换为 `net/http.Response`
- 支持 TLS 配置和连接管理

#### 2. 配置逻辑

在 `sendByte` 方法中添加了 HTTP/3 支持逻辑：

```go
// 非 Surf 模式下的 HTTP/3 支持
var useHTTP3 = req.http3 && req.surfBrowserProfile == SurfBrowserDisabled

if useHTTP3 {
    // 创建 HTTP/3 RoundTripper
    http3Transport := &http3.RoundTripper{
        TLSClientConfig: utlsConfig,
        QuicConfig: &uquic.Config{
            MaxIdleTimeout: time.Duration(rp.IdleConnTimeout) * time.Second,
        },
    }

    // 使用适配器
    httpClient.Transport = &http3TransportAdapter{
        http3Transport: http3Transport,
    }
}
```

## 使用方法

### 1. 非 Surf 模式

```go
req := gs.Session()
req.SetHTTP3(true)

res := req.Get("https://cloudflare-quic.com")
```

### 2. Surf 模式（带浏览器指纹）

```go
req := gs.Session()
req.SetSurfBrowserProfile(gs.SurfBrowserChrome142)
req.SetSurfOS(gs.SurfOSWindows)
req.SetHTTP3(true)

res := req.Get("https://www.google.com")
```

## 关键实现细节

### TLS 配置转换

由于 `http3.RoundTripper` 使用 `github.com/enetx/utls.Config`，而非标准库的 `crypto/tls.Config`，需要进行配置转换：

```go
utlsConfig := &utls.Config{
    InsecureSkipVerify: ts.TLSClientConfig.InsecureSkipVerify,
    ServerName:         ts.TLSClientConfig.ServerName,
    RootCAs:            ts.TLSClientConfig.RootCAs,
    MinVersion:         ts.TLSClientConfig.MinVersion,
    MaxVersion:         ts.TLSClientConfig.MaxVersion,
}

// 证书转换
if len(ts.TLSClientConfig.Certificates) > 0 {
    utlsCerts := make([]utls.Certificate, len(ts.TLSClientConfig.Certificates))
    for i, cert := range ts.TLSClientConfig.Certificates {
        utlsCerts[i] = utls.Certificate{
            Certificate: cert.Certificate,
            PrivateKey:  cert.PrivateKey,
        }
    }
    utlsConfig.Certificates = utlsCerts
}
```

### Request/Response 转换

适配器需要在 `net/http` 和 `enetx/http` 之间转换请求和响应对象：

**Request 转换**：
```go
enetxReq := &enetxhttp.Request{
    Method:           req.Method,
    URL:              req.URL,
    Header:           enetxhttp.Header(req.Header),
    Body:             req.Body,
    // ... 其他字段
}
enetxReq = enetxReq.WithContext(req.Context())
```

**Response 转换**：
```go
resp := &http.Response{
    Status:           enetxResp.Status,
    StatusCode:       enetxResp.StatusCode,
    Header:           http.Header(enetxResp.Header),
    Body:             enetxResp.Body,
    // ... 其他字段
}
```

## 限制和注意事项

### 1. 代理不支持
HTTP/3 (QUIC) 目前不支持通过代理服务器访问。这是 QUIC 协议本身的限制。

### 2. 服务器支持
并非所有 HTTPS 服务器都支持 HTTP/3。常见支持 HTTP/3 的服务：
- Cloudflare CDN
- Google 服务
- Facebook/Meta 服务
- 部分现代 CDN 服务

### 3. 协议回退
如果服务器不支持 HTTP/3，当前实现不会自动回退到 HTTP/2 或 HTTP/1.1，请求可能会失败。

### 4. 性能特性
- **优势**：在高延迟或丢包网络环境下性能更好
- **劣势**：首次连接可能比 TCP 慢（需要 QUIC 握手）

## 测试

### 测试示例
参考 `_examples/testHTTP3.go` 获取完整的测试示例。

### 验证 HTTP/3 连接
可以通过以下方式验证是否使用了 HTTP/3：
1. 检查响应头中的 `alt-svc` 字段
2. 使用网络抓包工具查看是否使用 UDP 443 端口
3. 查看服务器日志（如果可访问）

## 未来改进

1. **自动回退**：实现 HTTP/3 失败时自动回退到 HTTP/2/HTTP/1.1
2. **代理支持**：研究通过 HTTP/3 代理的可能性
3. **性能优化**：优化 QUIC 参数配置
4. **连接池**：实现 HTTP/3 连接复用和池化
5. **诊断工具**：添加 HTTP/3 连接诊断和调试功能

## 参考资料

- [RFC 9114 - HTTP/3](https://www.rfc-editor.org/rfc/rfc9114.html)
- [RFC 9000 - QUIC: A UDP-Based Multiplexed and Secure Transport](https://www.rfc-editor.org/rfc/rfc9000.html)
- [github.com/enetx/uquic](https://github.com/enetx/uquic)
- [github.com/enetx/surf](https://github.com/enetx/surf)
