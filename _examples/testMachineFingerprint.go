package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gs "github.com/ghostljj/gspider"
)

// MachineConfig 机器指纹配置（仅保存实际使用的差异化特征）
type MachineConfig struct {
	MinorVersion string `json:"minor_version"` // Chrome 小版本号（应用到 User-Agent）
	Viewport     string `json:"viewport"`      // 屏幕分辨率（应用到 Sec-CH-Viewport-Width/Height）
	Encoding     string `json:"encoding"`      // 压缩编码支持（应用到 Accept-Encoding）
	DNT          string `json:"dnt"`           // Do Not Track（应用到 DNT 头）
}

// GetOrCreateConfig 获取或创建机器配置
func GetOrCreateConfig() (*MachineConfig, error) {
	configPath := getConfigPath()

	// 1. 尝试读取已有配置
	if data, err := os.ReadFile(configPath); err == nil {
		var config MachineConfig
		if json.Unmarshal(data, &config) == nil {
			fmt.Println("✓ 使用已保存的配置")
			return &config, nil
		}
	}

	// 2. 首次运行，生成新配置
	fmt.Println("⚡ 首次运行，生成新的机器指纹配置...")
	config := generateNewConfig()

	// 3. 保存配置
	if err := saveConfig(config, configPath); err != nil {
		return nil, err
	}

	fmt.Println("✓ 配置已保存到:", configPath)
	return config, nil
}

// 生成配置文件路径（在应用目录下）
func getConfigPath() string {
	// 获取当前工作目录
	workDir, _ := os.Getwd()

	// 配置保存在 machine_config/ 目录下（相对于当前目录）
	configDir := filepath.Join(workDir, "machine_config")
	return filepath.Join(configDir, "fingerprint.json")
}

// 生成新的配置（基于随机种子）
func generateNewConfig() *MachineConfig {
	// 生成随机种子（16位十六进制字符串）
	seedBytes := make([]byte, 8)
	rand.Read(seedBytes)
	seed := hex.EncodeToString(seedBytes)

	// 将种子转换为数字索引
	seedInt := seedToInt(seed)

	config := &MachineConfig{}

	// 基于种子确定性选择特征

	// 1. Chrome 小版本号（模拟不同的更新时间）
	minorVersions := []string{
		"142.0.0.0",
		"142.0.6261.94",
		"142.0.6261.111",
		"142.0.6261.128",
		"142.0.6261.156",
		"142.0.6261.169",
	}
	config.MinorVersion = minorVersions[seedInt%len(minorVersions)]

	// 2. 屏幕分辨率（常见的显示器尺寸）
	viewports := []string{
		"1920x1080", // Full HD (最常见)
		"1366x768",  // 笔记本常见
		"2560x1440", // 2K 显示器
		"1440x900",  // MacBook Air 类似
		"1536x864",  // Surface 类设备
		"1600x900",  // 老款显示器
		"3840x2160", // 4K 显示器
		"1280x720",  // HD 显示器
	}
	config.Viewport = viewports[(seedInt/10)%len(viewports)]

	// 3. 压缩编码支持（部分用户浏览器/网络不支持新编码）
	encodings := []string{
		"gzip, deflate, br",       // 支持 Brotli（现代浏览器）
		"gzip, deflate",           // 不支持 Brotli（部分环境）
		"gzip, deflate, br, zstd", // 支持 Zstd（最新版本）
	}
	config.Encoding = encodings[(seedInt/100)%len(encodings)]

	// 4. DNT (Do Not Track) 设置（约 1/3 用户会开启）
	if seedInt%3 == 0 {
		config.DNT = "1"
	} else {
		config.DNT = ""
	}

	return config
}

// 保存配置到文件
func saveConfig(config *MachineConfig, configPath string) error {
	// 创建目录
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(configPath, data, 0644)
}

// 将种子字符串转换为整数索引
func seedToInt(seed string) int {
	var result int
	fmt.Sscanf(seed[:8], "%x", &result)
	if result < 0 {
		result = -result
	}
	return result
}

// ApplyConfig 应用配置到 Request
func ApplyConfig(req *gs.Request, config *MachineConfig) map[string]string {
	// 应用 Surf 指纹（TLS 层固定）
	req.SetSurfBrowserProfile(gs.SurfBrowserChrome142)
	req.SetSurfOS(gs.SurfOSWindows)

	// 构造自定义 User-Agent（包含小版本号）
	customUA := fmt.Sprintf(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
		config.MinorVersion,
	)

	// 构造请求头
	headers := map[string]string{
		"User-Agent":      customUA,
		"Accept-Encoding": config.Encoding,
	}

	// 如果启用 DNT
	if config.DNT != "" {
		headers["DNT"] = config.DNT
	}

	// 添加 Viewport 相关的客户端提示（Client Hints）
	// 这些是 Chrome 支持的标准 HTTP 头，用于传递设备信息
	headers["Sec-CH-Viewport-Width"] = getViewportWidth(config.Viewport)
	headers["Sec-CH-Viewport-Height"] = getViewportHeight(config.Viewport)

	// 添加 DPR (Device Pixel Ratio) - 基于分辨率推断
	headers["Sec-CH-DPR"] = getDPR(config.Viewport)

	return headers
}

// 从 Viewport 字符串中提取宽度
func getViewportWidth(viewport string) string {
	parts := strings.Split(viewport, "x")
	if len(parts) == 2 {
		return parts[0]
	}
	return "1920"
}

// 从 Viewport 字符串中提取高度
func getViewportHeight(viewport string) string {
	parts := strings.Split(viewport, "x")
	if len(parts) == 2 {
		return parts[1]
	}
	return "1080"
}

// 根据分辨率推断 DPR (Device Pixel Ratio)
func getDPR(viewport string) string {
	width := getViewportWidth(viewport)
	// 高分辨率屏幕通常有更高的 DPR
	switch {
	case width >= "3840": // 4K
		return "2"
	case width >= "2560": // 2K
		return "1.5"
	default:
		return "1"
	}
}

// PrintConfig 打印配置信息
func PrintConfig(config *MachineConfig) {
	line := ""
	for i := 0; i < 60; i++ {
		line += "="
	}

	fmt.Println("\n" + line)
	fmt.Println("📌 当前机器指纹配置")
	fmt.Println(line)
	fmt.Printf("🌐 浏览器:               Chrome 142 (固定)\n")
	fmt.Printf("💻 操作系统:             Windows 10 (固定)\n")
	fmt.Printf("📦 完整版本:             Chrome %s\n", config.MinorVersion)
	fmt.Printf("🖥️  屏幕分辨率:           %s\n", config.Viewport)
	fmt.Printf("📡 压缩编码:             %s\n", config.Encoding)
	if config.DNT != "" {
		fmt.Printf("🔒 Do Not Track:         已启用\n")
	} else {
		fmt.Printf("🔒 Do Not Track:         未启用\n")
	}
	fmt.Println(line + "\n")
}

// ResetConfig 重置配置（生成新的随机指纹）
func ResetConfig() error {
	configPath := getConfigPath()

	// 删除旧配置
	os.Remove(configPath)

	// 生成新配置
	config := generateNewConfig()
	return saveConfig(config, configPath)
}

// ============= 使用示例 =============

func main() {
	// 如果命令行参数包含 --reset，则重置配置
	if len(os.Args) > 1 && os.Args[1] == "--reset" {
		fmt.Println("🔄 正在重置配置...")
		if err := ResetConfig(); err != nil {
			fmt.Println("❌ 重置失败:", err)
			return
		}
		fmt.Println("✓ 配置已重置，请重新运行程序")
		return
	}

	// 1. 获取或创建机器配置
	config, err := GetOrCreateConfig()
	if err != nil {
		fmt.Println("❌ 配置加载失败:", err)
		return
	}

	// 2. 打印配置信息
	PrintConfig(config)

	// 3. 创建请求对象并应用配置
	req := gs.Session()
	headers := ApplyConfig(req, config)

	// 4. 发起测试请求
	fmt.Println("🚀 正在测试指纹...")
	fmt.Println()

	// 测试 1: TLS 指纹检测
	fmt.Println("【测试 1】TLS 指纹检测 (tls.peet.ws)")
	res1 := req.Get("https://tls.peet.ws/api/all", gs.OptHeader(headers))
	if res1.GetErr() != nil {
		fmt.Println("❌ 请求失败:", res1.GetErr())
	} else {
		fmt.Println("✓ 状态码:", res1.GetStatusCode())
		// 解析 JSON 响应（简化输出）
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(res1.GetContent()), &result); err == nil {
			if ja3, ok := result["tls"].(map[string]interface{}); ok {
				if hash, ok := ja3["ja3"].(string); ok {
					fmt.Println("✓ JA3 指纹:", hash)
				}
				if ja4, ok := ja3["ja4"].(string); ok {
					fmt.Println("✓ JA4 指纹:", ja4)
				}
			}
			if http2, ok := result["http2"].(map[string]interface{}); ok {
				if akamai, ok := http2["akamai_fingerprint"].(string); ok {
					fmt.Println("✓ HTTP/2 指纹:", akamai)
				}
			}
		}
	}
	fmt.Println()

	// 测试 2: 普通网站请求
	fmt.Println("【测试 2】普通网站请求 (httpbin.org)")
	res2 := req.Get("https://httpbin.org/headers", gs.OptHeader(headers))
	if res2.GetErr() != nil {
		fmt.Println("❌ 请求失败:", res2.GetErr())
	} else {
		fmt.Println("✓ 状态码:", res2.GetStatusCode())
		// 解析并打印发送的 headers
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(res2.GetContent()), &result); err == nil {
			if headers, ok := result["headers"].(map[string]interface{}); ok {
				fmt.Println("✓ 发送的 User-Agent:", headers["User-Agent"])
				fmt.Println("✓ 发送的 Accept-Encoding:", headers["Accept-Encoding"])
				if dnt, ok := headers["Dnt"]; ok {
					fmt.Println("✓ 发送的 DNT:", dnt)
				}
				// 显示 Viewport 相关的客户端提示
				if vw, ok := headers["Sec-Ch-Viewport-Width"]; ok {
					fmt.Println("✓ 视口宽度:", vw)
				}
				if vh, ok := headers["Sec-Ch-Viewport-Height"]; ok {
					fmt.Println("✓ 视口高度:", vh)
				}
				if dpr, ok := headers["Sec-Ch-Dpr"]; ok {
					fmt.Println("✓ 设备像素比:", dpr)
				}
			}
		}
	}
	fmt.Println()

	// 提示信息
	line := ""
	for i := 0; i < 60; i++ {
		line += "="
	}
	fmt.Println(line)
	fmt.Println("💡 提示:")
	fmt.Println("  - 配置文件位置:", getConfigPath())
	fmt.Println("  - 重置配置命令: go run testMachineFingerprint.go --reset")
	fmt.Println("  - 每次运行使用相同配置，不同机器配置不同")
	fmt.Println(line)
}
