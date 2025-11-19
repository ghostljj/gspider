# 机器指纹测试程序

## 功能说明

这个程序实现了"固定浏览器版本和操作系统，但每台机器使用不同指纹"的方案。

### 核心特性

- ✅ **TLS 指纹固定**: 所有机器使用 Chrome 142 + Windows 10 的 TLS 指纹
- ✅ **应用层差异化**: 通过小版本号、屏幕分辨率、编码支持等特征区分不同机器
- ✅ **配置持久化**: 首次运行生成随机配置，后续运行使用相同配置
- ✅ **随机种子**: 基于随机种子确定性生成所有特征，保证可重现

### 差异化特征

| 特征 | 说明 | 变体数量 |
|------|------|---------|
| **Chrome 小版本号** | 142.0.0.0 ~ 142.0.6261.169 | 6 种 |
| **屏幕分辨率** | 1920x1080, 2560x1440, 1366x768 等 | 8 种 |
| **压缩编码** | gzip/deflate/br/zstd 组合 | 3 种 |
| **DNT 开关** | 约 1/3 机器启用 Do Not Track | 2 种 |

理论组合数: 6 × 8 × 3 × 2 = **288 种不同指纹**

---

## 使用方法

### 1. 首次运行（自动生成配置）

```bash
cd _examples
go run testMachineFingerprint.go
```

**输出示例:**
```
⚡ 首次运行，生成新的机器指纹配置...
✓ 配置已保存到: D:\GoWork\gspider\_examples\machine_config\fingerprint.json

============================================================
📌 当前机器指纹配置
============================================================
🔑 种子 (Seed):          a3f5c8e9d2b1f4a6
🌐 浏览器:               Chrome 142
💻 操作系统:             Windows 10
📦 完整版本:             Chrome 142.0.6261.94
🖥️  屏幕分辨率:           2560x1440
📡 压缩编码:             gzip, deflate, br
🔒 Do Not Track:         已启用
============================================================

🚀 正在测试指纹...
```

### 2. 后续运行（使用已有配置）

```bash
go run testMachineFingerprint.go
```

**输出示例:**
```
✓ 使用已保存的配置

============================================================
📌 当前机器指纹配置
============================================================
🔑 种子 (Seed):          a3f5c8e9d2b1f4a6  # 与首次相同
🌐 浏览器:               Chrome 142
💻 操作系统:             Windows 10
📦 完整版本:             Chrome 142.0.6261.94  # 与首次相同
🖥️  屏幕分辨率:           2560x1440            # 与首次相同
📡 压缩编码:             gzip, deflate, br    # 与首次相同
🔒 Do Not Track:         已启用               # 与首次相同
============================================================
```

### 3. 重置配置（生成新指纹）

```bash
go run testMachineFingerprint.go --reset
```

这会删除旧配置并生成新的随机配置。

---

## 配置文件

配置保存在: `_examples/machine_config/fingerprint.json`

**示例内容:**
```json
{
  "seed": "a3f5c8e9d2b1f4a6",
  "browser": "Chrome 142",
  "os": "Windows 10",
  "minor_version": "142.0.6261.94",
  "viewport": "2560x1440",
  "encoding": "gzip, deflate, br",
  "dnt": "1"
}
```

---

## 在你的代码中使用

### 完整示例

```go
package main

import (
    "fmt"
    gs "github.com/ghostljj/gspider"
)

func main() {
    // 1. 加载机器配置
    config, err := GetOrCreateConfig()
    if err != nil {
        panic(err)
    }

    // 2. 创建请求对象
    req := gs.Session()

    // 3. 应用配置
    headers := ApplyConfig(req, config)

    // 4. 发起请求
    res := req.Get("https://example.com", gs.OptHeader(headers))
    fmt.Println(res.Text())
}
```

### 简化使用

```go
// 直接使用已配置的请求对象
func MakeRequest(url string) string {
    config, _ := GetOrCreateConfig()
    req := gs.Session()
    headers := ApplyConfig(req, config)

    res := req.Get(url, gs.OptHeader(headers))
    return res.Text()
}
```

---

## 多机器部署场景

### 场景 1: 自动模式（推荐）
每台服务器首次运行自动生成不同配置:

```bash
# 服务器 1
go run testMachineFingerprint.go  # 自动生成 seed: a3f5c8e9...

# 服务器 2
go run testMachineFingerprint.go  # 自动生成 seed: f1e2d3c4...

# 服务器 3
go run testMachineFingerprint.go  # 自动生成 seed: 7b8a9c0d...
```

### 场景 2: 手动指定种子
修改代码手动为每台服务器分配种子:

```go
// 服务器 1: 使用固定种子
func main() {
    seed := "1111111111111111"  // 手动指定
    config := GenerateConfigWithSeed(seed)
    saveConfig(config, getConfigPath())
    // ...
}
```

### 场景 3: 环境变量控制
通过环境变量传递种子:

```bash
# 服务器 1
export GSPIDER_SEED="1111111111111111"
go run testMachineFingerprint.go

# 服务器 2
export GSPIDER_SEED="2222222222222222"
go run testMachineFingerprint.go
```

---

## 验证指纹差异

运行测试程序会自动访问两个网站验证指纹:

### 1. TLS 指纹检测 (tls.peet.ws)
检测 JA3/JA4 指纹，验证 TLS 层配置是否正确。

### 2. HTTP Headers 检测 (httpbin.org)
检测应用层特征，验证 User-Agent、编码等差异。

---

## 常见问题

### Q1: 配置文件丢失怎么办？
A: 删除配置文件后，程序会自动生成新的随机配置。建议备份 `fingerprint.json`。

### Q2: 如何确保不同机器使用不同指纹？
A: 每台机器首次运行时会生成随机种子，不同种子生成不同指纹。

### Q3: TLS 指纹会随机变化吗？
A: 不会。所有机器使用相同的 Chrome 142 TLS 指纹，只有应用层特征不同。

### Q4: 能否手动修改配置？
A: 可以直接编辑 `fingerprint.json`，但需要确保数据格式正确。

---

## 技术原理

```
┌─────────────────────────────────────────┐
│         随机种子 (Seed)                  │
│      例: a3f5c8e9d2b1f4a6               │
└──────────────┬──────────────────────────┘
               │
               ├─> Hash 计算 → 索引生成
               │
    ┌──────────┴──────────┬──────────┬──────────┐
    │                     │          │          │
    v                     v          v          v
小版本号              屏幕分辨率   编码支持    DNT开关
142.0.6261.94        2560x1440   gzip,br     启用
(6种选项)            (8种选项)   (3种选项)   (2种选项)
    │                     │          │          │
    └──────────┬──────────┴──────────┴──────────┘
               │
               v
    ┌─────────────────────────┐
    │    应用层指纹 (唯一)     │
    │  + TLS 指纹 (固定)      │
    └─────────────────────────┘
               │
               v
         最终的机器指纹
```

---

## 注意事项

1. **TLS 指纹固定**: 所有机器的 TLS 握手特征相同（Chrome 142）
2. **应用层差异**: 通过 UA 小版本、分辨率等特征区分机器
3. **配置持久化**: 确保配置文件不丢失，否则会生成新指纹
4. **适用场景**: 适合需要统一浏览器版本但希望每台机器有独特特征的场景

---

## 更新日志

- **v1.0** (2025-01-09): 初始版本，支持基于随机种子的指纹生成
