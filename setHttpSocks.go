package gspider

//SetHttpProxy 设置Http代理 例：http://127.0.0.1:1081
func (ros *requests) SetHttpProxy(url string) *requests {
	ros.HttpProxyInfo = url
	return ros
}
func (ros *requests) GetHttpProxy() string {
	return ros.HttpProxyInfo
}

//SetSocks5Proxy 设置Socks5代理 例：127.0.0.1:7813  用户名密码空就 ""
func (ros *requests) SetSocks5Proxy(url, username, password string) *requests {
	ros.Socks5Address = url
	ros.Socks5User = username
	ros.Socks5Pass = password
	return ros
}
func (ros *requests) GetSocks5Proxy() (string, string, string) {
	return ros.Socks5Address, ros.Socks5User, ros.Socks5Pass
}
