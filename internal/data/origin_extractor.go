package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github/heimaolst/collectionbox/internal/biz"
	"net/url"
	"os"
	"strings"
)

// 1. 定义 JSON 结构体 (可以放在 data 层)
type originConfig struct {
	Supports []string `json:"support"`
	Items    []struct {
		Host   string `json:"host"`
		Origin string `json:"origin"`
	} `json:"items"`
}

// 2. 定义实现
type jsonOriginExtractor struct {
	originMap map[string]string
}

//  3. 构造函数 (替换你的 init())
//     它返回接口和 error
func NewJSONOriginExtractor(filePath string) (biz.OriginExtractor, error) {
	datas, err := os.ReadFile(filePath) // 路径由 main.go 传入
	if err != nil {
		return nil, fmt.Errorf("failed to read origin file: %w", err)
	}

	var cfg originConfig
	if err := json.Unmarshal(datas, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin file: %w", err)
	}

	// 填充 map
	originMap := make(map[string]string)
	for _, v := range cfg.Items {
		originMap[v.Host] = v.Origin
	}

	// 检查 map 是否为空
	if len(originMap) == 0 {
		return nil, fmt.Errorf("origin map is empty, check file: %s", filePath)
	}

	return &jsonOriginExtractor{originMap: originMap}, nil
}

// 4. 实现接口 (这里是您 GetOrgin 的逻辑)
func (e *jsonOriginExtractor) Extract(ctx context.Context, rawURL string) (string, error) {
	if rawURL == "" {
		return "", biz.ErrInvalidArgument.WithMessage("url cannot be empty")
	}

	// 1. 清理首尾空格
	rawURL = strings.TrimSpace(rawURL)

	// 2. 智能预处理：
	//    检查是否缺少 scheme
	preprocessedURL := rawURL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") && !strings.HasPrefix(rawURL, "//") {

		// 还要检查是否是其他 "坏" 协议
		if strings.Contains(rawURL, "://") {
			return "", biz.ErrInvalidArgument.WithMessage("unsupported protocol scheme")
		}

		// 【核心】为 "www.bilibili.com" 或 "bilibili.com/path" 这种输入
		// 手动添加 "//" 使其变为 "协议相对 URL"
		preprocessedURL = "//" + rawURL
	}

	// 3. 解析
	//    现在 url.Parse 几乎不可能失败了
	parsedURL, err := url.Parse(preprocessedURL)
	if err != nil {
		return "", biz.ErrInvalidArgument.WithMessage("invalid url format: " + err.Error())
	}

	// 4. 【关键】使用 .Hostname() 来获取纯净的主机名 (自动去除端口)
	hostname := parsedURL.Hostname()

	// 5. 检查 Hostname 是否为空
	//    (例如，用户只输入了 "http://" 或 "/")
	if hostname == "" {
		return "", biz.ErrInvalidArgument.WithMessage("url is missing a host")
	}

	// 6. 应用 www. 前缀规则
	host := strings.TrimPrefix(hostname, "www.")

	// 7. 查找 map
	if origin, ok := e.originMap[host]; ok {
		return origin, nil
	}

	// 8. 查找失败，返回客户端错误
	return "", biz.ErrInvalidArgument.WithMessage("unsupported origin: " + host)
}
