package outbound

import "context"

// TelegramNotifier sends messages to a linked subscriber's Telegram chat. This is
// separate from the events package's ChannelMessageFetcher (MTProto, read-only, used for
// scraping) — sending to individual subscribers is a Bot API concern with its own
// credential (a bot token, not the scraper's user session).
type TelegramNotifier interface {
	Send(ctx context.Context, chatID, message string) error
}
