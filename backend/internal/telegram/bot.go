package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/domain"
	"yardpass/internal/qr"
	"yardpass/internal/redis"

	"github.com/google/uuid"

	"go.uber.org/zap"
)

type Bot struct {
	token         string
	apiURL        string
	passService   domain.PassService
	residentRepo  domain.ResidentRepository
	apartmentRepo domain.ApartmentRepository
	qrGen         *qr.Generator
	redis         *redis.Client
	logger        *zap.Logger
	states        map[int64]*UserState
	location      *time.Location
}

type UserState struct {
	Step      string
	Data      map[string]interface{}
	ExpiresAt time.Time
}

const (
	StateWaitingGuestType  = "waiting_guest_type"
	StateWaitingCarPlate   = "waiting_car_plate"
	StateWaitingDuration   = "waiting_duration"
	StateWaitingCustomTime = "waiting_custom_time"
	StateWaitingGuestName  = "waiting_guest_name"
)

func NewBot(
	cfg *config.Config,
	passService domain.PassService,
	residentRepo domain.ResidentRepository,
	apartmentRepo domain.ApartmentRepository,
	qrGen *qr.Generator,
	redisClient *redis.Client,
	logger *zap.Logger,
) *Bot {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		logger.Warn("Failed to load Europe/Moscow timezone, using UTC", zap.Error(err))
		location = time.UTC
	}

	bot := &Bot{
		token:         cfg.Telegram.BotToken,
		apiURL:        fmt.Sprintf("https://api.telegram.org/bot%s", cfg.Telegram.BotToken),
		passService:   passService,
		residentRepo:  residentRepo,
		apartmentRepo: apartmentRepo,
		qrGen:         qrGen,
		redis:         redisClient,
		logger:        logger,
		states:        make(map[int64]*UserState),
		location:      location,
	}

	ctx := context.Background()
	if err := bot.SetMyCommands(ctx); err != nil {
		logger.Warn("Failed to set bot commands", zap.Error(err))
	} else {
		logger.Info("Bot commands menu set successfully")
	}

	return bot
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	From      *User  `json:"from"`
	Chat      *Chat  `json:"chat"`
	Text      string `json:"text"`
	Date      int64  `json:"date"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

func (b *Bot) ProcessUpdate(ctx context.Context, update Update) {
	if update.Message != nil {
		b.handleMessage(ctx, *update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallbackQuery(ctx, *update.CallbackQuery)
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg Message) {
	userID := msg.From.ID
	text := msg.Text

	if text == "/start" || text == "/create" || text == "/list" || text == "/revoke" {
		switch text {
		case "/start":
			b.handleStart(ctx, msg)
		case "/create":
			cb := CallbackQuery{
				ID:      "",
				From:    msg.From,
				Message: &msg,
				Data:    "create_pass",
			}
			b.handleCallbackQuery(ctx, cb)
		case "/list":
			cb := CallbackQuery{
				ID:      "",
				From:    msg.From,
				Message: &msg,
				Data:    "list_active",
			}
			b.handleCallbackQuery(ctx, cb)
		case "/revoke":
			cb := CallbackQuery{
				ID:      "",
				From:    msg.From,
				Message: &msg,
				Data:    "revoke_pass",
			}
			b.handleCallbackQuery(ctx, cb)
		}
		return
	}

	state := b.getState(userID)
	if state == nil {
		b.sendMessage(ctx, msg.Chat.ID, "Используйте /start для начала работы")
		return
	}

	switch state.Step {
	case StateWaitingGuestType:
		b.sendMessage(ctx, msg.Chat.ID, "Используйте кнопки для выбора типа гостя")
	case StateWaitingCarPlate:
		b.handleCarPlate(ctx, msg, state)
	case StateWaitingDuration:
		b.handleDuration(ctx, msg, state)
	case StateWaitingCustomTime:
		b.handleCustomTime(ctx, msg, state)
	case StateWaitingGuestName:
		b.handleGuestName(ctx, msg, state)
	default:
		b.sendMessage(ctx, msg.Chat.ID, "Неизвестное состояние. Используйте /start")
		b.clearState(userID)
	}
}

func (b *Bot) handleStart(ctx context.Context, msg Message) {
	userID := msg.From.ID

	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		b.sendMessage(ctx, msg.Chat.ID, "Вы не зарегистрированы как житель. Обратитесь к администратору.")
		return
	}

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "Выдать пропуск гостю", "callback_data": "create_pass"},
			},
			{
				{"text": "Мои активные пропуска", "callback_data": "list_active"},
			},
			{
				{"text": "Отозвать пропуск", "callback_data": "revoke_pass"},
			},
		},
	}

	b.sendMessageWithKeyboard(ctx, msg.Chat.ID, "Добро пожаловать в YardPass!\n\nВыберите действие:", keyboard)
}

func (b *Bot) handleCallbackQuery(ctx context.Context, cb CallbackQuery) {
	userID := cb.From.ID
	data := cb.Data

	switch data {
	case "create_pass":
		keyboard := map[string]interface{}{
			"inline_keyboard": [][]map[string]interface{}{
				{
					{"text": "🚗 На автомобиле", "callback_data": "guest_car"},
				},
				{
					{"text": "🚶 Пеший гость", "callback_data": "guest_pedestrian"},
				},
			},
		}
		b.setState(userID, &UserState{
			Step:      StateWaitingGuestType,
			Data:      make(map[string]interface{}),
			ExpiresAt: time.Now().Add(10 * time.Minute),
		})
		b.sendMessageWithKeyboard(ctx, cb.Message.Chat.ID, "Выберите тип гостя:", keyboard)
		b.answerCallbackQuery(ctx, cb.ID, "")

	case "list_active":
		b.listActivePasses(ctx, cb.Message.Chat.ID, userID)
		b.answerCallbackQuery(ctx, cb.ID, "")

	case "revoke_pass":
		b.showPassesForRevoke(ctx, cb.Message.Chat.ID, userID)
		b.answerCallbackQuery(ctx, cb.ID, "")

	case "guest_car":
		state := b.getState(userID)
		if state == nil {
			b.sendMessage(ctx, cb.Message.Chat.ID, "Сессия истекла. Начните заново с /start")
			b.answerCallbackQuery(ctx, cb.ID, "")
			return
		}
		state.Step = StateWaitingCarPlate
		b.setState(userID, state)
		b.sendMessage(ctx, cb.Message.Chat.ID, "Введите номер автомобиля (на английском, например: A123BC77):")
		b.answerCallbackQuery(ctx, cb.ID, "")

	case "guest_pedestrian":
		state := b.getState(userID)
		if state == nil {
			b.sendMessage(ctx, cb.Message.Chat.ID, "Сессия истекла. Начните заново с /start")
			b.answerCallbackQuery(ctx, cb.ID, "")
			return
		}
		state.Data["is_pedestrian"] = true
		state.Step = StateWaitingDuration
		b.setState(userID, state)
		keyboard := map[string]interface{}{
			"inline_keyboard": [][]map[string]interface{}{
				{
					{"text": "1 час", "callback_data": "duration_1h"},
					{"text": "2 часа", "callback_data": "duration_2h"},
				},
				{
					{"text": "4 часа", "callback_data": "duration_4h"},
					{"text": "До времени", "callback_data": "duration_custom"},
				},
			},
		}
		b.sendMessageWithKeyboard(ctx, cb.Message.Chat.ID, "Выберите срок действия пропуска:", keyboard)
		b.answerCallbackQuery(ctx, cb.ID, "")

	case "duration_1h", "duration_2h", "duration_4h", "duration_custom":
		state := b.getState(userID)
		if state == nil {
			b.sendMessage(ctx, cb.Message.Chat.ID, "Сессия истекла. Начните заново с /start")
			b.answerCallbackQuery(ctx, cb.ID, "")
			return
		}

		switch data {
		case "duration_1h":
			state.Data["duration"] = 1 * time.Hour
		case "duration_2h":
			state.Data["duration"] = 2 * time.Hour
		case "duration_4h":
			state.Data["duration"] = 4 * time.Hour
		case "duration_custom":
			state.Step = StateWaitingCustomTime
			b.setState(userID, state)
			b.sendMessage(ctx, cb.Message.Chat.ID, "Введите время окончания действия пропуска в формате ЧЧ:ММ (например, 22:00):")
			b.answerCallbackQuery(ctx, cb.ID, "")
			return
		}

		state.Step = StateWaitingGuestName
		b.setState(userID, state)
		b.sendMessage(ctx, cb.Message.Chat.ID, "Введите имя гостя (или отправьте '-' чтобы пропустить):")
		b.answerCallbackQuery(ctx, cb.ID, "")

	default:
		if strings.HasPrefix(data, "revoke_pass_") {
			passIDStr := strings.TrimPrefix(data, "revoke_pass_")
			passID, err := uuid.Parse(passIDStr)
			if err != nil {
				b.sendMessage(ctx, cb.Message.Chat.ID, "Ошибка: неверный ID пропуска")
				b.answerCallbackQuery(ctx, cb.ID, "")
				return
			}
			b.revokePass(ctx, cb.Message.Chat.ID, userID, passID)
			b.answerCallbackQuery(ctx, cb.ID, "")
			return
		}
	}
}

func (b *Bot) handleCarPlate(ctx context.Context, msg Message, state *UserState) {
	carPlate := msg.Text
	state.Data["car_plate"] = carPlate

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "1 час", "callback_data": "duration_1h"},
				{"text": "2 часа", "callback_data": "duration_2h"},
			},
			{
				{"text": "4 часа", "callback_data": "duration_4h"},
				{"text": "До времени", "callback_data": "duration_custom"},
			},
		},
	}

	state.Step = StateWaitingDuration
	b.setState(msg.From.ID, state)
	b.sendMessageWithKeyboard(ctx, msg.Chat.ID, "Выберите срок действия пропуска:", keyboard)
}

func (b *Bot) handleDuration(ctx context.Context, msg Message, state *UserState) {
	b.sendMessage(ctx, msg.Chat.ID, "Используйте кнопки для выбора")
}

func (b *Bot) handleCustomTime(ctx context.Context, msg Message, state *UserState) {
	timeStr := msg.Text
	now := time.Now().In(b.location)

	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		b.sendMessage(ctx, msg.Chat.ID, "Неверный формат времени. Введите в формате ЧЧ:ММ (например, 22:00)")
		return
	}

	targetTime := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, b.location)
	if targetTime.Before(now) {
		targetTime = targetTime.Add(24 * time.Hour)
	}

	state.Data["valid_to"] = targetTime.UTC()

	state.Data["valid_to"] = targetTime
	state.Step = StateWaitingGuestName
	b.setState(msg.From.ID, state)
	b.sendMessage(ctx, msg.Chat.ID, "Введите имя гостя (или отправьте '-' чтобы пропустить):")
}

func (b *Bot) handleGuestName(ctx context.Context, msg Message, state *UserState) {
	guestName := msg.Text
	if guestName != "-" {
		state.Data["guest_name"] = &guestName
	}

	b.createPassFromState(ctx, msg.Chat.ID, msg.From.ID, state)
	b.clearState(msg.From.ID)
}

func (b *Bot) createPassFromState(ctx context.Context, chatID int64, userID int64, state *UserState) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	var carPlate *string
	isPedestrian, _ := state.Data["is_pedestrian"].(bool)
	if !isPedestrian {
		carPlateStr, ok := state.Data["car_plate"].(string)
		if !ok || carPlateStr == "" {
			b.sendMessage(ctx, chatID, "Ошибка: номер автомобиля не указан")
			return
		}
		carPlate = &carPlateStr
	}

	now := time.Now().UTC()
	var validTo time.Time

	var duration time.Duration
	if d, ok := state.Data["duration"].(time.Duration); ok {
		duration = d
	} else if dNs, ok := state.Data["duration"].(float64); ok {
		duration = time.Duration(dNs)
	} else if dNs, ok := state.Data["duration"].(int64); ok {
		duration = time.Duration(dNs)
	}

	if duration > 0 {
		validTo = now.Add(duration)
	} else if validToTime, ok := state.Data["valid_to"].(time.Time); ok {
		if validToTime.Location() != time.UTC {
			validTo = validToTime.UTC()
		} else {
			validTo = validToTime
		}
	} else if validToStr, ok := state.Data["valid_to"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, validToStr); err == nil {
			validTo = parsedTime.UTC()
		} else {
			b.sendMessage(ctx, chatID, "Ошибка: время действия не указано")
			b.logger.Error("failed to parse valid_to", zap.String("valid_to_str", validToStr), zap.Error(err))
			return
		}
	} else {
		b.sendMessage(ctx, chatID, "Ошибка: время действия не указано")
		b.logger.Error("duration not found in state", zap.Any("state_data", state.Data))
		return
	}

	var guestName *string
	if gn, ok := state.Data["guest_name"].(*string); ok {
		guestName = gn
	}

	validFromUTC := now.UTC()
	validToUTC := validTo.UTC()

	req := domain.CreatePassRequest{
		ApartmentID: resident.ApartmentID,
		ResidentID:  &resident.ID,
		CarPlate:    carPlate,
		GuestName:   guestName,
		ValidFrom:   validFromUTC,
		ValidTo:     validToUTC,
	}

	pass, err := b.passService.CreatePass(ctx, req)
	if err != nil {
		b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при создании пропуска: %s", err.Error()))
		b.logger.Error("failed to create pass", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	qrPNG, err := b.qrGen.GenerateQR(ctx, pass.ID)
	if err != nil {
		b.sendMessage(ctx, chatID, fmt.Sprintf("Пропуск создан, но не удалось сгенерировать QR: %s", err.Error()))
		b.logger.Error("failed to generate QR", zap.Error(err), zap.String("pass_id", pass.ID.String()))
		return
	}

	var caption string
	if pass.CarPlate != nil {
		caption = fmt.Sprintf(
			"✅ Пропуск создан!\n\n"+
				"Тип: Автомобиль\n"+
				"Номер авто: %s\n"+
				"Действует до: %s\n"+
				"ID пропуска: %s",
			*pass.CarPlate,
			b.formatLocalTime(pass.ValidTo),
			pass.ID.String(),
		)
	} else {
		caption = fmt.Sprintf(
			"✅ Пропуск создан!\n\n"+
				"Тип: Пеший гость\n"+
				"Действует до: %s\n"+
				"ID пропуска: %s",
			b.formatLocalTime(pass.ValidTo),
			pass.ID.String(),
		)
	}
	if pass.GuestName != nil && *pass.GuestName != "" {
		caption = fmt.Sprintf("%s\nГость: %s", caption, *pass.GuestName)
	}

	err = b.sendPhoto(ctx, chatID, qrPNG, caption)
	if err != nil {
		b.logger.Error("failed to send photo", zap.Error(err))
	}
}

func (b *Bot) listActivePasses(ctx context.Context, chatID int64, userID int64) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	passes, err := b.passService.GetActivePassesByResident(ctx, resident.ID)
	if err != nil {
		b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при получении пропусков: %s", err.Error()))
		b.logger.Error("failed to get active passes", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	if len(passes) == 0 {
		b.sendMessage(ctx, chatID, "У вас нет активных пропусков")
		return
	}

	text := "Ваши активные пропуска:\n\n"
	for i, pass := range passes {
		guestName := ""
		if pass.GuestName != nil {
			guestName = fmt.Sprintf(" (%s)", *pass.GuestName)
		}

		var passType, identifier string
		if pass.CarPlate != nil {
			passType = "🚗"
			identifier = *pass.CarPlate
		} else {
			passType = "🚶"
			identifier = "Пеший гость"
		}

		text += fmt.Sprintf("%d. %s %s%s\n   Действует до: %s\n   ID: %s\n\n",
			i+1,
			passType,
			identifier,
			guestName,
			b.formatLocalTime(pass.ValidTo),
			pass.ID.String()[:8],
		)
	}

	b.sendMessage(ctx, chatID, text)
}

func (b *Bot) formatLocalTime(t time.Time) string {
	return t.In(b.location).Format("15:04 02.01.2006")
}

func (b *Bot) showPassesForRevoke(ctx context.Context, chatID int64, userID int64) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	passes, err := b.passService.GetActivePassesByResident(ctx, resident.ID)
	if err != nil {
		b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при получении пропусков: %s", err.Error()))
		b.logger.Error("failed to get active passes for revoke", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	if len(passes) == 0 {
		b.sendMessage(ctx, chatID, "У вас нет активных пропусков для отзыва")
		return
	}

	var keyboardRows [][]map[string]interface{}

	text := "Выберите пропуск для отзыва:\n\n"
	for i, pass := range passes {
		guestName := ""
		if pass.GuestName != nil {
			guestName = fmt.Sprintf(" (%s)", *pass.GuestName)
		}

		var passType, identifier string
		if pass.CarPlate != nil {
			passType = "🚗"
			identifier = *pass.CarPlate
		} else {
			passType = "🚶"
			identifier = "Пеший гость"
		}

		text += fmt.Sprintf("%d. %s %s%s\n   Действует до: %s\n\n",
			i+1,
			passType,
			identifier,
			guestName,
			b.formatLocalTime(pass.ValidTo),
		)

		buttonText := fmt.Sprintf("%s %s", passType, identifier)
		if len(buttonText) > 64 {
			buttonText = buttonText[:61] + "..."
		}
		keyboardRows = append(keyboardRows, []map[string]interface{}{
			{"text": buttonText, "callback_data": fmt.Sprintf("revoke_pass_%s", pass.ID.String())},
		})
	}

	keyboard := map[string]interface{}{
		"inline_keyboard": keyboardRows,
	}

	b.sendMessageWithKeyboard(ctx, chatID, text, keyboard)
}

func (b *Bot) revokePass(ctx context.Context, chatID int64, userID int64, passID uuid.UUID) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	activePasses, err := b.passService.GetActivePassesByResident(ctx, resident.ID)
	if err != nil {
		b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при проверке пропуска: %s", err.Error()))
		return
	}

	var passInfo string
	found := false
	for _, p := range activePasses {
		if p.ID == passID {
			found = true
			if p.CarPlate != nil {
				passInfo = fmt.Sprintf("🚗 %s", *p.CarPlate)
			} else {
				passInfo = "🚶 Пеший гость"
			}
			if p.GuestName != nil {
				passInfo += fmt.Sprintf(" (%s)", *p.GuestName)
			}
			break
		}
	}

	if !found {
		b.sendMessage(ctx, chatID, "Ошибка: пропуск не найден или не принадлежит вам")
		return
	}

	err = b.passService.RevokePass(ctx, passID, 0)
	if err != nil {
		b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при отзыве пропуска: %s", err.Error()))
		b.logger.Error("failed to revoke pass", zap.Error(err), zap.String("pass_id", passID.String()), zap.Int64("user_id", userID))
		return
	}

	b.sendMessage(ctx, chatID, fmt.Sprintf("✅ Пропуск отозван:\n%s\n\nID: %s", passInfo, passID.String()[:8]))
}

func (b *Bot) getState(userID int64) *UserState {
	key := fmt.Sprintf("bot_state:%d", userID)
	stateJSON, err := b.redis.Get(context.Background(), key)
	if err == nil && stateJSON != "" {
		var state UserState
		if json.Unmarshal([]byte(stateJSON), &state) == nil {
			if time.Now().Before(state.ExpiresAt) {
				if durationNs, ok := state.Data["duration"].(float64); ok {
					state.Data["duration"] = time.Duration(durationNs)
				}
				if validToStr, ok := state.Data["valid_to"].(string); ok {
					if validToTime, err := time.Parse(time.RFC3339, validToStr); err == nil {
						state.Data["valid_to"] = validToTime
					}
				}
				return &state
			}
		}
	}

	state, exists := b.states[userID]
	if !exists {
		return nil
	}

	if time.Now().After(state.ExpiresAt) {
		delete(b.states, userID)
		return nil
	}

	return state
}

func (b *Bot) setState(userID int64, state *UserState) {
	key := fmt.Sprintf("bot_state:%d", userID)

	stateCopy := *state
	stateCopyData := make(map[string]interface{})
	for k, v := range state.Data {
		if duration, ok := v.(time.Duration); ok {
			stateCopyData[k] = int64(duration)
		} else if validToTime, ok := v.(time.Time); ok {
			stateCopyData[k] = validToTime.Format(time.RFC3339)
		} else {
			stateCopyData[k] = v
		}
	}
	stateCopy.Data = stateCopyData

	stateJSON, _ := json.Marshal(stateCopy)
	b.redis.Set(context.Background(), key, stateJSON, 10*time.Minute)

	b.states[userID] = state
}

func (b *Bot) clearState(userID int64) {
	key := fmt.Sprintf("bot_state:%d", userID)
	b.redis.Delete(context.Background(), key)
	delete(b.states, userID)
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string) error {
	return b.sendMessageWithKeyboard(ctx, chatID, text, nil)
}

func (b *Bot) sendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard interface{}) error {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}

	return b.callAPI(ctx, "sendMessage", payload)
}

func (b *Bot) sendPhoto(ctx context.Context, chatID int64, photo []byte, caption string) error {
	url := fmt.Sprintf("%s/sendPhoto", b.apiURL)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	writer.WriteField("caption", caption)

	part, err := writer.CreateFormFile("photo", "qr.png")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(photo); err != nil {
		return fmt.Errorf("failed to write photo: %w", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (b *Bot) answerCallbackQuery(ctx context.Context, callbackQueryID string, text string) error {
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
	}
	if text != "" {
		payload["text"] = text
	}

	return b.callAPI(ctx, "answerCallbackQuery", payload)
}

func (b *Bot) GetUpdates(ctx context.Context, offset int64) ([]Update, error) {
	payload := map[string]interface{}{
		"offset":  offset,
		"timeout": 10,
	}

	url := fmt.Sprintf("%s/getUpdates", b.apiURL)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telegram API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram API returned not ok")
	}

	return result.Result, nil
}

func (b *Bot) SetMyCommands(ctx context.Context) error {
	commands := []map[string]string{
		{"command": "start", "description": "Главное меню"},
		{"command": "create", "description": "Выдать пропуск гостю"},
		{"command": "list", "description": "Мои активные пропуска"},
		{"command": "revoke", "description": "Отозвать пропуск"},
	}

	payload := map[string]interface{}{
		"commands": commands,
	}

	return b.callAPI(ctx, "setMyCommands", payload)
}

func (b *Bot) callAPI(ctx context.Context, method string, payload map[string]interface{}) error {
	url := fmt.Sprintf("%s/%s", b.apiURL, method)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
