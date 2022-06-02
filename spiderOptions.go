package gspider

import "time"

//spiderOptions 配置类
type spiderOptions struct {
	encode           string        // 编码 默认 Auto 中文 GB18030  或 UTF-8
	timeout          time.Duration // 连接超时
	readWriteTimeout time.Duration // 读写超时
	keepAliveTimeout time.Duration // 保持连接超时
	httpProxyInfo    string        // 设置Http代理 例：http://127.0.0.1:1081
	socks5Address    string        //Socks5地址 例：127.0.0.1:7813
	socks5User       string        //Socks5 用户名
	socks5Pass       string        //Socks5 密码
}

//defaultSpiderOptions 默认配置参数
func defaultSpiderOptions() spiderOptions {
	return spiderOptions{
		encode:           "Auto",
		timeout:          30,
		readWriteTimeout: 30,
		keepAliveTimeout: 30,
	}
}

//--------------------------------------------------------------------------------------------------------------
// SpiderOption 采集基本设置
type SpiderOptionInterface interface {
	apply(*spiderOptions) //使用接口
}

//apply 使用这个匿名函数，针对传入的对象，进行操作
func (fdo *funcSpiderOption) apply(so *spiderOptions) {
	fdo.anyfun(so) //执行具体传递进来的函数
}

// funcSpiderOption wraps a function that modifies spiderOptions into an
// implementation of the SpiderOptions interface.
type funcSpiderOption struct {
	anyfun func(*spiderOptions)
}

//newFuncSpiderOption 新建一个匿名函数实体。
//返回接口地址
func newFuncSpiderOption(anonfun func(*spiderOptions)) *funcSpiderOption {
	return &funcSpiderOption{
		anyfun: anonfun,
	}
}

//--------------------------------------------------------------------------------------------------------------
//Encode 设置编码 ， 返回接口指针( 新建一个函数，不执行的，返回他的地址而已)
func Encode(encode string) SpiderOptionInterface {
	//return &funcSpiderOption{
	//	anyfun: func(o *spiderOptions) { //创建一个匿名函数
	//		o.encode = encode
	//	},
	//}

	//下面更简洁而已，上门原理一致
	return newFuncSpiderOption(func(o *spiderOptions) { //创建一个匿名函数
		o.encode = encode
	})
}

//Timeout 设置 连接超时
func Timeout(t time.Duration) SpiderOptionInterface {
	return newFuncSpiderOption(func(o *spiderOptions) { //创建一个匿名函数
		o.timeout = t
	})
}

//ReadWriteTimeout 设置 读写超时
func ReadWriteTimeout(t time.Duration) SpiderOptionInterface {
	return newFuncSpiderOption(func(o *spiderOptions) { //创建一个匿名函数
		o.readWriteTimeout = t
	})
}

//KeepAliveTimeout 设置 保持链接超时
func KeepAliveTimeout(t time.Duration) SpiderOptionInterface {
	return newFuncSpiderOption(func(o *spiderOptions) { //创建一个匿名函数
		o.keepAliveTimeout = t
	})
}
