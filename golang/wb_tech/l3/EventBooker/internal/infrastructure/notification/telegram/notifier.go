package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/entity"
	"go.uber.org/zap"
)

const defaultBaseURL = "https://api.telegram.org"

type Config struct {
	Enabled  bool
	BotToken string
	BaseURL  string
}

type Notifier struct {
	bot     *tgbot.Bot
	logger  *zap.Logger
	enabled bool
}

func New(cfg Config, logger *zap.Logger) *Notifier {
	if logger == nil {
		logger = zap.NewNop()
	}

	notifier := &Notifier{
		logger: logger,
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	if _, err := url.ParseRequestURI(baseURL); err != nil {
		logger.Warn("telegram notifier disabled due to invalid base url", zap.String("base_url", baseURL), zap.Error(err))
		return notifier
	}

	botToken := strings.TrimSpace(cfg.BotToken)
	if !cfg.Enabled || botToken == "" {
		logger.Info("telegram notifier disabled due to missing configuration")
		return notifier
	}

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	botClient, err := tgbot.New(
		botToken,
		tgbot.WithServerURL(strings.TrimRight(baseURL, "/")),
		tgbot.WithSkipGetMe(),
		tgbot.WithHTTPClient(5*time.Second, httpClient),
	)
	if err != nil {
		logger.Warn("telegram notifier disabled due to bot init failure", zap.Error(err))
		return notifier
	}

	notifier.bot = botClient
	notifier.enabled = true

	return notifier
}

func (n *Notifier) NotifyBookingExpired(ctx context.Context, notice entity.ExpiredBookingNotice) error {
	if !n.enabled || n.bot == nil {
		return nil
	}

	chatID := strings.TrimSpace(notice.User.TelegramChatID)
	if chatID == "" {
		return nil
	}

	_, err := n.bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   formatExpiredBookingMessage(notice),
	})
	if err != nil {
		return fmt.Errorf("send telegram notification: %w", err)
	}

	return nil
}

func formatExpiredBookingMessage(notice entity.ExpiredBookingNotice) string {
	return fmt.Sprintf(
		"Booking expired\nEvent: %s\nUser: %s\nExpired at: %s",
		notice.Event.Name,
		notice.User.Name,
		notice.Booking.UpdatedAt.UTC().Format(time.RFC3339),
	)
}
