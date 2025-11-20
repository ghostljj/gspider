package gspider

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

//--------------------------------------------------------------------------------------------------------------

// Request 这是一个请求对象
type Request struct {
	LocalIP     string // 本地 网络 IP
	UserAgent   string
	cancel      context.CancelFunc
	cancelCause context.CancelCauseFunc
	cancelCtx   context.Context
	// 父级上下文与其取消函数，用于统一取消同一 Request 上的所有并发请求
	baseCtx         context.Context
	baseCancelCause context.CancelCauseFunc
	cancelMu        sync.Mutex

	HttpProxyInfo string // 设置Http代理 例：http://127.0.0.1:1081
	HttpProxyAuto bool   // 自动获取http_proxy变量 默认不开启
	Socks5Address string // Socks5地址 例：127.0.0.1:7813
	Socks5User    string // Socks5 用户名
	Socks5Pass    string // Socks5 密码

	cookieJar       http.CookieJar // CookieJar
	Verify          bool           // https 默认不验证ssl
	tlsClientConfig *tls.Config    // 证书验证配置

	defaultHeaderTemplate map[string]string //发送 请求 头 一些默认值

	wgDone             sync.WaitGroup
	chHttpResponse     chan *http.Response
	chHttpResponseOnce sync.Once // 标记 chHttpResponse 是否已关闭
	ChUploaded         chan *int64
	chUploadedOnce     sync.Once // 标记 ChUploaded 是否已关闭
	ChContentItem      chan []byte
	chContentItemOnce  sync.Once // 标记 ChContentItem 是否已关闭
	groupCtxs          map[string]context.Context
	groupCancelCauses  map[string]context.CancelCauseFunc
	groupCounts        map[string]int // 分组活动请求计数，用于自动清理空分组

    // 默认 Surf 指纹配置
    http3              bool               // 是否启用 HTTP/3（QUIC）指纹
    surfBrowserProfile SurfBrowserProfile // Surf 浏览器+版本指纹
    surfClose          bool               // 是否强制短连接（Connection: close）
    surfOS             SurfOS             // Surf 操作系统指纹
    disableHTTP2       bool
}

// SetHTTP3 启用或关闭 HTTP/3（QUIC）指纹,还没成熟，不支持代理
func (req *Request) SetHTTP3(enable bool) {
	req.http3 = enable
}

// SetSurfBrowserProfile 设置默认的浏览器+版本指纹。设置后即视为启用 Surf 模式。
func (req *Request) SetSurfBrowserProfile(profile SurfBrowserProfile) {
	req.surfBrowserProfile = profile
	req.UserAgent = req.GetSurfUserAgent()
}

// SetSurfClose 控制 Surf 模式下的连接复用；true 表示强制短连接（Connection: close）
func (req *Request) SetSurfClose(enable bool) {
	req.surfClose = enable
}

// SetSurfOS 使用枚举设置操作系统（Windows/Android/iOS/MacOS/Linux/Random 等）
func (req *Request) SetSurfOS(kind SurfOS) {
    req.surfOS = kind
    req.UserAgent = req.GetSurfUserAgent()
}

func (req *Request) SetDisableHTTP2(enable bool) {
    req.disableHTTP2 = enable
}

func (req *Request) Cancel() {
	req.cancelMu.Lock()
	defer req.cancelMu.Unlock()
	if req.cancelCause != nil {
		// 提供可观察的取消原因
		req.cancelCause(errors.New("manual cancel"))
		return
	}
	if req.cancel != nil {
		req.cancel()
	}
}

// CancelAll 取消同一 Request 上所有并发中的请求（取消父级上下文）
func (req *Request) CancelAll() {
	req.cancelMu.Lock()
	defer req.cancelMu.Unlock()
	if req.baseCancelCause != nil {
		req.baseCancelCause(errors.New("cancel all"))
	}
}

// defaultRequestOptions 默认配置参数
func defaultRequest() *Request {
    req := Request{
        UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
        Verify:        false,
        HttpProxyAuto: false,
    }

	// 为该 Request 创建一个父级可取消的上下文，以便 CancelAll 统一取消
	base, baseCancel := context.WithCancelCause(context.Background())
	req.baseCtx = base
	req.baseCancelCause = baseCancel
	// 默认请求期上下文为父级上下文
	req.cancelCtx = base
	req.CookieJarReset()
	req.defaultHeaderTemplate = make(map[string]string)
	req.defaultHeaderTemplate["accept-encoding"] = "gzip, deflate, br"
	req.defaultHeaderTemplate["accept-language"] = "zh-CN,zh;q=0.9"
	req.defaultHeaderTemplate["connection"] = "keep-alive"
	req.defaultHeaderTemplate["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"

	// 默认禁用 Surf 模式，需通过 SetSurfBrowserProfile 显式启用
	req.surfBrowserProfile = SurfBrowserDisabled
	// Surf 相关默认值
    req.http3 = false
    req.surfClose = true // Surf 模式默认短连接，更可控
    req.surfOS = SurfOSDefault
    req.disableHTTP2 = false

	return &req
}

// 已移除 refreshSurfUserAgent：改为使用无副作用的 GetSurfUserAgent 在需要时读取。

// GetSurfUserAgent 返回当前配置下（Surf 指纹启用时）与指纹一致的 UA 字符串。
// - 启用 Surf 时：根据浏览器家族与 OS 计算并返回 UA（不触发网络请求）。
// - 未启用 Surf 时：返回当前 req.UserAgent（传统模式下由用户设定）。
func (req *Request) GetSurfUserAgent() string {
	if req == nil {
		return ""
	}
	if req.surfBrowserProfile == SurfBrowserDisabled {
		return req.UserAgent
	}
	// 归一化 OS（与 refreshSurfUserAgent 保持一致）
	osKind := req.surfOS
	switch osKind {
	case SurfOSDefault:
		osKind = SurfOSWindows
	case SurfOSRandomDesktop:
		osKind = SurfOSWindows
	case SurfOSRandomMobile:
		osKind = SurfOSAndroid
	}
	// 家族选择（Firefox/Tor -> 对应 UA；其他 -> Chrome 家族 UA）
	switch req.surfBrowserProfile {
	case SurfBrowserFirefoxStable, SurfBrowserFirefox55, SurfBrowserFirefox56, SurfBrowserFirefox63, SurfBrowserFirefox65,
		SurfBrowserFirefox99, SurfBrowserFirefox102, SurfBrowserFirefox105, SurfBrowserFirefox120,
		SurfBrowserFirefox141, SurfBrowserFirefox144, SurfBrowserFirefoxPrivate144:
		return uaFirefox(osKind)
	case SurfBrowserTor, SurfBrowserTorPrivate:
		return uaTor(osKind)
	default:
		return uaChrome(osKind)
	}
}

// GetSurfUserAgentByProfileVersion 返回一个按当前档位版本号“调整”的 UA 字符串，仅供插件消费。
// 注意：Surf 实际发送的 UA 固定为家族版本（如 Chrome 142/Firefox 144/Tor 128），此方法不会影响真实请求头。
func (req *Request) GetSurfUserAgentByProfileVersion() string {
	if req == nil {
		return ""
	}
	// 未启用 Surf：直接返回现有 UA
	if req.surfBrowserProfile == SurfBrowserDisabled {
		return req.UserAgent
	}
	base := req.GetSurfUserAgent()
	// 计算家族与主版本
	family := "chrome"
	major := 142
	switch req.surfBrowserProfile {
	// firefox 家族
	case SurfBrowserFirefox55:
		family, major = "firefox", 55
	case SurfBrowserFirefox56:
		family, major = "firefox", 56
	case SurfBrowserFirefox63:
		family, major = "firefox", 63
	case SurfBrowserFirefox65:
		family, major = "firefox", 65
	case SurfBrowserFirefox99:
		family, major = "firefox", 99
	case SurfBrowserFirefox102:
		family, major = "firefox", 102
	case SurfBrowserFirefox105:
		family, major = "firefox", 105
	case SurfBrowserFirefox120:
		family, major = "firefox", 120
	case SurfBrowserFirefox141:
		family, major = "firefox", 141
	case SurfBrowserFirefox144, SurfBrowserFirefoxStable, SurfBrowserFirefoxPrivate144:
		family, major = "firefox", 144
	// tor（按 firefox 家族，版本固定 128）
	case SurfBrowserTor, SurfBrowserTorPrivate:
		family, major = "tor", 128
	// chrome 家族（含 edge/safari/ios/android/randomized/default）
	case SurfBrowserChrome58:
		family, major = "chrome", 58
	case SurfBrowserChrome62:
		family, major = "chrome", 62
	case SurfBrowserChrome70:
		family, major = "chrome", 70
	case SurfBrowserChrome72:
		family, major = "chrome", 72
	case SurfBrowserChrome83:
		family, major = "chrome", 83
	case SurfBrowserChrome87:
		family, major = "chrome", 87
	case SurfBrowserChrome96:
		family, major = "chrome", 96
	case SurfBrowserChrome100:
		family, major = "chrome", 100
	case SurfBrowserChrome102:
		family, major = "chrome", 102
	case SurfBrowserChrome106:
		family, major = "chrome", 106
	case SurfBrowserChrome120, SurfBrowserChrome120PQ:
		family, major = "chrome", 120
	case SurfBrowserChrome142:
		family, major = "chrome", 142
	case SurfBrowserEdge85:
		family, major = "chrome", 85
	case SurfBrowserEdge106:
		family, major = "chrome", 106
	case SurfBrowserDefault, SurfBrowserChromeStable, SurfBrowserEdgeStable,
		SurfBrowserRandomized, SurfBrowserRandomizedALPN, SurfBrowserRandomizedNoALPN,
		SurfBrowserSafari, SurfBrowserIOS, SurfBrowserIOS11, SurfBrowserIOS12, SurfBrowserIOS13, SurfBrowserIOS14,
		SurfBrowserAndroid:
		family, major = "chrome", 142
	}
	// 根据家族调整版本号（仅替换字符串中的版本位）
	switch family {
	case "chrome":
		if strings.Contains(base, "CriOS/") {
			base = replaceVersionToken(base, "CriOS/", fmt.Sprintf("%d.0.0.0", major))
		} else {
			base = replaceVersionToken(base, "Chrome/", fmt.Sprintf("%d.0.0.0", major))
		}
		return base
	case "firefox":
		// Firefox 基础 UA 中的版本位：rv:144.0、Firefox/144.0 以及 Android 的 Gecko/144.0
		base = strings.Replace(base, "rv:144.0", fmt.Sprintf("rv:%d.0", major), 1)
		base = strings.Replace(base, "Firefox/144.0", fmt.Sprintf("Firefox/%d.0", major), 1)
		base = strings.Replace(base, "Gecko/144.0", fmt.Sprintf("Gecko/%d.0", major), 1)
		return base
	case "tor":
		// Tor UA 固定为 128，此处直接返回 base
		return base
	default:
		return base
	}
}

// replaceVersionToken 在 UA 中查找形如 token 后的版本片段（直到下一个空格）并替换为 newVer。
func replaceVersionToken(ua, token, newVer string) string {
	p := strings.Index(ua, token)
	if p < 0 {
		return ua
	}
	start := p + len(token)
	// 寻找分隔符（空格或结束）
	end := strings.Index(ua[start:], " ")
	if end < 0 {
		end = len(ua) - start
	}
	return ua[:start] + newVer + ua[start+end:]
}

// —— UA 映射（与 vendor/enetx/surf 保持一致的近似值）——
func uaChrome(osKind SurfOS) string {
	switch osKind {
	case SurfOSWindows, SurfOSDefault:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
	case SurfOSMacOS:
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
	case SurfOSLinux:
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
	case SurfOSAndroid, SurfOSRandomMobile:
		return "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Mobile Safari/537.36"
	case SurfOSIOS:
		return "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/142.0.7444.77 Mobile/15E148 Safari/604.1"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
	}
}

func uaFirefox(osKind SurfOS) string {
	switch osKind {
	case SurfOSWindows, SurfOSDefault:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:144.0) Gecko/20100101 Firefox/144.0"
	case SurfOSMacOS:
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:144.0) Gecko/20100101 Firefox/144.0"
	case SurfOSLinux:
		return "Mozilla/5.0 (X11; Linux x86_64; rv:144.0) Gecko/20100101 Firefox/144.0"
	case SurfOSAndroid, SurfOSRandomMobile:
		return "Mozilla/5.0 (Android 16; Mobile; rv:144.0) Gecko/144.0 Firefox/144.0"
	case SurfOSIOS:
		return "Mozilla/5.0 (iPhone; CPU iPhone OS 18_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/144.0 Mobile/15E148 Safari/605.1.15"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:144.0) Gecko/20100101 Firefox/144.0"
	}
}

func uaTor(osKind SurfOS) string {
	switch osKind {
	case SurfOSWindows, SurfOSDefault:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0"
	case SurfOSMacOS:
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:128.0) Gecko/20100101 Firefox/128.0"
	case SurfOSLinux:
		return "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"
	case SurfOSAndroid, SurfOSRandomMobile:
		return "Mozilla/5.0 (Android 10; Mobile; rv:128.0) Gecko/134.0 Firefox/128.0"
	case SurfOSIOS:
		return "Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/128.3 Mobile/15E148 Safari/605.1.15"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0"
	}
}

// Session
// 创建Request对象
func Session() *Request {
	return defaultRequest()
}

func SessionWithContext(cancelCtx context.Context) *Request {
	gs := defaultRequest()
	// 将传入的上下文包装为父级可取消上下文，便于统一取消
	base, baseCancel := context.WithCancelCause(cancelCtx)
	gs.baseCtx = base
	gs.baseCancelCause = baseCancel
	gs.cancelCtx = base
	return gs
}

// 安全关闭 chHttpResponse
func (req *Request) safeCloseHttpResponseChan() {
	if req.chHttpResponse != nil {
		req.chHttpResponseOnce.Do(func() {
			close(req.chHttpResponse)
			req.chHttpResponse = nil
		})
	}
}

// 安全关闭 ChUploaded
func (req *Request) safeCloseUploadedChan() {
	if req.ChUploaded != nil {
		req.chUploadedOnce.Do(func() {
			close(req.ChUploaded)
			req.ChUploaded = nil
		})
	}
}

// 安全关闭 ChContentItem
func (req *Request) safeCloseContentItemChan() {
	if req.ChContentItem != nil {
		req.chContentItemOnce.Do(func() {
			close(req.ChContentItem)
			req.ChContentItem = nil
		})
	}
}

// SetTLSClientFile (server.ca)
// 单向 TLS，只验证 server.ca证书链
func (req *Request) SetTLSClientFile(serverCaFile string) {
	byteServerCa, err := os.ReadFile(serverCaFile)
	if err != nil {
		Log.Fatal("ServerCaFile:", err)
	}
	req.SetTLSClient(byteServerCa)
}

// SetTLSClient (server.ca)
// 单向 TLS，只验证 server.ca证书链
func (req *Request) SetTLSClient(serverCa []byte) {

	req.tlsClientConfig = &tls.Config{RootCAs: LoadCa(serverCa),
		Certificates: []tls.Certificate{}} //无需客户端证书
	req.Verify = true
}

// SetmTLSClientFile ("client.crt", "client.key", "server.ca")
// 双向 mTLS  客户端证书  + 服务器 server.ca证书链
func (req *Request) SetmTLSClientFile(clientCrtFile, clientKeyFile, serverCaFile string) {
	byteClientCrt, err := os.ReadFile(clientCrtFile)
	if err != nil {
		Log.Fatal("ClientCaFile:", err)
	}
	byteClientKey, err := os.ReadFile(clientKeyFile)
	if err != nil {
		Log.Fatal("ClientKeyFile:", err)
	}
	byteServerCa, err := os.ReadFile(serverCaFile)
	if err != nil {
		Log.Fatal("ServerCaFile:", err)
	}
	req.SetmTLSClient(byteClientCrt, byteClientKey, byteServerCa)
}

// SetmTLSClient ("client.crt", "client.key", "server.ca")
// 双向 mTLS  客户端证书  + 服务器 server.ca证书链  使用纯字符串可配置在应用中一起生成
func (req *Request) SetmTLSClient(clientCrt, clientKey, serverCa []byte) {
	pair, e := tls.X509KeyPair(clientCrt, clientKey)
	if e != nil {
		Log.Fatal("LoadX509KeyPair:", e)
	}
	//双向 mTLS  客户端证书  + 服务器 server.ca证书链
	req.tlsClientConfig = &tls.Config{RootCAs: LoadCa(serverCa),
		Certificates: []tls.Certificate{pair}} //还需要客户端证书
	req.Verify = true
}

func (req *Request) CancelGroup(group string) {
	req.cancelMu.Lock()
	defer req.cancelMu.Unlock()
	if req.groupCancelCauses != nil {
		if cf, ok := req.groupCancelCauses[group]; ok && cf != nil {
			cf(errors.New("cancel group"))
		}
		delete(req.groupCancelCauses, group)
		if req.groupCtxs != nil {
			delete(req.groupCtxs, group)
		}
		if req.groupCounts != nil {
			delete(req.groupCounts, group)
		}
	}
}

func (req *Request) CancelGroupAll() {
	req.cancelMu.Lock()
	defer req.cancelMu.Unlock()
	if req.groupCancelCauses != nil {
		for g, cf := range req.groupCancelCauses {
			if cf != nil {
				cf(errors.New("cancel group all"))
			}
			delete(req.groupCancelCauses, g)
			delete(req.groupCtxs, g)
			if req.groupCounts != nil {
				delete(req.groupCounts, g)
			}
		}
	}
}
