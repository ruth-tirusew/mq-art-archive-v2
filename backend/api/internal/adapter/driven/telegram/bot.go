// Package telegram implements the Telegram Bot API for subscriber-facing digest
// delivery and account linking. This is deliberately separate from
// internal/adapter/driven/events, which uses MTProto (a user-session login, via gotd/td)
// to read channels for scraping — a bot token and a user session are different Telegram
// credentials with different capabilities, and mixing them up would be a real footgun.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const apiBase = "https://api.telegram.org/bot"

type Bot struct {
	token  string
	client *http.Client
}

func NewBot(token string) *Bot {
	return &Bot{
		token:  token,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (b *Bot) Send(ctx context.Context, chatID, message string) error {
	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     message,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}
	_, err := b.call(ctx, "sendMessage", payload, 15*time.Second)
	return err
}

// update is the subset of Telegram's Update object this bot cares about: an incoming
// text message, enough to read /start <token> and /stop.
type update struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

// Handler is called once per incoming text message, with the chat ID (as a string, since
// that's how it's stored and compared elsewhere) and the raw message text.
type Handler func(ctx context.Context, chatID, text string)

// PollUpdates long-polls getUpdates until ctx is cancelled. It has no webhook and needs no
// public HTTPS endpoint — appropriate for a long-lived process like cmd/api, which is
// already always-on, rather than standing up a separate receiver.
func (b *Bot) PollUpdates(ctx context.Context, handle Handler) error {
	var offset int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := b.getUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Transient network/API error: back off briefly and keep polling rather than
			// exiting — a linking flow that silently stops working is worse than one that
			// logs and retries.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			if u.Message == nil || strings.TrimSpace(u.Message.Text) == "" {
				continue
			}
			handle(ctx, fmt.Sprintf("%d", u.Message.Chat.ID), strings.TrimSpace(u.Message.Text))
		}
	}
}

func (b *Bot) getUpdates(ctx context.Context, offset int64) ([]update, error) {
	payload := map[string]any{
		"offset":  offset,
		"timeout": 30,
	}
	raw, err := b.call(ctx, "getUpdates", payload, 35*time.Second)
	if err != nil {
		return nil, err
	}
	var body struct {
		Result []update `json:"result"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("decode getUpdates response: %w", err)
	}
	return body.Result, nil
}

func (b *Bot) call(ctx context.Context, method string, payload map[string]any, timeout time.Duration) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(callCtx, http.MethodPost, apiBase+b.token+"/"+method, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram %s: %w", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram %s: read response: %w", method, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("telegram %s: status %d: %s", method, resp.StatusCode, string(body))
	}
	return body, nil
}
