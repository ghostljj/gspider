package gspider

import "time"

//SetHttpProxy 设置Http代理 例：http://127.0.0.1:1081
func (s *Spider) SetHttpProxy(url string) *Spider {
	s.dopts.httpProxyInfo = url
	return s
}
func (s *Spider) GetHttpProxy() string {
	return s.dopts.httpProxyInfo
}

//SetSocks5Proxy 设置Socks5代理 例：127.0.0.1:7813  用户名密码空就 ""
func (s *Spider) SetSocks5Proxy(url, username, password string) *Spider {
	s.dopts.socks5Address = url
	s.dopts.socks5User = username
	s.dopts.socks5Pass = password
	return s
}
func (s *Spider) GetSocks5Proxy() (string, string, string) {
	return s.dopts.socks5Address, s.dopts.socks5User, s.dopts.socks5Pass
}

//SetEncode 设置编码
func (s *Spider) SetEncode(encode string) *Spider {
	s.dopts.encode = encode
	return s
}
func (s *Spider) GetEncode() string {
	return s.dopts.encode
}

//SetTimeOut 设置 连接超时
func (s *Spider) SetTimeOut(t time.Duration) *Spider {
	s.dopts.timeout = t
	return s
}

//SetReadWriteTimeout 设置 读写超时
func (s *Spider) SetReadWriteTimeout(t time.Duration) *Spider {
	s.dopts.readWriteTimeout = t
	return s
}

//SetKeepAliveTimeout 设置 保持链接超时
func (s *Spider) SetKeepAliveTimeout(t time.Duration) {
	s.dopts.keepAliveTimeout = t
}
