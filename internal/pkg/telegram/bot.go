package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/akza/akza-api/internal/domain"
	"go.uber.org/zap"
)

type Bot struct {
	token      string
	chatID     string
	httpClient *http.Client
	log        *zap.Logger
}

func New(token, chatID string, log *zap.Logger) *Bot {
	return &Bot{
		token:      token,
		chatID:     chatID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		log:        log,
	}
}

func (b *Bot) IsConfigured() bool { return b.token != "" && b.chatID != "" }

type sendMessagePayload struct {
	ChatID      string          `json:"chat_id"`
	Text        string          `json:"text"`
	ParseMode   string          `json:"parse_mode"`
	ReplyMarkup *inlineKeyboard `json:"reply_markup,omitempty"`
}

type inlineKeyboard struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

type inlineButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

func (b *Bot) SendOrderNotification(order *domain.Order, variant *domain.ProductVariant, siteBase string) {
	if !b.IsConfigured() { return }
	go func() {
		if err := b.sendOrder(order, variant, siteBase); err != nil {
			b.log.Error("telegram notify failed",
				zap.Int64("order_id", order.ID),
				zap.Error(err),
			)
		}
	}()
}

func (b *Bot) sendOrder(order *domain.Order, variant *domain.ProductVariant, siteBase string) error {
	phone := ""
	if order.Phone != nil { phone = "  📞 " + *order.Phone }
	comment := ""
	if order.Comment != nil && *order.Comment != "" { comment = "\n💬 " + *order.Comment }

	text := fmt.Sprintf(
		"🛍 <b>Новая заявка #%d</b>\n\n"+
			"Изделие: <code>%s</code>\n"+
			"Клиент: <b>%s</b>  @%s%s%s",
		order.ID, variant.Slug,
		order.CustomerName, order.TelegramUsername, phone, comment,
	)

	payload := sendMessagePayload{
		ChatID: b.chatID, Text: text, ParseMode: "HTML",
		ReplyMarkup: &inlineKeyboard{
			InlineKeyboard: [][]inlineButton{{
				{Text: "💬 Чат с клиентом", URL: "https://t.me/" + order.TelegramUsername},
				{Text: "🔗 Открыть изделие", URL: siteBase + "/variants/" + variant.Slug},
			}},
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)
	resp, err := b.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return fmt.Errorf("telegram API returned %d", resp.StatusCode) }
	return nil
}
