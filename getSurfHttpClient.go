package gspider

import (
    "context"
    "crypto/tls"
    "fmt"
    "net"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"

    "github.com/enetx/surf"
    "golang.org/x/net/proxy"
)

// —— Surf 枚举类型：更稳妥的系统与浏览器版本设置 ——
type SurfOS int

const (
	SurfOSDefault       SurfOS = iota // 默认桌面（与 Surf 的稳定画像一致）
	SurfOSWindows                     // 桌面 Windows
	SurfOSMacOS                       // 桌面 macOS（当前映射到桌面画像）
	SurfOSLinux                       // 桌面 Linux（当前映射到桌面画像）
	SurfOSAndroid                     // 移动 Android
	SurfOSIOS                         // 移动 iOS
	SurfOSRandomDesktop               // 随机桌面（Windows/macOS/Linux 的稳定画像）
	SurfOSRandomMobile                // 随机移动（Android/iOS 的稳定画像）
)

type SurfBrowserProfile int

const (
	// Disabled 表示未启用 Surf；仅当设置为非 Disabled 档位时才启用 Surf
	SurfBrowserDisabled SurfBrowserProfile = iota
	SurfBrowserDefault                     // 默认稳定版（按浏览器稳定画像）
	// Chrome 指纹档位（可扩展；未知档位将回退到稳定版）
	SurfBrowserChromeStable
	SurfBrowserChrome58
	SurfBrowserChrome62
	SurfBrowserChrome70
	SurfBrowserChrome72
	SurfBrowserChrome83
	SurfBrowserChrome87
	SurfBrowserChrome96
	SurfBrowserChrome100
	SurfBrowserChrome102
	SurfBrowserChrome106
	SurfBrowserChrome120
	SurfBrowserChrome120PQ
	SurfBrowserChrome142
	// Edge（按 Chrome 家族处理）
	SurfBrowserEdgeStable
	SurfBrowserEdge85
	SurfBrowserEdge106
	// Firefox 家族
	SurfBrowserFirefoxStable
	SurfBrowserFirefox55
	SurfBrowserFirefox56
	SurfBrowserFirefox63
	SurfBrowserFirefox65
	SurfBrowserFirefox99
	SurfBrowserFirefox102
	SurfBrowserFirefox105
	SurfBrowserFirefox120
	SurfBrowserFirefox141
	SurfBrowserFirefox144
	SurfBrowserFirefoxPrivate144
	// Tor（按 Firefox 家族处理）
	SurfBrowserTor
	SurfBrowserTorPrivate
	// iOS/Safari 家族
	SurfBrowserSafari  // Safari 自动档位
	SurfBrowserAndroid // Android OkHttp 指纹
	SurfBrowserIOS
	SurfBrowserIOS11
	SurfBrowserIOS12
	SurfBrowserIOS13
	SurfBrowserIOS14
	// Randomized（不固定版本）
	SurfBrowserRandomized
	SurfBrowserRandomizedALPN
	SurfBrowserRandomizedNoALPN
)

func (req *Request) getSurfHttpClient(rp *RequestOptions, res *Response) *http.Client {
	// Surf 模式：仅当 Request 上的 SurfBrowserProfile 为非 Disabled 时启用
	if req.surfBrowserProfile == SurfBrowserDisabled {
		return &http.Client{}
	}
	// 允许整体超时为可选：当 ReadWriteTimeout<=0 时，不设置客户端超时（无限）
	var httpClient *http.Client
	// 当开启 Surf 浏览器指纹模拟时，使用其 Std() 客户端以保留指纹特性

	imp := surf.NewClient().Builder().Impersonate()
	// 使用 Request 上的枚举设置系统
	switch req.surfOS {
	case SurfOSWindows, SurfOSDefault:
		imp = imp.Windows()
	case SurfOSAndroid:
		imp = imp.Android()
	case SurfOSIOS:
		imp = imp.IOS()
	case SurfOSRandomDesktop, SurfOSRandomMobile:
		imp = imp.RandomOS()
	case SurfOSMacOS:
		imp = imp.MacOS()
	case SurfOSLinux:
		imp = imp.Linux()
	default:
		imp = imp.Windows()
	}
	// 浏览器+版本：使用枚举 SurfBrowserProfile（按家族选择 Builder）
	var b *surf.Builder
	switch req.surfBrowserProfile {
	// Chrome 家族（含 Edge/Safari/iOS/Randomized/Android 默认走 Chrome）
	case SurfBrowserDefault,
		SurfBrowserChromeStable, SurfBrowserChrome58, SurfBrowserChrome62, SurfBrowserChrome70, SurfBrowserChrome72,
		SurfBrowserChrome83, SurfBrowserChrome87, SurfBrowserChrome96, SurfBrowserChrome100, SurfBrowserChrome102,
		SurfBrowserChrome106, SurfBrowserChrome120, SurfBrowserChrome120PQ, SurfBrowserChrome142,
		SurfBrowserEdgeStable, SurfBrowserEdge85, SurfBrowserEdge106,
		SurfBrowserRandomized, SurfBrowserRandomizedALPN, SurfBrowserRandomizedNoALPN,
		SurfBrowserSafari, SurfBrowserIOS, SurfBrowserIOS11, SurfBrowserIOS12, SurfBrowserIOS13, SurfBrowserIOS14,
		SurfBrowserAndroid:
		b = imp.Chrome()
	// Firefox 家族（含 Tor）
	case SurfBrowserFirefoxStable, SurfBrowserFirefox55, SurfBrowserFirefox56, SurfBrowserFirefox63, SurfBrowserFirefox65,
		SurfBrowserFirefox99, SurfBrowserFirefox102, SurfBrowserFirefox105, SurfBrowserFirefox120,
		SurfBrowserFirefox141, SurfBrowserFirefox144, SurfBrowserFirefoxPrivate144,
		SurfBrowserTor, SurfBrowserTorPrivate:
		b = imp.FireFox()
	default:
		b = imp.Chrome()
	}

	// 指纹版本（JA3/JA4）：按枚举档位设置（对已知方法进行精确映射，其余家族稳定版）
	ja := b.JA()
	switch req.surfBrowserProfile {
	// —— Chrome 家族 ——
	case SurfBrowserChromeStable:
		b = ja.Chrome()
	case SurfBrowserChrome58:
		b = ja.Chrome58()
	case SurfBrowserChrome62:
		b = ja.Chrome62()
	case SurfBrowserChrome70:
		b = ja.Chrome70()
	case SurfBrowserChrome72:
		b = ja.Chrome72()
	case SurfBrowserChrome83:
		b = ja.Chrome83()
	case SurfBrowserChrome87:
		b = ja.Chrome87()
	case SurfBrowserChrome96:
		b = ja.Chrome96()
	case SurfBrowserChrome100:
		b = ja.Chrome100()
	case SurfBrowserChrome102:
		b = ja.Chrome102()
	case SurfBrowserChrome106:
		b = ja.Chrome106()
	case SurfBrowserChrome142:
		b = ja.Chrome142()
	case SurfBrowserChrome120:
		b = ja.Chrome120()
	case SurfBrowserChrome120PQ:
		b = ja.Chrome120PQ()
	// —— Edge（按 Chrome 家族处理） ——
	case SurfBrowserEdgeStable:
		b = ja.Edge()
	case SurfBrowserEdge85:
		b = ja.Edge85()
	case SurfBrowserEdge106:
		b = ja.Edge106()
	// —— Firefox 家族 ——
	case SurfBrowserFirefoxStable:
		b = ja.Firefox()
	case SurfBrowserFirefox55:
		b = ja.Firefox55()
	case SurfBrowserFirefox56:
		b = ja.Firefox56()
	case SurfBrowserFirefox63:
		b = ja.Firefox63()
	case SurfBrowserFirefox65:
		b = ja.Firefox65()
	case SurfBrowserFirefox99:
		b = ja.Firefox99()
	case SurfBrowserFirefox102:
		b = ja.Firefox102()
	case SurfBrowserFirefox105:
		b = ja.Firefox105()
	case SurfBrowserFirefox120:
		b = ja.Firefox120()
	case SurfBrowserFirefox141:
		b = ja.Firefox141()
	case SurfBrowserFirefox144:
		b = ja.Firefox144()
	case SurfBrowserFirefoxPrivate144:
		b = ja.FirefoxPrivate144()
	// —— Tor ——
	case SurfBrowserTor:
		b = ja.Tor()
	case SurfBrowserTorPrivate:
		b = ja.TorPrivate()
	// —— iOS/Safari 家族 ——
	case SurfBrowserSafari:
		b = ja.Safari()
	case SurfBrowserIOS:
		b = ja.IOS()
	case SurfBrowserIOS11:
		b = ja.IOS11()
	case SurfBrowserIOS12:
		b = ja.IOS12()
	case SurfBrowserIOS13:
		b = ja.IOS13()
	case SurfBrowserIOS14:
		b = ja.IOS14()
	// —— Randomized ——
	case SurfBrowserRandomized:
		b = ja.Randomized()
	case SurfBrowserRandomizedALPN:
		b = ja.RandomizedALPN()
	case SurfBrowserRandomizedNoALPN:
		b = ja.RandomizedNoALPN()
	// —— Android OkHttp ——
	case SurfBrowserAndroid:
		b = ja.Android()
	}

	// 代理映射：优先 SOCKS5，其次显式 HTTP(S) 代理，最后环境代理
	if len(req.Socks5Address) > 0 {
		//socks5://user:pass@host:port
		proxyStr := strings.TrimSpace(req.Socks5Address)
		b = b.Proxy(proxyStr)
	} else if len(req.HttpProxyInfo) > 0 {
		//https://user:pass@host:port
		proxyStr := strings.TrimSpace(req.HttpProxyInfo)
		b = b.Proxy(proxyStr)
	} else if req.HttpProxyAuto {
		var envProxy string
		for _, key := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
			if val := os.Getenv(key); len(strings.TrimSpace(val)) > 0 {
				envProxy = strings.TrimSpace(val)
				break
			}
		}
		if len(envProxy) > 0 {
			b = b.Proxy(envProxy)
		}
	}
	client := b.Build()
	httpClient = client.Std()
	// 尝试应用 mTLS/证书配置到标准客户端（若底层为 *http.Transport）
	if req.Verify && req.tlsClientConfig != nil {
		if tr, ok := httpClient.Transport.(*http.Transport); ok {
			tr.TLSClientConfig = req.tlsClientConfig
		}
	} else {
		// 跳过证书验证（若可用）
		if tr, ok := httpClient.Transport.(*http.Transport); ok {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}

    // Surf 模式下尽可能绑定本地 IP、设置代理并覆盖 DialContext（仅在 *http.Transport 下生效；HTTP/3 不适用）
    if tr, ok := httpClient.Transport.(*http.Transport); ok {
        // 若用户要求强制短连接，在传输层禁用 Keep-Alive
        if req.surfClose {
            tr.DisableKeepAlives = true
        }
        // 应用握手与响应头等待超时（Surf 下可用时）
        if rp.TLSHandshakeTimeout > 0 {
            tr.TLSHandshakeTimeout = time.Duration(rp.TLSHandshakeTimeout) * time.Second
        }
        if rp.ResponseHeaderTimeout > 0 {
            tr.ResponseHeaderTimeout = time.Duration(rp.ResponseHeaderTimeout) * time.Second
        }
        if rp.ExpectContinueTimeout > 0 {
            tr.ExpectContinueTimeout = time.Duration(rp.ExpectContinueTimeout) * time.Second
        }
        if rp.IdleConnTimeout > 0 {
            tr.IdleConnTimeout = time.Duration(rp.IdleConnTimeout) * time.Second
        }
        baseDialer := &net.Dialer{
            Timeout:   time.Duration(rp.Timeout) * time.Second,
            KeepAlive: time.Duration(rp.KeepAliveTimeout) * time.Second,
        }
        if len(req.LocalIP) > 0 {
            var localTCPAddr *net.TCPAddr
            if isIPAddress(req.LocalIP) {
                ip := net.ParseIP(req.LocalIP)
                if ip == nil {
                    res.resBytes = []byte(fmt.Sprintf("无效的IP地址: %s", req.LocalIP))
                    res.err = fmt.Errorf("无效的IP地址: %s", req.LocalIP)
                    return nil
                }
                localTCPAddr = &net.TCPAddr{IP: ip, Port: 0}
            } else {
                addr, err := net.ResolveIPAddr("ip4", req.LocalIP)
                if err != nil {
                    addr, err = net.ResolveIPAddr("ip6", req.LocalIP)
                    if err != nil {
                        res.resBytes = []byte(fmt.Sprintf("域名解析失败: %v", err))
                        res.err = err
                        return nil
                    }
                }
                localTCPAddr = &net.TCPAddr{IP: addr.IP, Port: 0}
            }
            baseDialer.LocalAddr = localTCPAddr
        }

        // —— 代理设置（JA 指纹模式下库侧跳过了 Proxy 中间件，这里在 Transport 上显式配置）——
        if req.HttpProxyAuto {
            tr.Proxy = http.ProxyFromEnvironment
        }
        if len(req.HttpProxyInfo) > 0 {
            if proxyURL, err := url.Parse(strings.TrimSpace(req.HttpProxyInfo)); err == nil {
                tr.Proxy = http.ProxyURL(proxyURL)
            } else {
                res.resBytes = []byte(err.Error())
                res.err = err
                return nil
            }
        }

        tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
            conn, err := baseDialer.DialContext(ctx, network, addr)
            if err != nil {
                return nil, err
            }
            if rp.TcpDelay > 0 {
                time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
            }
            return conn, nil
        }

        // SOCKS5 代理优先：当设置了 Socks5Address 时，覆盖 DialContext 使用 SOCKS5
        if len(req.Socks5Address) > 0 {
            var socksAuth *proxy.Auth
            if len(req.Socks5User) > 0 {
                socksAuth = &proxy.Auth{User: req.Socks5User, Password: req.Socks5Pass}
            }
            socksDialer, err := proxy.SOCKS5("tcp", strings.TrimSpace(req.Socks5Address), socksAuth, baseDialer)
            if err != nil {
                res.resBytes = []byte(err.Error())
                res.err = err
                return nil
            }
            if ctxDialer, ok := socksDialer.(proxy.ContextDialer); ok {
                tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
                    conn, err := ctxDialer.DialContext(ctx, network, addr)
                    if err != nil {
                        return nil, err
                    }
                    if rp.TcpDelay > 0 {
                        time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
                    }
                    return conn, nil
                }
            } else {
                tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
                    conn, err := socksDialer.Dial(network, addr)
                    if err != nil {
                        return nil, err
                    }
                    if rp.TcpDelay > 0 {
                        time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
                    }
                    return conn, nil
                }
            }
        }
    }

	return httpClient
}
