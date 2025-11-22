package gspider

import (
    "bufio"
    "context"
    "crypto/tls"
    "encoding/base64"
    "fmt"
    "net"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"

    "github.com/enetx/surf"
    "golang.org/x/net/proxy"
    enhttp "github.com/enetx/http"
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

    // 指纹版本（JA3/JA4）：始终启用 JA 指纹设置
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

	// 保留默认 ALPN 由 Surf 配置，代理场景不再强制 HTTP/1.1

    // 代理映射：优先 SOCKS5，其次显式 HTTP(S) 代理，最后环境代理
    if len(req.Socks5Address) > 0 {
        //socks5://user:pass@host:port
        proxyStr := strings.TrimSpace(req.Socks5Address)
        b = b.Proxy(proxyStr)
    } else if len(req.HttpProxyInfo) > 0 {
        //https://user:pass@host:port
        proxyStr := strings.TrimSpace(req.HttpProxyInfo)
        b = b.Proxy(proxyStr)
        // 若 HTTP 代理包含认证信息，设置 CONNECT 阶段的 Proxy-Authorization 头，提升隧道建立成功率
        if u, err := url.Parse(proxyStr); err == nil && u != nil && len(u.User.Username()) > 0 {
            user := u.User.Username()
            pass, _ := u.User.Password()
            token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
            b = b.With(func(client *surf.Client) error {
                if t, ok := client.GetTransport().(*enhttp.Transport); ok {
                    if t.ProxyConnectHeader == nil {
                        t.ProxyConnectHeader = make(enhttp.Header)
                    }
                    t.ProxyConnectHeader.Set("Proxy-Authorization", "Basic "+token)
                }
                return nil
            })
        }
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

		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 统一代理选择：优先 Socks5，其次显式 HttpProxyInfo，最后环境变量
			proxyStr := ""
			if len(req.Socks5Address) > 0 {
				proxyStr = strings.TrimSpace(req.Socks5Address)
			} else if len(req.HttpProxyInfo) > 0 {
				proxyStr = strings.TrimSpace(req.HttpProxyInfo)
			} else if req.HttpProxyAuto {
				for _, key := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
					if val := os.Getenv(key); len(strings.TrimSpace(val)) > 0 {
						proxyStr = strings.TrimSpace(val)
						break
					}
				}
			}

			if req.debug {
				Log.Printf("debug: dial begin url=%s proxy=%s network=%s addr=%s", res.reqUrl, proxyStr, network, addr)
			}

			if len(proxyStr) > 0 {
				u, err := url.Parse(proxyStr)
				if err == nil && u != nil && strings.HasPrefix(strings.ToLower(u.Scheme), "http") {
				if req.debug {
					Log.Printf("debug: http proxy parsed scheme=%s host=%s", u.Scheme, u.Host)
				}
				// HTTP/HTTPS 代理：执行 CONNECT
					conn, err := baseDialer.DialContext(ctx, network, u.Host)
					if err != nil {
						if req.debug {
							Log.Printf("debug: dial proxy host failed host=%s err=%v", u.Host, err)
						}
						return nil, err
					}
					auth := ""
					if u.User != nil {
						user := u.User.Username()
						pass, _ := u.User.Password()
						if len(user) > 0 {
							token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
							auth = "Proxy-Authorization: Basic " + token + "\r\n"
						}
					}
					tryConnect := func(headerHost string, keepAlive bool) (net.Conn, *bufio.Reader, error) {
						if req.debug {
							Log.Printf("debug: CONNECT headerHost=%s keepAlive=%v", headerHost, keepAlive)
						}
						if keepAlive {
							_, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n%sProxy-Connection: keep-alive\r\nConnection: keep-alive\r\n\r\n", addr, headerHost, auth)
						} else {
							_, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n%sProxy-Connection: close\r\nConnection: close\r\n\r\n", addr, headerHost, auth)
						}
						if err != nil {
							if req.debug {
								Log.Printf("debug: write CONNECT err=%v", err)
							}
							return nil, nil, err
						}
						br := bufio.NewReader(conn)
						statusLine, err := br.ReadString('\n')
						if err != nil {
							if req.debug {
								Log.Printf("debug: read CONNECT status err=%v", err)
							}
							return nil, nil, err
						}
						if req.debug {
							Log.Printf("debug: CONNECT status=%s", strings.TrimSpace(statusLine))
						}
                        if !strings.Contains(statusLine, "200") {
                            if req.debug {
                                Log.Printf("debug: CONNECT not 200, fallback to HTTP/1.0")
                            }
                            _, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.0\r\nHost: %s\r\n%s\r\n\r\n", addr, headerHost, auth)
                            if err != nil {
                                return nil, nil, fmt.Errorf("proxy CONNECT failed: %s", strings.TrimSpace(statusLine))
                            }
                            br = bufio.NewReader(conn)
                            statusLine, err = br.ReadString('\n')
                            if req.debug {
                                Log.Printf("debug: CONNECT(1.0) status=%s err=%v", strings.TrimSpace(statusLine), err)
                            }
                            if err != nil || !strings.Contains(statusLine, "200") {
                                return nil, nil, fmt.Errorf("proxy CONNECT failed: %s", strings.TrimSpace(statusLine))
                            }
                        }
						for {
							line, err := br.ReadString('\n')
							if err != nil {
								if req.debug {
									Log.Printf("debug: read CONNECT header err=%v", err)
								}
								return nil, nil, err
							}
							if line == "\r\n" {
								break
							}
						}
						return conn, br, nil
					}
					hostOnly := u.Hostname()
					if hostOnly == "" {
						hostOnly = addr
					}
					if _, _, err = tryConnect(hostOnly, true); err != nil {
						if _, _, err = tryConnect(addr, false); err != nil {
							_ = conn.Close()
							return nil, err
						}
					}
					if rp.TcpDelay > 0 {
						time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
					}
					return conn, nil
				}
				// SOCKS5 或无 scheme：按 SOCKS5 处理
				var socksAuth *proxy.Auth
				// 若采用 URL 形式携带认证，解析之
				u, _ = url.Parse(proxyStr)
				if u != nil && u.User != nil {
					user := u.User.Username()
					pass, _ := u.User.Password()
					if len(user) > 0 {
						socksAuth = &proxy.Auth{User: user, Password: pass}
					}
				}
				host := proxyStr
				if u != nil && len(u.Host) > 0 {
					host = u.Host
				}
					dialer, err := proxy.SOCKS5("tcp", host, socksAuth, baseDialer)
					if err != nil {
						if req.debug {
							Log.Printf("debug: socks5 dialer create host=%s err=%v", host, err)
						}
						return nil, err
					}
					if ctxDialer, ok := dialer.(proxy.ContextDialer); ok {
						conn, err := ctxDialer.DialContext(ctx, network, addr)
						if err != nil {
							if req.debug {
								Log.Printf("debug: socks5 DialContext err=%v", err)
							}
							return nil, err
						}
						if rp.TcpDelay > 0 {
							time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
						}
						return conn, nil
					}
					conn, err := dialer.Dial(network, addr)
					if err != nil {
						if req.debug {
							Log.Printf("debug: socks5 Dial err=%v", err)
						}
						return nil, err
					}
					if rp.TcpDelay > 0 {
						time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
					}
					return conn, nil
				}

			// 直连
			conn, err := baseDialer.DialContext(ctx, network, addr)
			if err != nil {
				if req.debug {
					Log.Printf("debug: direct dial err=%v", err)
				}
				return nil, err
			}
			if rp.TcpDelay > 0 {
				time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
			}
			if req.debug {
				Log.Printf("debug: direct dial ok")
			}
			return conn, nil
		}

		// 统一 DialContext 已处理所有代理情形，无需额外覆盖
	}

	return httpClient
}
