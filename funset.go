package gspider

import "time"

//SetHttpProxy 设置Http代理 例：http://127.0.0.1:1081
func (s *Spider) SetHttpProxy(url string) *Spider {
	s.httpProxyInfo = url
	return s
}
func (s *Spider) GetHttpProxy() string {
	return s.httpProxyInfo
}

//SetSocks5Proxy 设置Socks5代理 例：127.0.0.1:7813  用户名密码空就 ""
func (s *Spider) SetSocks5Proxy(url, username, password string) *Spider {
	s.socks5Address = url
	s.socks5User = username
	s.socks5Pass = password
	return s
}
func (s *Spider) GetSocks5Proxy() (string, string, string) {
	return s.socks5Address, s.socks5User, s.socks5Pass
}

//SetEncode 设置编码
func (s *Spider) SetEncode(encode string) *Spider {
	s.encode = encode
	return s
}
func (s *Spider) GetEncode() string {
	return s.encode
}

//SetTimeOut 设置 连接超时
func (s *Spider) SetTimeOut(t time.Duration) *Spider {
	s.timeout = t
	return s
}

//SetReadWriteTimeout 设置 读写超时
func (s *Spider) SetReadWriteTimeout(t time.Duration) *Spider {
	s.readWriteTimeout = t
	return s
}

//SetKeepAliveTimeout 设置 保持链接超时
func (s *Spider) SetKeepAliveTimeout(t time.Duration) {
	s.keepAliveTimeout = t
}
