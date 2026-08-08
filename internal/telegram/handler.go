package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/root1/calories-tracker/internal/domain"
	"github.com/root1/calories-tracker/internal/service"
)

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	From      *User  `json:"from"`
	Chat      *Chat  `json:"chat"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type inlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type BotClient struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewBotClient(token string, timeout time.Duration) *BotClient {
	return &BotClient{
		token: token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: "https://api.telegram.org",
	}
}

func (c *BotClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.sendMessage(ctx, chatID, text, nil)
}

func (c *BotClient) SendMessageWithInlineKeyboard(
	ctx context.Context,
	chatID int64,
	text string,
	rows [][]inlineKeyboardButton,
) error {
	return c.sendMessage(ctx, chatID, text, map[string]any{
		"inline_keyboard": rows,
	})
}

func (c *BotClient) AnswerCallbackQuery(ctx context.Context, callbackID, text string) error {
	payload := map[string]any{
		"callback_query_id": callbackID,
	}
	if text != "" {
		payload["text"] = text
	}
	return c.postJSON(ctx, "answerCallbackQuery", payload)
}

func (c *BotClient) sendMessage(ctx context.Context, chatID int64, text string, replyMarkup any) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	return c.postJSON(ctx, "sendMessage", payload)
}

func (c *BotClient) postJSON(ctx context.Context, method string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", method, err)
	}

	url := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s request: %w", method, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s response: %w", method, err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s status %d: %s", method, resp.StatusCode, string(respBody))
	}
	return nil
}

type Handler struct {
	meals  *service.MealService
	auth   *service.AuthService
	bot    *BotClient
	webURL string
	logger *slog.Logger
}

func NewHandler(
	meals *service.MealService,
	auth *service.AuthService,
	bot *BotClient,
	webURL string,
	logger *slog.Logger,
) *Handler {
	webURL = strings.TrimRight(strings.TrimSpace(webURL), "/")
	if webURL == "" {
		webURL = "https://calories.fittracker.store"
	}
	return &Handler{
		meals:  meals,
		auth:   auth,
		bot:    bot,
		webURL: webURL,
		logger: logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var update Update
	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))

	go h.handleUpdate(r.Context(), update)
}

func (h *Handler) handleUpdate(parent context.Context, update Update) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 45*time.Second)
	defer cancel()

	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update.CallbackQuery)
		return
	}

	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return
	}

	msg := update.Message
	telegramID := msg.From.ID
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)
	lang := h.auth.LanguageForTelegram(ctx, telegramID)

	if text == "" {
		_ = h.bot.SendMessage(ctx, chatID, msgTextOnly(lang))
		return
	}

	if strings.HasPrefix(text, "/") {
		h.handleCommand(ctx, chatID, telegramID, text, lang)
		return
	}

	meal, err := h.meals.LogMealFromText(ctx, telegramID, text)
	if err != nil {
		h.logger.Error("log meal failed", "err", err, "telegram_id", telegramID)
		_ = h.bot.SendMessage(ctx, chatID, msgParseFailed(lang))
		return
	}

	if err := h.bot.SendMessage(ctx, chatID, service.FormatMealReply(meal, lang)); err != nil {
		h.logger.Error("send reply failed", "err", err)
	}
}

func (h *Handler) handleCallback(ctx context.Context, cb *CallbackQuery) {
	if cb.From == nil {
		return
	}
	telegramID := cb.From.ID
	chatID := telegramID
	if cb.Message != nil && cb.Message.Chat != nil {
		chatID = cb.Message.Chat.ID
	}

	data := strings.TrimSpace(cb.Data)
	if strings.HasPrefix(data, "lang:") {
		lang := strings.TrimPrefix(data, "lang:")
		if !domain.IsValidLanguage(lang) {
			_ = h.bot.AnswerCallbackQuery(ctx, cb.ID, "")
			return
		}
		if _, err := h.auth.SetLanguage(ctx, telegramID, lang); err != nil {
			h.logger.Error("set language", "err", err, "telegram_id", telegramID)
			_ = h.bot.AnswerCallbackQuery(ctx, cb.ID, msgLanguageSaveFailed(lang))
			return
		}
		_ = h.bot.AnswerCallbackQuery(ctx, cb.ID, msgLanguageSavedShort(lang))
		_ = h.bot.SendMessage(ctx, chatID, msgLanguageSaved(lang))
		return
	}

	_ = h.bot.AnswerCallbackQuery(ctx, cb.ID, "")
}

func (h *Handler) handleCommand(ctx context.Context, chatID, telegramID int64, text, lang string) {
	switch {
	case strings.HasPrefix(text, "/start"):
		code, err := h.auth.CreateWebLinkCode(ctx, telegramID)
		if err != nil {
			h.logger.Error("create web link code", "err", err, "telegram_id", telegramID)
			_ = h.bot.SendMessage(ctx, chatID, msgCodeFailed(lang))
			return
		}
		_ = h.bot.SendMessage(ctx, chatID, msgStart(h.webURL, code))
		_ = h.bot.SendMessageWithInlineKeyboard(ctx, chatID, msgChooseLanguage(), languageKeyboard())
	case strings.HasPrefix(text, "/language"), strings.HasPrefix(text, "/lang"):
		_ = h.bot.SendMessageWithInlineKeyboard(ctx, chatID, msgChooseLanguage(), languageKeyboard())
	case strings.HasPrefix(text, "/help"):
		_ = h.bot.SendMessage(ctx, chatID, msgHelp(lang, h.webURL))
	default:
		_ = h.bot.SendMessage(ctx, chatID, msgUnknownCommand(lang))
	}
}

func languageKeyboard() [][]inlineKeyboardButton {
	return [][]inlineKeyboardButton{
		{
			{Text: "🇺🇦 Українська", CallbackData: "lang:uk"},
			{Text: "🇬🇧 English", CallbackData: "lang:en"},
		},
	}
}

func msgStart(webURL, code string) string {
	return "Calories Tracker\n\n" +
		"Website / Сайт: " + webURL + "\n\n" +
		"Registration code / Код реєстрації: " + code + "\n" +
		"Valid 15 minutes / Дійсний 15 хвилин.\n\n" +
		"Open the site, sign up with email, password and this code to see your analytics.\n" +
		"Відкрий сайт, зареєструйся з email, паролем і цим кодом — побачиш свою аналітику."
}

func msgChooseLanguage() string {
	return "Обери мову відповідей бота / Choose the bot reply language:"
}

func msgLanguageSaved(lang string) string {
	if lang == domain.LangEN {
		return "Language saved: English.\n\nSend what you ate as text — I’ll count calories and macros.\nYou can mention the day, e.g. “yesterday I ate steak”.\n\nCommands: /help, /language\nWebsite always available after /start."
	}
	return "Мову збережено: українська.\n\nНадішли що з'їв текстом — порахую калорії та БЖВ.\nМожна вказати день, наприклад: «вчора з'їв стейк».\n\nКоманди: /help, /language\nПосилання на сайт завжди в /start."
}

func msgLanguageSavedShort(lang string) string {
	if lang == domain.LangEN {
		return "English selected"
	}
	return "Обрано українську"
}

func msgLanguageSaveFailed(lang string) string {
	if lang == domain.LangEN {
		return "Failed to save language"
	}
	return "Не вдалося зберегти мову"
}

func msgCodeFailed(lang string) string {
	if lang == domain.LangEN {
		return "Could not create a registration code. Try /start again."
	}
	return "Не вдалося створити код. Спробуй /start ще раз."
}

func msgHelp(lang, webURL string) string {
	if lang == domain.LangEN {
		return "Commands:\n/start — website link + registration code + language\n/language — change language\n/help — help\n\nOr just send meals as text, e.g.:\noatmeal with milk, 2 eggs, banana\n\nYou can set the day: “yesterday I ate steak”.\n\nWebsite: " + webURL
	}
	return "Команди:\n/start — посилання на сайт + код реєстрації + мова\n/language — змінити мову\n/help — допомога\n\nАбо просто напиши страви текстом, наприклад:\nвівсянка з молоком, 2 яйця, банан\n\nМожна вказати день: «вчора з'їв стейк».\n\nСайт: " + webURL
}

func msgUnknownCommand(lang string) string {
	if lang == domain.LangEN {
		return "Unknown command. Send a meal description or /help."
	}
	return "Невідома команда. Надішли опис їжі або /help."
}

func msgTextOnly(lang string) string {
	if lang == domain.LangEN {
		return "Send a text description of your meals."
	}
	return "Надішли текстовий опис страв."
}

func msgParseFailed(lang string) string {
	if lang == domain.LangEN {
		return "Could not parse or save the meal. Please try again."
	}
	return "Не вдалося розпарсити або зберегти прийом їжі. Спробуй ще раз."
}
