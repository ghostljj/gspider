package gspider

//SetHttpProxy 设置Http代理 例：http://127.0.0.1:1081
func (ros *request) SetHttpProxy(url string) *request {
	ros.HttpProxyInfo = url
	return ros
}
func (ros *request) GetHttpProxy() string {
	return ros.HttpProxyInfo
}

//SetSocks5Proxy 设置Socks5代理 例：127.0.0.1:7813  用户名密码空就 ""
func (ros *request) SetSocks5Proxy(url, username, password string) *request {
	ros.Socks5Address = url
	ros.Socks5User = username
	ros.Socks5Pass = password
	return ros
}
func (ros *request) GetSocks5Proxy() (string, string, string) {
	return ros.Socks5Address, ros.Socks5User, ros.Socks5Pass
}
