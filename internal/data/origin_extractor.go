package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github/heimaolst/collectionbox/internal/biz"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// 1. 定义 JSON 结构体 (可以放在 data 层)
type originConfig struct {
	Supports []string `json:"support"`
	Items    []struct {
		Host   string `json:"host"`
		Origin string `json:"origin"`
	} `json:"items"`
}

var urlRegex = regexp.MustCompile(`\b(https?://[^\s]+|[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})\b`)

// 2. 定义实现
type jsonOriginExtractor struct {
	originMap map[string]string
}

// urlOrigin was a local helper; use biz.URLOriPair instead for cross-layer use.

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
// Extract 从文本中提取*第一个*受支持的 Origin
// func (e *jsonOriginExtractor) Extract(ctx context.Context, rawText string) (string, error) {
// 	if rawText == "" {
// 		return "", biz.ErrInvalidArgument.WithMessage("url cannot be empty")
// 	}

// 	// 1. 使用正则查找第一个匹配的 URL 字符串
// 	foundURL := urlRegex.FindString(rawText)

// 	if foundURL == "" {
// 		return "", biz.ErrInvalidArgument.WithMessage("no valid URL found in input text")
// 	}

// 	// 2. 将找到的 URL (例如 "mail.google.com") 交给辅助函数去解析
// 	return e.parseAndFindOrigin(foundURL)
// }

// --- 方案二：ExtractAll (提取所有匹配的) ---

func (e *jsonOriginExtractor) ExtractAll(ctx context.Context, rawText string) ([]biz.URLOriPair, error) {
	if rawText == "" {
		return nil, biz.ErrInvalidArgument.WithMessage("url cannot be empty")
	}

	// 1. 使用正则查找所有匹配的
	foundURLs := urlRegex.FindAllString(rawText, -1)

	if len(foundURLs) == 0 {
		return nil, biz.ErrInvalidArgument.WithMessage("no valid URL found in input text")
	}

	// 去重：同一个 URL+Origin 只返回一次
	seen := make(map[string]struct{})
	pairs := make([]biz.URLOriPair, 0, len(foundURLs))

	for _, foundURL := range foundURLs {
		// 2. 尝试解析并查找 (复用辅助函数)
		origin, err := e.parseAndFindOrigin(foundURL)
		if err != nil {
			continue
		}
		key := strings.TrimSpace(foundURL) + "|" + origin
		if _, ok := seen[key]; ok {
			continue
		}
		pairs = append(pairs, biz.URLOriPair{URL: strings.TrimSpace(foundURL), Origin: origin})
		seen[key] = struct{}{}
	}

	if len(pairs) == 0 {
		return nil, biz.ErrInvalidArgument.WithMessage("no *supported* origin found in input text")
	}

	return pairs, nil
}
func (e *jsonOriginExtractor) parseAndFindOrigin(urlToParse string) (string, error) {
	// 1. Trim
	preprocessedURL := strings.TrimSpace(urlToParse)

	// 2. 智能预处理 (你原有的逻辑)
	if !strings.HasPrefix(preprocessedURL, "http://") && !strings.HasPrefix(preprocessedURL, "https://") && !strings.HasPrefix(preprocessedURL, "//") {
		// 检查是否是其他 "坏" 协议
		if strings.Contains(preprocessedURL, "://") {
			return "", biz.ErrInvalidArgument.WithMessage("unsupported protocol scheme")
		}
		// 手动添加 "//" 使其变为 "协议相对 URL"
		preprocessedURL = "//" + preprocessedURL
	}

	// 3. 解析
	parsedURL, err := url.Parse(preprocessedURL)
	if err != nil {
		return "", biz.ErrInvalidArgument.WithMessage("invalid url format: " + err.Error())
	}

	// 4. 获取 Hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return "", biz.ErrInvalidArgument.WithMessage("url is missing a host")
	}

	// 5. 【关键修改】使用 publicsuffix 来获取 "eTLD+1" (例如: gemini.com)

	host, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		// 如果 `hostname` 是 "localhost"、IP 地址或无效域名(如 "README.md")
		// publicsuffix 会返回错误。

		// 我们可以检查是否是 "README.md" 这类情况 (不包含点)
		// 但更简单的做法是回退到使用原始 hostname
		// 这样你的 originMap 仍然可以支持 "localhost" 或特定 IP
		if hostname == "localhost" {
			host = hostname
		} else {
			// 如果不是 localhost 且解析失败 (比如 "README.md")
			// 我们可以直接返回错误，因为它肯定不在 originMap 中
			return "", biz.ErrInvalidArgument.WithMessage("invalid host: " + hostname)
		}
	}

	// 6. 查找 map (现在 host 已经是 "gemini.com" 这样的格式了)
	if origin, ok := e.originMap[host]; ok {
		return origin, nil
	}

	// 7. 查找失败
	return "", biz.ErrInvalidArgument.WithMessage("unsupported origin: " + host)
}
