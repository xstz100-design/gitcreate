package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ─────────────────── 距离缓存 ───────────────────

type distanceCacheEntry struct {
	km        float64
	expiresAt time.Time
}

var (
	distanceCacheMu sync.Mutex
	distanceCacheMap = make(map[string]distanceCacheEntry)
)

const distanceCacheTTL = 7 * 24 * time.Hour // 7 天

func distanceCacheKey(origin, destination string) string {
	return strings.ToLower(strings.TrimSpace(origin)) + "||" + strings.ToLower(strings.TrimSpace(destination))
}

func getDistanceCache(key string) (float64, bool) {
	distanceCacheMu.Lock()
	defer distanceCacheMu.Unlock()
	if e, ok := distanceCacheMap[key]; ok && time.Now().Before(e.expiresAt) {
		return e.km, true
	}
	return 0, false
}

func setDistanceCache(key string, km float64) {
	distanceCacheMu.Lock()
	defer distanceCacheMu.Unlock()
	// 顺手清理过期条目（简单防泄漏）
	for k, e := range distanceCacheMap {
		if time.Now().After(e.expiresAt) {
			delete(distanceCacheMap, k)
		}
	}
	distanceCacheMap[key] = distanceCacheEntry{km: km, expiresAt: time.Now().Add(distanceCacheTTL)}
}

// ─────────────────── Distance Matrix ───────────────────

// GetDistanceKM 通过 Google Maps Distance Matrix API 计算两地距离（公里）。
// 结果缓存 7 天，相同 origin/destination 不重复调用 API。
func GetDistanceKM(origin, destination, apiKey string) (float64, error) {
	if apiKey == "" {
		return 0, fmt.Errorf("Google Maps API key 未配置")
	}
	if origin == "" || destination == "" {
		return 0, fmt.Errorf("origin 或 destination 为空")
	}

	cacheKey := distanceCacheKey(origin, destination)
	if km, ok := getDistanceCache(cacheKey); ok {
		return km, nil
	}

	reqURL := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/distancematrix/json?origins=%s&destinations=%s&key=%s",
		url.QueryEscape(origin),
		url.QueryEscape(destination),
		apiKey,
	)

	resp, err := http.Get(reqURL) //nolint:gosec
	if err != nil {
		return 0, fmt.Errorf("无法连接 Google Maps API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
		Rows   []struct {
			Elements []struct {
				Status   string `json:"status"`
				Distance struct {
					Value int `json:"value"` // 单位：米
				} `json:"distance"`
			} `json:"elements"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("解析 Google Maps 响应失败: %w", err)
	}
	if result.Status != "OK" {
		return 0, fmt.Errorf("Google Maps API 返回错误: %s", result.Status)
	}
	if len(result.Rows) == 0 || len(result.Rows[0].Elements) == 0 {
		return 0, fmt.Errorf("Google Maps 未返回距离数据")
	}
	elem := result.Rows[0].Elements[0]
	if elem.Status != "OK" {
		return 0, fmt.Errorf("Google Maps 路径计算失败: %s", elem.Status)
	}

	// 米 → 公里，保留2位小数
	km := float64(elem.Distance.Value) / 1000.0
	setDistanceCache(cacheKey, km)
	return km, nil
}
