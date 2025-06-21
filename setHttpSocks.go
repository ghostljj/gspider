package gspider

// SetHttpProxy 设置Http代理 例：http://127.0.0.1:1081
func (ros *Request) SetHttpProxy(url string) *Request {
	ros.HttpProxyInfo = url
	return ros
}
func (ros *Request) GetHttpProxy() string {
	return ros.HttpProxyInfo
}

// SetLocalIP 设置本地ip 例如，127.0.0.1   或者 example.com
func (ros *Request) SetLocalIP(url string) *Request {
	ros.LocalIP = url
	return ros
}

func (ros *Request) GetLocalIP() string {
	return ros.LocalIP
}

// SetSocks5Proxy 设置Socks5代理 例：127.0.0.1:7813  用户名密码空就 ""
func (ros *Request) SetSocks5Proxy(url, username, password string) *Request {
	ros.Socks5Address = url
	ros.Socks5User = username
	ros.Socks5Pass = password
	return ros
}
func (ros *Request) GetSocks5Proxy() (string, string, string) {
	return ros.Socks5Address, ros.Socks5User, ros.Socks5Pass
}
