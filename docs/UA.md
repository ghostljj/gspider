# 浏览器指纹 UA 获取与版本说明

当你启用 Surf 指纹（`req.SetSurfBrowserProfile(...)` 与 `req.SetSurfOS(...)`）时，推荐使用无副作用的读取方法：

```go
ua := req.GetSurfUserAgent() // 与指纹一致的家族 UA，零网络成本
```

说明：Surf 的 UA 映射按“浏览器家族”固定到特定版本：
- Chrome 家族：`142.x`
- Firefox 家族：`144.x`
- Tor 家族：`128.x`

所以 UA 不会随你选择的 JA 档位主版本变化，这与 Surf 上游实现一致。`GetSurfUserAgent()` 返回的就是 Surf 实际会发送的家族 UA（真实发送由 Surf 在构建请求时生成）。

## 验证真实发送的 UA（轻量探测）

如果你需要验证“实际发出的 UA”，可以发起一次轻量 `HEAD`/`GET` 请求后读取：

```go
res := req.Head("https://httpbin.org/headers")
fmt.Println("Sent UA:", res.GetReqHeader().Get("User-Agent"))
```

## 插件用：按档位主版本替换的 UA

若你的第三方插件需要一个“带档位主版本号”的 UA 字符串（仅做展示或逻辑分支，不用于真实请求），可使用：

```go
ua := req.GetSurfUserAgentByProfileVersion()
```

这会基于当前家族 UA，将其版本号替换为你选择的 JA 档位主版本（例如 `Chrome/58.0.0.0`）。请注意：这只是“推测用”的 UA，Surf 仍会发送家族固定版本。如需覆盖真实请求头，请直接设置：

```go
req.UserAgent = "你的 UA"
// 或
req.OptHeader("User-Agent", "你的 UA")
```