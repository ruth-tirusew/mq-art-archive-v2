package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/mq/api/config"
)

func main() {
	cfg := config.Load()
	if cfg.TelegramAPIID == 0 || cfg.TelegramAPIHash == "" {
		log.Fatal("set TELEGRAM_API_ID and TELEGRAM_API_HASH (from https://my.telegram.org/apps)")
	}

	reader := bufio.NewReader(os.Stdin)
	phone := strings.TrimSpace(os.Getenv("TELEGRAM_PHONE"))
	if phone == "" {
		fmt.Print("Phone number (international format): ")
		line, _ := reader.ReadString('\n')
		phone = strings.TrimSpace(line)
	}
	if phone == "" {
		log.Fatal("phone number required (TELEGRAM_PHONE or prompt)")
	}

	code := strings.TrimSpace(os.Getenv("TELEGRAM_CODE"))
	password := strings.TrimSpace(os.Getenv("TELEGRAM_2FA_PASSWORD"))

	session := telegram.FileSessionStorage{Path: cfg.TelegramSessionPath}
	client := telegram.NewClient(cfg.TelegramAPIID, cfg.TelegramAPIHash, telegram.Options{
		SessionStorage: &session,
	})

	flow := auth.NewFlow(
		auth.Constant(phone, password, auth.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
			if code != "" {
				return code, nil
			}
			fmt.Print("OTP code: ")
			line, err := reader.ReadString('\n')
			return strings.TrimSpace(line), err
		})),
		auth.SendCodeOptions{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("not authorized after login")
		}
		fmt.Printf("Session saved to %s\n", cfg.TelegramSessionPath)
		return nil
	}); err != nil {
		log.Fatalf("login failed: %v", err)
	}
}
