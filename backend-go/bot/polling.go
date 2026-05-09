package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"wholesale/database"
	"wholesale/services"
)

// TelegramUpdate 表示来自 getUpdates 的单条更新
type TelegramUpdate struct {
	UpdateID      int64                  `json:"update_id"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query"`
	Message       *TelegramIncomingMsg   `json:"message"`
}

type TelegramCallbackQuery struct {
	ID      string           `json:"id"`
	Data    string           `json:"data"`
	Message *TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	Chat *TelegramChat `json:"chat"`
}

type TelegramIncomingMsg struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      *TelegramChat `json:"chat"`
	Text      string        `json:"text"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TelegramChat struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Type      string `json:"type"`
}

type getUpdatesResponse struct {
	OK     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

// StartPolling 在独立 goroutine 中长轮询 Telegram 更新
// 参数 stopCh 关闭时停止轮询
func StartPolling(botToken string, stopCh <-chan struct{}) {
	if botToken == "" {
		fmt.Println("[bot] TG_BOT_TOKEN 未配置，跳过 Bot 轮询")
		return
	}

	var offset int64
	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Println("[bot] 开始 Telegram Bot 长轮询...")

	for {
		select {
		case <-stopCh:
			fmt.Println("[bot] 停止轮询")
			return
		default:
		}

		updates, err := fetchUpdates(client, botToken, offset)
		if err != nil {
			fmt.Printf("[bot] getUpdates 错误: %v，5 秒后重试\n", err)
			select {
			case <-stopCh:
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		for _, upd := range updates {
			if upd.UpdateID >= offset {
				offset = upd.UpdateID + 1
			}
			if upd.CallbackQuery != nil {
				// 构造 services.HandleBotCallback 期望的 map 格式
				cqMap := map[string]interface{}{
					"id":   upd.CallbackQuery.ID,
					"data": upd.CallbackQuery.Data,
				}
				updateMap := map[string]interface{}{
					"callback_query": cqMap,
				}
				services.HandleBotCallback(database.DB, updateMap)
			}
			if upd.Message != nil && upd.Message.Chat != nil {
				chat := upd.Message.Chat
				title := chat.Title
				if title == "" {
					title = chat.FirstName
					if chat.LastName != "" {
						title += " " + chat.LastName
					}
				}
				services.UpsertRecentChat(chat.ID, title, chat.Type)
				fmt.Printf("[bot] 收到消息: chat_id=%d type=%s title=%q\n", chat.ID, chat.Type, title)

				// 私聊消息：处理 /start 及账号关联引导
				if chat.Type == "private" && upd.Message.From != nil {
					services.HandlePrivateMessage(database.DB, upd.Message.From.ID, upd.Message.From.FirstName, upd.Message.From.LastName, upd.Message.Text)
				}
			}
		}
	}
}

func fetchUpdates(client *http.Client, token string, offset int64) ([]TelegramUpdate, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=20&allowed_updates=[%%22callback_query%%22,%%22message%%22]&offset=%d",
		token, offset)

	resp, err := client.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result getUpdatesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("Telegram API 返回错误: %s", string(body))
	}
	return result.Result, nil
}
