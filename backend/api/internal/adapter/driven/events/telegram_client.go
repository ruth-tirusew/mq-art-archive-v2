package events

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

// ChannelMessageFetcher fetches recent messages from Telegram channels.
type ChannelMessageFetcher interface {
	FetchChannelMessages(ctx context.Context, channel string, since time.Time, limit int) ([]TelegramMessage, error)
}

type TelegramClientConfig struct {
	APIID       int
	APIHash     string
	SessionPath string
}

// GotdFetcher implements ChannelMessageFetcher using MTProto via gotd.
type GotdFetcher struct {
	apiID       int
	apiHash     string
	sessionPath string
}

func NewGotdFetcher(cfg TelegramClientConfig) *GotdFetcher {
	return &GotdFetcher{
		apiID:       cfg.APIID,
		apiHash:     cfg.APIHash,
		sessionPath: cfg.SessionPath,
	}
}

func (f *GotdFetcher) FetchChannelMessages(ctx context.Context, channel string, since time.Time, limit int) ([]TelegramMessage, error) {
	if f.apiID == 0 || f.apiHash == "" {
		return nil, fmt.Errorf("telegram api credentials not configured")
	}

	session := telegram.FileSessionStorage{Path: f.sessionPath}
	client := telegram.NewClient(f.apiID, f.apiHash, telegram.Options{
		SessionStorage: &session,
	})

	var messages []TelegramMessage
	err := client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("telegram session not authorized; run cmd/telegram-login")
		}

		api := client.API()
		resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: channel,
		})
		if err != nil {
			return fmt.Errorf("resolve %s: %w", channel, err)
		}

		var peer tg.InputPeerClass
		for _, ch := range resolved.Chats {
			switch c := ch.(type) {
			case *tg.Channel:
				peer = &tg.InputPeerChannel{ChannelID: c.ID, AccessHash: c.AccessHash}
			case *tg.Chat:
				peer = &tg.InputPeerChat{ChatID: c.ID}
			}
		}
		if peer == nil {
			return fmt.Errorf("channel %s not found", channel)
		}

		iter := query.Messages(api).GetHistory(peer).BatchSize(limit).Iter()
		count := 0
		for iter.Next(ctx) {
			elem := iter.Value()
			msg, ok := elem.Msg.(*tg.Message)
			if !ok || msg == nil {
				continue
			}
			msgTime := time.Unix(int64(msg.Date), 0).UTC()
			if !since.IsZero() && !msgTime.After(since) {
				break
			}
			messages = append(messages, TelegramMessage{
				Channel:   channel,
				MessageID: msg.ID,
				Text:      msg.Message,
				Date:      msgTime,
			})
			count++
			if count >= limit {
				break
			}
		}
		return iter.Err()
	})
	return messages, err
}

// LoginInteractive performs phone/OTP auth and persists the session.
func LoginInteractive(ctx context.Context, cfg TelegramClientConfig, phone, code, password string) error {
	session := telegram.FileSessionStorage{Path: cfg.SessionPath}
	client := telegram.NewClient(cfg.APIID, cfg.APIHash, telegram.Options{
		SessionStorage: &session,
	})

	flow := auth.NewFlow(
		auth.Constant(phone, password, auth.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
			return code, nil
		})),
		auth.SendCodeOptions{},
	)

	return client.Run(ctx, func(ctx context.Context) error {
		return client.Auth().IfNecessary(ctx, flow)
	})
}
