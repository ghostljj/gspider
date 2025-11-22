package gspider

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

func buildTransport(req *Request, rp *RequestOptions) *http.Transport {
	ts := &http.Transport{}
	if rp.IdleConnTimeout > 0 {
		ts.IdleConnTimeout = time.Duration(rp.IdleConnTimeout) * time.Second
	}
	ts.TLSHandshakeTimeout = time.Duration(rp.TLSHandshakeTimeout) * time.Second
	ts.ResponseHeaderTimeout = time.Duration(rp.ResponseHeaderTimeout) * time.Second
	if rp.ExpectContinueTimeout > 0 {
		ts.ExpectContinueTimeout = time.Duration(rp.ExpectContinueTimeout) * time.Second
	}
	if req.surfClose {
		ts.DisableKeepAlives = true
	}
	if req.Verify && req.tlsClientConfig != nil {
		ts.TLSClientConfig = req.tlsClientConfig
	} else {
		ts.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if len(req.HttpProxyInfo) > 0 {
		if purl, err := url.Parse(strings.TrimSpace(req.HttpProxyInfo)); err == nil && purl != nil {
			ts.Proxy = http.ProxyURL(purl)
			if purl.User != nil {
				user := purl.User.Username()
				pass, _ := purl.User.Password()
				if len(user) > 0 {
					token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
					if ts.ProxyConnectHeader == nil {
						ts.ProxyConnectHeader = make(http.Header)
					}
					ts.ProxyConnectHeader.Set("Proxy-Authorization", "Basic "+token)
				}
			}
		}
	} else if req.HttpProxyAuto {
		ts.Proxy = http.ProxyFromEnvironment
	}
	baseDialer := &net.Dialer{Timeout: time.Duration(rp.Timeout) * time.Second, KeepAlive: time.Duration(rp.KeepAliveTimeout) * time.Second}
	if len(req.LocalIP) > 0 {
		var localTCPAddr *net.TCPAddr
		if isIPAddress(req.LocalIP) {
			ip := net.ParseIP(req.LocalIP)
			if ip != nil {
				localTCPAddr = &net.TCPAddr{IP: ip, Port: 0}
			}
		} else {
			addr, err := net.ResolveIPAddr("ip4", req.LocalIP)
			if err != nil {
				addr, err = net.ResolveIPAddr("ip6", req.LocalIP)
			}
			if err == nil {
				localTCPAddr = &net.TCPAddr{IP: addr.IP, Port: 0}
			}
		}
		if localTCPAddr != nil {
			baseDialer.LocalAddr = localTCPAddr
		}
	}
	ts.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := baseDialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		if rp.TcpDelay > 0 {
			time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
		}
		return conn, nil
	}
	if len(req.Socks5Address) > 0 {
		var auth *proxy.Auth
		if len(req.Socks5User) > 0 {
			auth = &proxy.Auth{User: req.Socks5User, Password: req.Socks5Pass}
		}
		if d, err := proxy.SOCKS5("tcp", req.Socks5Address, auth, baseDialer); err == nil {
			if cd, ok := d.(proxy.ContextDialer); ok {
				ts.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					c, e := cd.DialContext(ctx, network, addr)
					if e != nil {
						return nil, e
					}
					if rp.TcpDelay > 0 {
						time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
					}
					return c, nil
				}
			} else {
				ts.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					c, e := d.Dial(network, addr)
					if e != nil {
						return nil, e
					}
					if rp.TcpDelay > 0 {
						time.Sleep(time.Duration(rp.TcpDelay) * time.Second)
					}
					return c, nil
				}
			}
		}
	}
	return ts
}
