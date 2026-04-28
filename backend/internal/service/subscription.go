package service

import (
	"context"
	"crypto/rand"
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

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

type EmailSender interface {
	Send(to, subject, body string) error
}

type SubscriptionService struct {
	buildingRepo domain.BuildingRepository
	userRepo     domain.UserRepository
	emailSender  EmailSender
}

func NewSubscriptionService(
	buildingRepo domain.BuildingRepository,
	userRepo domain.UserRepository,
	emailSender EmailSender,
) *SubscriptionService {
	return &SubscriptionService{
		buildingRepo: buildingRepo,
		userRepo:     userRepo,
		emailSender:  emailSender,
	}
}

func (s *SubscriptionService) Purchase(ctx context.Context, req domain.PurchaseSubscriptionRequest) (*domain.PurchaseSubscriptionResponse, error) {
	if strings.TrimSpace(req.BuildingName) == "" {
		return nil, fmt.Errorf("building_name is required")
	}

	if err := validatePaymentFields(req); err != nil {
		return nil, err
	}

	building := &domain.Building{Name: strings.TrimSpace(req.BuildingName)}
	if err := s.buildingRepo.Create(ctx, building); err != nil {
		return nil, fmt.Errorf("create building: %w", err)
	}

	adminCreds, err := s.createUserForRole(ctx, building.ID, building.Name, req.Email, "admin")
	if err != nil {
		return nil, err
	}

	guardCreds, err := s.createUserForRole(ctx, building.ID, building.Name, req.Email, "guard")
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
		return nil, fmt.Errorf("send credentials email: %w", err)
	}

	return &domain.PurchaseSubscriptionResponse{
		BuildingID:      building.ID,
		BuildingName:    building.Name,
		SubscriptionFee: subscriptionFeeRub,
		Period:          subscriptionPeriod,
		Email:           req.Email,
		Accounts: []domain.AccountCredentials{
			adminCreds,
			guardCreds,
		},
		Message: "Subscription is paid. Credentials were sent to email.",
	}, nil
}

func (s *SubscriptionService) createUserForRole(
	ctx context.Context,
	buildingID int64,
	buildingName string,
	email string,
	role string,
) (domain.AccountCredentials, error) {
	username, err := s.generateUniqueUsername(ctx, buildingName, role)
	if err != nil {
		return domain.AccountCredentials{}, err
	}

	password, err := generatePassword(12)
	if err != nil {
		return domain.AccountCredentials{}, fmt.Errorf("generate password: %w", err)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return domain.AccountCredentials{}, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     username,
		Email:        &email,
		PasswordHash: passwordHash,
		Role:         role,
		BuildingID:   &buildingID,
		Status:       "active",
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return domain.AccountCredentials{}, fmt.Errorf("create %s user: %w", role, err)
	}

	return domain.AccountCredentials{Username: username, Password: password}, nil
}

func (s *SubscriptionService) generateUniqueUsername(ctx context.Context, buildingName, role string) (string, error) {
	base := normalizeSlug(buildingName)
	if base == "" {
		base = "building"
	}

	for range 10 {
		suffix, err := randomDigits(4)
		if err != nil {
			return "", fmt.Errorf("generate username suffix: %w", err)
		}

		username := fmt.Sprintf("%s_%s_%s", role, base, suffix)
		existing, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			return "", fmt.Errorf("check username uniqueness: %w", err)
		}
		if existing == nil {
			return username, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique username")
}

func validatePaymentFields(req domain.PurchaseSubscriptionRequest) error {
	cardDigits := onlyDigits(req.CardNumber)
	if len(cardDigits) < 12 || len(cardDigits) > 19 {
		return fmt.Errorf("invalid card number")
	}
	if len(strings.TrimSpace(req.CardHolder)) < 3 {
		return fmt.Errorf("invalid card holder")
	}
	if matched := regexp.MustCompile(`^(0[1-9]|1[0-2])\/\d{2}$`).MatchString(strings.TrimSpace(req.Expiry)); !matched {
		return fmt.Errorf("invalid expiry")
	}
	cvvDigits := onlyDigits(req.CVV)
	if len(cvvDigits) < 3 || len(cvvDigits) > 4 {
		return fmt.Errorf("invalid cvv")
	}

	return nil
}

func generatePassword(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	if length <= 0 {
		return "", fmt.Errorf("invalid password length")
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

func randomDigits(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid random digits length")
	}
	var out strings.Builder
	out.Grow(length)
	for range length {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		out.WriteByte(byte('0' + n.Int64()))
	}
	return out.String(), nil
}

func normalizeSlug(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	normalized := nonAlnum.ReplaceAllString(lower, "_")
	normalized = strings.Trim(normalized, "_")
	if len(normalized) > 20 {
		return normalized[:20]
	}
	return normalized
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
