package bot

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	tb "gopkg.in/telebot.v3"
)

// parseAllowedChatIDs reads the ALLOWED_CHAT_IDS env var (comma-separated list of
// Telegram chat IDs) and returns them as a set. If the env var is empty, returns nil
// (meaning all users are allowed).
func parseAllowedChatIDs() map[int64]bool {
	raw := helpers.Env("ALLOWED_CHAT_IDS")
	if raw == "" {
		return nil
	}
	allowed := make(map[int64]bool)
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			allowed[id] = true
		}
	}
	return allowed
}

func CreateBot(bc *BotController) IBot {
	const ENV_TG_BOT_API_KEY = "BOT_API_KEY"
	botToken := helpers.Env(ENV_TG_BOT_API_KEY)
	if botToken == "" {
		log.Fatalf("Please provide Telegram bot API key as ENV var '%s'", ENV_TG_BOT_API_KEY)
	}

	allowedChatIDs := parseAllowedChatIDs()
	if allowedChatIDs != nil {
		bc.Logf(INFO, nil, "Chat ID allowlist enabled: only %d user(s) permitted", len(allowedChatIDs))
	}

	poller := &tb.LongPoller{Timeout: 20 * time.Second}
	userGuardPoller := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		// Determine chat ID from message or callback
		var chatID int64
		if upd.Message != nil {
			chatID = upd.Message.Chat.ID
		} else if upd.Callback != nil {
			if upd.Callback.Message != nil {
				chatID = upd.Callback.Message.Chat.ID
			} else {
				// Can't determine chat ID for callback without message; allow through
				bc.Logf(TRACE, nil, "Callback without message context. Proceeding.")
				return true
			}
		} else {
			return true
		}

		// Enforce allowlist
		if allowedChatIDs != nil && !allowedChatIDs[chatID] {
			bc.Logf(WARN, nil, "Rejected message from non-allowed chat ID %d", chatID)
			return false
		}

		message := upd.Message
		if message == nil && upd.Callback != nil {
			bc.Logf(TRACE, nil, "Message was nil. Seems to have been a callback. Proceeding.")
			return true
		}
		// TODO: Start goroutine to update data?
		err := bc.Repo.EnrichUserData(message)
		if err != nil {
			bc.Logf(ERROR, nil, "Error encountered in middlewarePoller: %s", err.Error())
		}
		return true
	})

	b, err := tb.NewBot(tb.Settings{
		Token:   botToken,
		Poller:  userGuardPoller,
		OnError: func(e error, context tb.Context) { bc.Logf(WARN, nil, "%s - context: %v", e.Error(), context) },
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Bot{bot: b}
}
