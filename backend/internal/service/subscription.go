package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"yardpass/internal/auth"
	"yardpass/internal/domain"
)

const (
	subscriptionFeeRub = int64(200000)
	subscriptionPeriod = "1 year"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type SubscriptionService struct {
	buildingRepo domain.BuildingRepository
	ruleRepo     domain.RuleRepository
	userRepo     domain.UserRepository
	emailSender  EmailSender
}

func NewSubscriptionService(
	buildingRepo domain.BuildingRepository,
	ruleRepo domain.RuleRepository,
	userRepo domain.UserRepository,
	emailSender EmailSender,
) *SubscriptionService {
	return &SubscriptionService{
		buildingRepo: buildingRepo,
		ruleRepo:     ruleRepo,
		userRepo:     userRepo,
		emailSender:  emailSender,
	}
}

func (s *SubscriptionService) Purchase(ctx context.Context, req domain.PurchaseSubscriptionRequest) (*domain.PurchaseSubscriptionResponse, error) {
	if strings.TrimSpace(req.BuildingName) == "" {
		return nil, errors.New("Укажите название здания (building_name).")
	}
	if req.ApartmentCount <= 0 {
		return nil, errors.New("Количество квартир (apartment_count) должно быть больше нуля.")
	}

	if err := validatePaymentFields(req); err != nil {
		return nil, err
	}

	emailNorm := strings.ToLower(strings.TrimSpace(req.Email))
	if emailNorm == "" {
		return nil, errors.New("Укажите корректный email.")
	}
	emailTaken, err := s.userRepo.GetByNormalizedEmail(ctx, emailNorm)
	if err != nil {
		return nil, errors.New("Не удалось проверить email. Попробуйте позже.")
	}
	if emailTaken != nil {
		return nil, errors.New("Указанный email уже зарегистрирован.")
	}

	normalizedBuildingName := normalizeBuildingName(req.BuildingName)
	if normalizedBuildingName == "" {
		return nil, errors.New("Укажите корректное название здания.")
	}

	building, err := s.findOrCreateBuilding(ctx, normalizedBuildingName, req.ApartmentCount)
	if err != nil {
		return nil, err
	}

	adminCreds, err := s.createUserForRole(ctx, building.ID, req.Email, "admin")
	if err != nil {
		return nil, err
	}

	guardCreds, err := s.createUserForRole(ctx, building.ID, req.Email, "guard")
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf(
		"Оплата подписки YardPass на сумму %d RUB успешно подтверждена.\n\nЗдание: %s\n\nДоступы:\n1) Администратор\n   Логин: %s\n   Пароль: %s\n\n2) Охранник\n   Логин: %s\n   Пароль: %s\n",
		subscriptionFeeRub,
		building.Name,
		adminCreds.Username,
		adminCreds.Password,
		guardCreds.Username,
		guardCreds.Password,
	)

	if err := s.emailSender.Send(req.Email, "YardPass: доступы к системе", message); err != nil {
		return nil, errors.New("Не удалось отправить письмо с доступами. Проверьте email и попробуйте позже.")
	}

	return &domain.PurchaseSubscriptionResponse{
		BuildingID:      building.ID,
		BuildingName:    building.Name,
		ApartmentCount:  building.ApartmentCount,
		SubscriptionFee: subscriptionFeeRub,
		Period:          subscriptionPeriod,
		Email:           req.Email,
		Accounts: []domain.AccountCredentials{
			adminCreds,
			guardCreds,
		},
		Message: "Подписка оплачена. Данные для входа отправлены на указанный email.",
	}, nil
}

func (s *SubscriptionService) createUserForRole(
	ctx context.Context,
	buildingID int64,
	email string,
	role string,
) (domain.AccountCredentials, error) {
	username, err := s.generateUniqueUsername(ctx, role)
	if err != nil {
		return domain.AccountCredentials{}, err
	}

	password, err := generatePassword(12)
	if err != nil {
		return domain.AccountCredentials{}, errors.New("Не удалось сформировать пароль. Попробуйте позже.")
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return domain.AccountCredentials{}, errors.New("Не удалось сохранить пароль. Попробуйте позже.")
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		BuildingID:   &buildingID,
		Status:       "active",
	}
	if role == "admin" {
		user.Email = &email
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return domain.AccountCredentials{}, fmt.Errorf("Не удалось создать учётную запись (%s). Попробуйте позже.", role)
	}

	return domain.AccountCredentials{Username: username, Password: password}, nil
}

func (s *SubscriptionService) generateUniqueUsername(ctx context.Context, role string) (string, error) {
	for range 24 {
		suffix, err := randomHexSuffix(6)
		if err != nil {
			return "", errors.New("Не удалось сформировать логин. Попробуйте позже.")
		}

		username := fmt.Sprintf("%s_%s", role, suffix)
		existing, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			return "", errors.New("Не удалось проверить уникальность логина. Попробуйте позже.")
		}
		if existing == nil {
			return username, nil
		}
	}

	return "", errors.New("Не удалось подобрать уникальный логин. Попробуйте позже.")
}

func validatePaymentFields(req domain.PurchaseSubscriptionRequest) error {
	cardDigits := onlyDigits(req.CardNumber)
	if len(cardDigits) < 12 || len(cardDigits) > 19 {
		return errors.New("Некорректный номер карты.")
	}
	if len(strings.TrimSpace(req.CardHolder)) < 3 {
		return errors.New("Укажите имя держателя карты (не менее 3 символов).")
	}
	if matched := regexp.MustCompile(`^(0[1-9]|1[0-2])\/\d{2}$`).MatchString(strings.TrimSpace(req.Expiry)); !matched {
		return errors.New("Некорректный срок действия карты (формат ММ/ГГ).")
	}
	cvvDigits := onlyDigits(req.CVV)
	if len(cvvDigits) < 3 || len(cvvDigits) > 4 {
		return errors.New("Некорректный CVV-код.")
	}

	return nil
}

func generatePassword(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	if length <= 0 {
		return "", errors.New("внутренняя ошибка: некорректная длина пароля")
	}
	var out strings.Builder
	out.Grow(length)
	max := big.NewInt(int64(len(alphabet)))
	for range length {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out.WriteByte(alphabet[n.Int64()])
	}
	return out.String(), nil
}

// randomHexSuffix returns a lowercase hex string of exactly hexLen characters (hexLen must be even).
func randomHexSuffix(hexLen int) (string, error) {
	if hexLen <= 0 || hexLen%2 != 0 {
		return "", errors.New("внутренняя ошибка: некорректная длина hex-суффикса")
	}
	buf := make([]byte, hexLen/2)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *SubscriptionService) findOrCreateBuilding(ctx context.Context, normalizedName string, apartmentCount int32) (*domain.Building, error) {
	buildings, err := s.buildingRepo.List(ctx)
	if err != nil {
		return nil, errors.New("Не удалось загрузить список зданий. Попробуйте позже.")
	}

	for _, existing := range buildings {
		if strings.EqualFold(normalizeBuildingName(existing.Name), normalizedName) {
			if apartmentCount > existing.ApartmentCount {
				updated, err := s.buildingRepo.UpdateApartmentCount(ctx, existing.ID, apartmentCount)
				if err != nil {
					return nil, errors.New("Не удалось обновить количество квартир. Попробуйте позже.")
				}
				if updated != nil {
					return updated, nil
				}
			}
			copy := existing
			return &copy, nil
		}
	}

	building := &domain.Building{
		Name:           normalizedName,
		ApartmentCount: apartmentCount,
	}
	if err := s.buildingRepo.Create(ctx, building); err != nil {
		if !isUniqueViolation(err) {
			return nil, errors.New("Не удалось создать здание. Попробуйте позже.")
		}
		buildings2, err2 := s.buildingRepo.List(ctx)
		if err2 != nil {
			return nil, errors.New("Не удалось загрузить список зданий. Попробуйте позже.")
		}
		for _, existing := range buildings2 {
			if strings.EqualFold(normalizeBuildingName(existing.Name), normalizedName) {
				if apartmentCount > existing.ApartmentCount {
					updated, err3 := s.buildingRepo.UpdateApartmentCount(ctx, existing.ID, apartmentCount)
					if err3 != nil {
						return nil, errors.New("Не удалось обновить количество квартир. Попробуйте позже.")
					}
					if updated != nil {
						return updated, nil
					}
				}
				copyB := existing
				return &copyB, nil
			}
		}
		return nil, errors.New("Не удалось создать здание. Попробуйте позже.")
	}
	if err := s.ensureDefaultRule(ctx, building.ID); err != nil {
		return nil, err
	}
	return building, nil
}

func (s *SubscriptionService) ensureDefaultRule(ctx context.Context, buildingID int64) error {
	existingRule, err := s.ruleRepo.GetByBuildingID(ctx, buildingID)
	if err != nil {
		return errors.New("Не удалось загрузить правила здания. Попробуйте позже.")
	}
	if existingRule != nil {
		return nil
	}

	rule := &domain.Rule{
		BuildingID:                 buildingID,
		DailyPassLimitPerApartment: 5,
		MaxPassDurationHours:       24,
	}
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return errors.New("Не удалось создать правила по умолчанию. Попробуйте позже.")
	}
	return nil
}

func onlyDigits(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
