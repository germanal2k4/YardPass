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

	"yardpass/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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
		b.updatesTotal.WithLabelValues("message").Inc()
		b.handleMessage(ctx, *update.Message)
	} else if update.CallbackQuery != nil {
		b.updatesTotal.WithLabelValues("callback_query").Inc()
		b.handleCallbackQuery(ctx, *update.CallbackQuery)
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg Message) {
	userID := msg.From.ID
	text := msg.Text

	if text == "/start" || text == "/create" || text == "/list" || text == "/revoke" || text == "/personal" || text == "/mycar" {
		switch text {
		case "/start":
			b.commandsTotal.WithLabelValues("start").Inc()
			b.handleStart(ctx, msg)
		case "/create":
			b.commandsTotal.WithLabelValues("create").Inc()
			cb := CallbackQuery{
				ID:      "",
				From:    msg.From,
				Message: &msg,
				Data:    "create_pass",
			}
			b.handleCallbackQuery(ctx, cb)
		case "/list":
			b.commandsTotal.WithLabelValues("list").Inc()
			cb := CallbackQuery{
				ID:      "",
				From:    msg.From,
				Message: &msg,
				Data:    "list_active",
			}
			b.handleCallbackQuery(ctx, cb)
		case "/revoke":
			b.commandsTotal.WithLabelValues("revoke").Inc()
			cb := CallbackQuery{
				ID:      "",
				From:    msg.From,
				Message: &msg,
				Data:    "revoke_pass",
			}
			b.handleCallbackQuery(ctx, cb)
		case "/personal":
			b.commandsTotal.WithLabelValues("personal").Inc()
			b.sendPersonalPass(ctx, msg.Chat.ID, msg.From.ID)
		case "/mycar":
			b.commandsTotal.WithLabelValues("mycar").Inc()
			cb := CallbackQuery{
				ID:      "",
				From:    msg.From,
				Message: &msg,
				Data:    "my_car",
			}
			b.handleCallbackQuery(ctx, cb)
		}
		return
	}

	state := b.getState(userID)
	if state == nil {
		_ = b.sendMessage(ctx, msg.Chat.ID, "Используйте /start для начала работы")
		return
	}

	switch state.Step {
	case StateWaitingGuestType:
		_ = b.sendMessage(ctx, msg.Chat.ID, "Используйте кнопки для выбора типа гостя")
	case StateWaitingCarPlate:
		b.handleCarPlate(ctx, msg, state)
	case StateWaitingDuration:
		b.handleDuration(ctx, msg, state)
	case StateWaitingCustomTime:
		b.handleCustomTime(ctx, msg, state)
	case StateWaitingGuestName:
		b.handleGuestName(ctx, msg, state)
	case StateWaitingResidentCarPlate:
		b.handleResidentCarPlate(ctx, msg, state)
	default:
		_ = b.sendMessage(ctx, msg.Chat.ID, "Неизвестное состояние. Используйте /start")
		b.clearState(userID)
	}
}

func (b *Bot) handleStart(ctx context.Context, msg Message) {
	userID := msg.From.ID

	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		_ = b.sendMessage(ctx, msg.Chat.ID, "Вы не зарегистрированы как житель. Обратитесь к администратору.")
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
			{
				{"text": "🔐 Личный постоянный пропуск", "callback_data": "personal_pass"},
			},
			{
				{"text": "🚗 Моя машина (для пропуска)", "callback_data": "my_car"},
			},
		},
	}

	_ = b.sendMessageWithKeyboard(ctx, msg.Chat.ID, "Добро пожаловать в YardPass!\n\nВыберите действие:", keyboard)
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
		_ = b.sendMessageWithKeyboard(ctx, cb.Message.Chat.ID, "Выберите тип гостя:", keyboard)
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "list_active":
		b.listActivePasses(ctx, cb.Message.Chat.ID, userID)
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "revoke_pass":
		b.showPassesForRevoke(ctx, cb.Message.Chat.ID, userID)
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "personal_pass":
		b.sendPersonalPass(ctx, cb.Message.Chat.ID, userID)
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "my_car":
		b.showMyCarMenu(ctx, cb.Message.Chat.ID, userID)
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "register_car":
		b.setState(userID, &UserState{
			Step:      StateWaitingResidentCarPlate,
			Data:      make(map[string]interface{}),
			ExpiresAt: time.Now().Add(10 * time.Minute),
		})
		_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Введите номер вашего автомобиля (например: А123ВС77 или A123BC77):")
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "remove_car":
		b.removeResidentCar(ctx, cb.Message.Chat.ID, userID)
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "guest_car":
		state := b.getState(userID)
		if state == nil {
			_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Сессия истекла. Начните заново с /start")
			_ = b.answerCallbackQuery(ctx, cb.ID, "")
			return
		}
		state.Step = StateWaitingCarPlate
		b.setState(userID, state)
		_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Введите номер автомобиля (на английском, например: A123BC77):")
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "guest_pedestrian":
		state := b.getState(userID)
		if state == nil {
			_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Сессия истекла. Начните заново с /start")
			_ = b.answerCallbackQuery(ctx, cb.ID, "")
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
		_ = b.sendMessageWithKeyboard(ctx, cb.Message.Chat.ID, "Выберите срок действия пропуска:", keyboard)
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	case "duration_1h", "duration_2h", "duration_4h", "duration_custom":
		state := b.getState(userID)
		if state == nil {
			_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Сессия истекла. Начните заново с /start")
			_ = b.answerCallbackQuery(ctx, cb.ID, "")
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
			_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Введите время окончания действия пропуска в формате ЧЧ:ММ (например, 22:00):")
			_ = b.answerCallbackQuery(ctx, cb.ID, "")
			return
		}

		state.Step = StateWaitingGuestName
		b.setState(userID, state)
		_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Введите имя гостя (или отправьте '-' чтобы пропустить):")
		_ = b.answerCallbackQuery(ctx, cb.ID, "")

	default:
		if strings.HasPrefix(data, "revoke_pass_") {
			passIDStr := strings.TrimPrefix(data, "revoke_pass_")
			passID, err := uuid.Parse(passIDStr)
			if err != nil {
				_ = b.sendMessage(ctx, cb.Message.Chat.ID, "Ошибка: неверный ID пропуска")
				_ = b.answerCallbackQuery(ctx, cb.ID, "")
				return
			}
			b.revokePass(ctx, cb.Message.Chat.ID, userID, passID)
			_ = b.answerCallbackQuery(ctx, cb.ID, "")
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
	_ = b.sendMessageWithKeyboard(ctx, msg.Chat.ID, "Выберите срок действия пропуска:", keyboard)
}

func (b *Bot) handleDuration(ctx context.Context, msg Message, state *UserState) {
	_ = b.sendMessage(ctx, msg.Chat.ID, "Используйте кнопки для выбора")
}

func (b *Bot) handleCustomTime(ctx context.Context, msg Message, state *UserState) {
	timeStr := msg.Text
	now := time.Now().In(b.location)

	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		_ = b.sendMessage(ctx, msg.Chat.ID, "Неверный формат времени. Введите в формате ЧЧ:ММ (например, 22:00)")
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
	_ = b.sendMessage(ctx, msg.Chat.ID, "Введите имя гостя (или отправьте '-' чтобы пропустить):")
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
		_ = b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	var carPlate *string
	isPedestrian, _ := state.Data["is_pedestrian"].(bool)
	if !isPedestrian {
		carPlateStr, ok := state.Data["car_plate"].(string)
		if !ok || carPlateStr == "" {
			_ = b.sendMessage(ctx, chatID, "Ошибка: номер автомобиля не указан")
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
			_ = b.sendMessage(ctx, chatID, "Ошибка: время действия не указано")
			b.logger.Error("failed to parse valid_to", zap.String("valid_to_str", validToStr), zap.Error(err))
			return
		}
	} else {
		_ = b.sendMessage(ctx, chatID, "Ошибка: время действия не указано")
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
		_ = b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при создании пропуска: %s", err.Error()))
		b.logger.Error("failed to create pass", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	qrPNG, err := b.qrGen.GenerateQR(ctx, pass.ID)
	if err != nil {
		_ = b.sendMessage(ctx, chatID, fmt.Sprintf("Пропуск создан, но не удалось сгенерировать QR: %s", err.Error()))
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
		_ = b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	passes, err := b.passService.GetActivePassesByResident(ctx, resident.ID)
	if err != nil {
		_ = b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при получении пропусков: %s", err.Error()))
		b.logger.Error("failed to get active passes", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	if len(passes) == 0 {
		_ = b.sendMessage(ctx, chatID, "У вас нет активных пропусков")
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

	_ = b.sendMessage(ctx, chatID, text)
}

func (b *Bot) formatLocalTime(t time.Time) string {
	return t.In(b.location).Format("15:04 02.01.2006")
}

func (b *Bot) showPassesForRevoke(ctx context.Context, chatID int64, userID int64) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		_ = b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	passes, err := b.passService.GetActivePassesByResident(ctx, resident.ID)
	if err != nil {
		_ = b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при получении пропусков: %s", err.Error()))
		b.logger.Error("failed to get active passes for revoke", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	if len(passes) == 0 {
		_ = b.sendMessage(ctx, chatID, "У вас нет активных пропусков для отзыва")
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

	_ = b.sendMessageWithKeyboard(ctx, chatID, text, keyboard)
}

func (b *Bot) revokePass(ctx context.Context, chatID int64, userID int64, passID uuid.UUID) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		_ = b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	activePasses, err := b.passService.GetActivePassesByResident(ctx, resident.ID)
	if err != nil {
		_ = b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при проверке пропуска: %s", err.Error()))
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
		_ = b.sendMessage(ctx, chatID, "Ошибка: пропуск не найден или не принадлежит вам")
		return
	}

	err = b.passService.RevokePass(ctx, passID, 0)
	if err != nil {
		_ = b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при отзыве пропуска: %s", err.Error()))
		b.logger.Error("failed to revoke pass", zap.Error(err), zap.String("pass_id", passID.String()), zap.Int64("user_id", userID))
		return
	}

	_ = b.sendMessage(ctx, chatID, fmt.Sprintf("✅ Пропуск отозван:\n%s\n\nID: %s", passInfo, passID.String()[:8]))
}

func (b *Bot) sendPersonalPass(ctx context.Context, chatID int64, userID int64) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		_ = b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	apartment, err := b.apartmentRepo.GetByID(ctx, resident.ApartmentID)
	if err != nil || apartment == nil {
		_ = b.sendMessage(ctx, chatID, "Ошибка: квартира не найдена")
		return
	}

	token := b.passService.GenerateResidentPersonalPassToken(resident.TelegramID)
	qrPNG, err := b.qrGen.GenerateRawQR(ctx, token)
	if err != nil {
		_ = b.sendMessage(ctx, chatID, "Не удалось сгенерировать личный пропуск")
		return
	}

	caption := fmt.Sprintf(
		"🔐 Личный постоянный пропуск резидента\n\n"+
			"🏠 Квартира: %s\n",
		apartment.Number,
	)

	if resident.CarPlate != nil && *resident.CarPlate != "" {
		caption += fmt.Sprintf("🚗 Автомобиль: %s\n", *resident.CarPlate)
	} else {
		caption += "🚗 Автомобиль: не зарегистрирован\n"
	}

	if resident.Name != nil && *resident.Name != "" {
		caption += fmt.Sprintf("👤 Жилец: %s\n", *resident.Name)
	}

	caption += "\n📋 Содержимое QR-кода (токен доступа):\n" + token +
		"\n\nПропуск не имеет срока действия. Покажите охраннику для прохода."

	_ = b.sendPhoto(ctx, chatID, qrPNG, caption)
}

func (b *Bot) showMyCarMenu(ctx context.Context, chatID int64, userID int64) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		_ = b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	var text string
	var keyboard map[string]interface{}

	if resident.CarPlate != nil && *resident.CarPlate != "" {
		text = fmt.Sprintf(
			"🚗 Ваш зарегистрированный автомобиль: *%s*\n\n"+
				"Номер автомобиля отображается охраннику при проверке вашего личного QR-пропуска.",
			*resident.CarPlate,
		)
		keyboard = map[string]interface{}{
			"inline_keyboard": [][]map[string]interface{}{
				{{"text": "✏️ Изменить номер", "callback_data": "register_car"}},
				{{"text": "🗑 Удалить автомобиль", "callback_data": "remove_car"}},
			},
		}
	} else {
		text = "🚗 У вас не зарегистрирован автомобиль для постоянного пропуска.\n\n" +
			"Зарегистрируйте ваш автомобиль, и его номер будет отображаться охраннику при проверке личного QR-пропуска."
		keyboard = map[string]interface{}{
			"inline_keyboard": [][]map[string]interface{}{
				{{"text": "➕ Зарегистрировать автомобиль", "callback_data": "register_car"}},
			},
		}
	}

	_ = b.sendMessageWithKeyboard(ctx, chatID, text, keyboard)
}

func (b *Bot) handleResidentCarPlate(ctx context.Context, msg Message, state *UserState) {
	carPlateRaw := strings.TrimSpace(msg.Text)
	userID := msg.From.ID
	b.clearState(userID)

	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		_ = b.sendMessage(ctx, msg.Chat.ID, "Ошибка: житель не найден")
		return
	}

	if err := b.residentRepo.SetCarPlate(ctx, resident.ID, &carPlateRaw); err != nil {
		_ = b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Ошибка при сохранении номера: %s", err.Error()))
		b.logger.Error("failed to set resident car plate", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	_ = b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf(
		"✅ Автомобиль зарегистрирован: *%s*\n\n"+
			"Теперь при проверке вашего личного QR-пропуска охранник увидит номер автомобиля.\n\n"+
			"Получите обновлённый QR-код: /personal",
		carPlateRaw,
	))
}

func (b *Bot) removeResidentCar(ctx context.Context, chatID int64, userID int64) {
	resident, err := b.residentRepo.GetByTelegramID(ctx, userID)
	if err != nil || resident == nil {
		_ = b.sendMessage(ctx, chatID, "Ошибка: житель не найден")
		return
	}

	if err := b.residentRepo.SetCarPlate(ctx, resident.ID, nil); err != nil {
		_ = b.sendMessage(ctx, chatID, fmt.Sprintf("Ошибка при удалении автомобиля: %s", err.Error()))
		return
	}

	_ = b.sendMessage(ctx, chatID, "✅ Автомобиль удалён из вашего профиля.")
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
	_ = b.redis.Set(context.Background(), key, stateJSON, 10*time.Minute)

	b.states[userID] = state
}

func (b *Bot) clearState(userID int64) {
	key := fmt.Sprintf("bot_state:%d", userID)
	_ = b.redis.Delete(context.Background(), key)
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

	if err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		return fmt.Errorf("write chat_id field: %w", err)
	}
	if err := writer.WriteField("caption", caption); err != nil {
		return fmt.Errorf("write caption field: %w", err)
	}

	part, err := writer.CreateFormFile("photo", "qr.png")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(photo); err != nil {
		return fmt.Errorf("write photo: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			b.logger.Error("Failed to close response body", zap.Error(err))
		}
	}()

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
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.apiErrorsTotal.WithLabelValues("getUpdates", "500").Inc()
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			b.logger.Error("Failed to close response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		b.apiErrorsTotal.WithLabelValues("getUpdates", strconv.Itoa(resp.StatusCode)).Inc()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telegram API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !result.OK {
		b.apiErrorsTotal.WithLabelValues("getUpdates", "500").Inc()
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
		{"command": "personal", "description": "Личный постоянный пропуск"},
		{"command": "mycar", "description": "Мой автомобиль для постоянного пропуска"},
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
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.apiErrorsTotal.WithLabelValues(method, "500").Inc()
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			b.logger.Error("Failed to close response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		b.apiErrorsTotal.WithLabelValues(method, strconv.Itoa(resp.StatusCode)).Inc()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
