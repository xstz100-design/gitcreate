package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// 全局共享 HTTP client，避免每次调用都创建新连接（socket 泄漏）
var telegramClient = &http.Client{
	Timeout: 15 * time.Second,
}

// SendTelegramMessage 同步发送 Telegram 消息，返回 message_id（失败返回 0）
func SendTelegramMessage(botToken, chatID, text string, replyMarkup interface{}) int64 {
	if botToken == "" || chatID == "" {
		return 0
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	return doTelegramPost(botToken, url, payload)
}

// SendTelegramPhoto 发送图片消息（caption 为 HTML 格式文字），图片 URL 为公网可访问地址
// 失败时自动降级为纯文字消息
func SendTelegramPhoto(botToken, chatID, photoURL, caption string, replyMarkup interface{}) int64 {
	if botToken == "" || chatID == "" {
		return 0
	}
	if photoURL == "" {
		return SendTelegramMessage(botToken, chatID, caption, replyMarkup)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"photo":      photoURL,
		"caption":    caption,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	msgID := doTelegramPost(botToken, url, payload)
	if msgID == 0 {
		// 图片发送失败时降级为文字
		return SendTelegramMessage(botToken, chatID, caption, replyMarkup)
	}
	return msgID
}

func doTelegramPost(botToken, url string, payload map[string]interface{}) int64 {
	body, _ := json.Marshal(payload)
	resp, err := telegramClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[Telegram] 发送失败: %v", err)
		return 0
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil || !result.OK {
		log.Printf("[Telegram] 响应异常: %s", string(respBytes))
		return 0
	}
	return result.Result.MessageID
}

// AnswerCallbackQuery 响应 inline button 点击
func AnswerCallbackQuery(botToken, callbackQueryID, text string) {
	if botToken == "" {
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", botToken)
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
	}
	body, _ := json.Marshal(payload)
	resp, err := telegramClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[Telegram] answerCallbackQuery 失败: %v", err)
		return
	}
	resp.Body.Close()
}

// EscapeHTML HTML 转义
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// CheckGroupMembership 检查指定用户是否是 Telegram 群组/频道成员
// 成员状态: creator / administrator / member / restricted → true
// 非成员: left / kicked → false
func CheckGroupMembership(botToken, chatID string, userID int64) bool {
	if botToken == "" || chatID == "" {
		return false
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChatMember?chat_id=%s&user_id=%d",
		botToken, chatID, userID)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url) //nolint:gosec
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil || !result.OK {
		return false
	}
	switch result.Result.Status {
	case "creator", "administrator", "member", "restricted":
		return true
	}
	return false
}

// SendTelegramMediaGroup 发送多张图片媒体组，然后单独发送带按钮的文字消息。
// imageURLs 为空时降级为纯文字消息；只有1张时使用 sendPhoto。
// Telegram 的 sendMediaGroup 不支持 reply_markup，所以文字+按钮单独发。
func SendTelegramMediaGroup(botToken, chatID string, imageURLs []string, caption string, replyMarkup interface{}) int64 {
	if botToken == "" || chatID == "" {
		return 0
	}
	if len(imageURLs) == 0 {
		return SendTelegramMessage(botToken, chatID, caption, replyMarkup)
	}
	if len(imageURLs) == 1 {
		return SendTelegramPhoto(botToken, chatID, imageURLs[0], caption, replyMarkup)
	}

	// 构建 media 数组，每张图片独立发送（无 caption，避免重复）
	media := make([]map[string]interface{}, len(imageURLs))
	for i, imgURL := range imageURLs {
		media[i] = map[string]interface{}{
			"type":  "photo",
			"media": imgURL,
		}
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMediaGroup", botToken)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"media":   media,
	}
	doTelegramPost(botToken, apiURL, payload)

	// 图片组之后单独发送文字+按钮消息
	return SendTelegramMessage(botToken, chatID, caption, replyMarkup)
}
