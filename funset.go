package gspider

import "time"

//SetHttpProxy 设置Http代理 例：http://127.0.0.1:1081
func (s *Spider) SetHttpProxy(url string) {
	s.dopts.httpProxyInfo = url
}

//SetSocks5Proxy 设置Socks5代理 例：127.0.0.1:7813  用户名密码空就 ""
func (s *Spider) SetSocks5Proxy(url, username, password string) {
	s.dopts.socks5Address = url
	s.dopts.socks5User = username
	s.dopts.socks5Pass = password
}

//SetEncode 设置编码
func (s *Spider) SetEncode(encode string) {
	s.dopts.encode = encode
}

//SetTimeOut 设置 连接超时
func (s *Spider) SetTimeOut(t time.Duration) {
	s.dopts.timeout = t
}

//SetReadWriteTimeout 设置 读写超时
func (s *Spider) SetReadWriteTimeout(t time.Duration) {
	s.dopts.readWriteTimeout = t
}

//SetKeepAliveTimeout 设置 保持链接超时
func (s *Spider) SetKeepAliveTimeout(t time.Duration) {
	s.dopts.keepAliveTimeout = t
}
