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

// httpURLRegex: 匹配以 http/https 开头的 URL，遇到空白或常见分隔符就停止。
var httpURLRegex = regexp.MustCompile(`https?://[^\s"'<>()]+`)

// bareURLRegex: 兜底匹配没有协议前缀的域名/链接，比如 bilibili.com/video/xxx。
// 先匹配类似 example.com，再把后面的 path 一并拿上，直到空白或分隔符。
var bareURLRegex = regexp.MustCompile(`\b[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}[^\s"'<>()]*`)

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

func (e *jsonOriginExtractor) ExtractAll(ctx context.Context, rawText string) ([]biz.URLOriPair, error) {
	if rawText == "" {
		return nil, biz.ErrInvalidArgument.WithMessage("url cannot be empty")
	}

	// 1. 先抓 http/https URL
	httpMatches := httpURLRegex.FindAllString(rawText, -1)
	// 2. 再抓裸域名/链接，尽量覆盖没写协议的情况
	bareMatches := bareURLRegex.FindAllString(rawText, -1)

	if len(httpMatches) == 0 && len(bareMatches) == 0 {
		return nil, biz.ErrInvalidArgument.WithMessage("no valid URL found in input text")
	}

	// 去重：同一个“规范化 host+path+query”+Origin 只返回一次
	seen := make(map[string]struct{})
	pairs := make([]biz.URLOriPair, 0, len(httpMatches)+len(bareMatches))

	// 规范化 URL：统一成 https:// + host（去掉前缀 www.）+ path + query
	normalizeKey := func(u string) string {
		u = strings.TrimSpace(u)
		if u == "" {
			return ""
		}

		// 如果没有协议，补一个 https://，方便解析
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "//") {
			u = "https://" + u
		}

		parsed, err := url.Parse(u)
		if err != nil {
			return ""
		}
		host := strings.TrimPrefix(parsed.Hostname(), "www.")
		if host == "" {
			return ""
		}
		path := parsed.EscapedPath()
		query := parsed.RawQuery
		if query != "" {
			return host + path + "?" + query
		}
		return host + path
	}

	process := func(foundURL string) {
		origin, err := e.parseAndFindOrigin(foundURL)
		if err != nil {
			return
		}
		cleanURL := strings.TrimSpace(foundURL)
		normKey := normalizeKey(cleanURL)
		if normKey == "" {
			return
		}
		key := normKey + "|" + origin
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		pairs = append(pairs, biz.URLOriPair{URL: cleanURL, Origin: origin})
	}

	// 对 httpMatches 里的每一个匹配，再按内部 http/https 切分，处理“多个 URL 黏在一起”的情况
	splitHTTP := func(raw string) []string {
		var res []string
		i := 0
		for i < len(raw) {
			idx := strings.Index(raw[i:], "http")
			if idx == -1 {
				break
			}
			idx += i
			// 确认是 http:// 或 https://
			if !strings.HasPrefix(raw[idx:], "http://") && !strings.HasPrefix(raw[idx:], "https://") {
				i = idx + 4
				continue
			}
			// 找下一个 http/https 的起点，当前 URL 到那里结束
			nextHTTP := strings.Index(raw[idx+7:], "http://")
			nextHTTPS := strings.Index(raw[idx+8:], "https://")
			next := -1
			if nextHTTP != -1 {
				next = idx + 7 + nextHTTP
			}
			if nextHTTPS != -1 {
				cand := idx + 8 + nextHTTPS
				if next == -1 || cand < next {
					next = cand
				}
			}
			if next == -1 {
				res = append(res, raw[idx:])
				break
			}
			res = append(res, raw[idx:next])
			i = next
		}
		return res
	}

	for _, u := range httpMatches {
		for _, part := range splitHTTP(u) {
			process(part)
		}
	}
	for _, u := range bareMatches {
		process(u)
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
