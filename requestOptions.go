package gspider

//--------------------------------------------------------------------------------------------------------------

type requestOptions struct {
	RefererUrl    string            //来源url
	PostData      string            //提交内容
	IsGetJson     int               //是否接收 Json  -1不发 0否 1是
	IsPostJson    int               //是否提交 Json  -1不发 0否 1是
	Header        map[string]string //头参数
	Cookie        string            //cookie  单独url
	CookieAll     string            //cookieAll 根url+单独url
	RedirectCount int               //重定向次数
}

//defaultRequestOptions 默认配置参数
func defaultRequestOptions() requestOptions {
	return requestOptions{
		RefererUrl:    "",
		PostData:      "",
		IsPostJson:    -1,
		IsGetJson:     -1,
		Header:        make(map[string]string),
		Cookie:        "",
		RedirectCount: 10,
	}
}

func NewRequestOptions(opts ...requestOptionInterface) *requestOptions {
	dros := defaultRequestOptions()
	for _, opt := range opts {
		opt.apply(&dros) //这里是塞入实体，针对实体赋值
	}
	return &dros
}

//--------------------------------------------------------------------------------------------------------------

// NewRequestOptions请求参数 采集基本接口
type requestOptionInterface interface {
	apply(*requestOptions)
}

//funcRequestOption 定义面的接口使用
type funcRequestOption struct {
	anyfun func(*requestOptions)
}

//apply 实现上面的接口，使用这个匿名函数，针对传入的对象，进行操作
func (fro *funcRequestOption) apply(ro *requestOptions) {
	fro.anyfun(ro)
}

//newFuncRequestOption 新建一个匿名函数实体。
//返回接口地址
func newFuncRequestOption(anonfun func(*requestOptions)) *funcRequestOption {
	return &funcRequestOption{
		anyfun: anonfun,
	}
}

//OptRefererUrl 设置来源地址，返回接口指针(新建一个函数，不执行的，返回他的地址而已)
func OptRefererUrl(refererUrl string) requestOptionInterface {
	//return &funcRequestOption{
	//	anyfun: func(ro *RequestOptions) {
	//		ro.refererUrl = refererUrl
	//	},
	//}
	//下面更简洁而已，上门原理一致
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.RefererUrl = refererUrl
	})
}

//OptPostData 提交内容
func OptPostData(postData string) requestOptionInterface {
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.PostData = postData
	})
}

//OptRequestIsGetJson 是否接收Json   isGetJson -1不发 0否 1是
func OptRequestIsGetJson(isGetJson int) requestOptionInterface {
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.IsGetJson = isGetJson
	})
}

//OptRequestIsPostJson 是否提交Json  isPostJson -1不发 0否 1是
func OptRequestIsPostJson(isPostJson int) requestOptionInterface {
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.IsPostJson = isPostJson
	})
}

//OptRequestHeader 设置发送头
func OptRequestHeader(header map[string]string) requestOptionInterface {
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.Header = header
	})
}

//OptRequestCookie 设置当前Url cookie
func OptRequestCookie(cookie string) requestOptionInterface {
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.Cookie = cookie
	})
}

//OptRequestCookieAll 设置当前Url+根Url cookie
func OptRequestCookieAll(cookieAll string) requestOptionInterface {
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.CookieAll = cookieAll
	})
}

//OptRequestRedirectCount 重定向次数
func OptRequestRedirectCount(redirectCount int) requestOptionInterface {
	return newFuncRequestOption(func(ro *requestOptions) {
		ro.RedirectCount = redirectCount
	})
}
