package service

import (
	"context"
	"errors"
	"testing"

	"yardpass/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockEmailSender struct{ mock.Mock }

func (m *mockEmailSender) Send(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func validPurchaseRequest(buildingName, email string) domain.PurchaseSubscriptionRequest {
	return domain.PurchaseSubscriptionRequest{
		Email:          email,
		BuildingName:   buildingName,
		ApartmentCount: 100,
		CardNumber:     "4111111111111111",
		CardHolder:     "IVAN PETROV",
		Expiry:         "12/30",
		CVV:            "123",
	}
}

func TestSubscriptionService_Purchase_RejectsDuplicateBuilding(t *testing.T) {
	t.Run("reject when building already exists by name lookup", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		emailer := new(mockEmailSender)
		ctx := context.Background()

		buildingRepo.On("List", ctx).Return([]domain.Building{
			{ID: 1, Name: "ЖК Солнечный", ApartmentCount: 50},
		}, nil)

		svc := NewSubscriptionService(buildingRepo, ruleRepo, userRepo, emailer)
		req := validPurchaseRequest("  жк солнечный  ", "Admin@Example.com")

		resp, err := svc.Purchase(ctx, req)
		assert.Nil(t, resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "уже зарегистрировано")

		buildingRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		emailer.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
		userRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("reject when DB returns unique-violation race", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		emailer := new(mockEmailSender)
		ctx := context.Background()

		buildingRepo.On("List", ctx).Return([]domain.Building{}, nil)
		buildingRepo.On("Create", ctx, mock.AnythingOfType("*domain.Building")).
			Return(&pgconn.PgError{Code: "23505", ConstraintName: "idx_buildings_name_ci"})

		svc := NewSubscriptionService(buildingRepo, ruleRepo, userRepo, emailer)
		req := validPurchaseRequest("ЖК Новый", "admin@example.com")

		resp, err := svc.Purchase(ctx, req)
		assert.Nil(t, resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "уже зарегистрировано")
	})

	t.Run("non-unique DB error surfaces a generic message", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		emailer := new(mockEmailSender)
		ctx := context.Background()

		buildingRepo.On("List", ctx).Return([]domain.Building{}, nil)
		buildingRepo.On("Create", ctx, mock.AnythingOfType("*domain.Building")).
			Return(errors.New("boom"))

		svc := NewSubscriptionService(buildingRepo, ruleRepo, userRepo, emailer)
		req := validPurchaseRequest("ЖК Новый", "admin@example.com")

		resp, err := svc.Purchase(ctx, req)
		assert.Nil(t, resp)
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "уже зарегистрировано")
		assert.Contains(t, err.Error(), "Не удалось создать здание")
	})
}

// Same email may register multiple buildings — one operator onboards several
// JKs, and each registration yields its own admin/guard pair tied to a new building.
func TestSubscriptionService_Purchase_AllowsSameEmailForDifferentBuilding(t *testing.T) {
	userRepo := new(mockUserRepo)
	buildingRepo := new(mockBuildingRepo)
	ruleRepo := new(mockRuleRepo)
	emailer := new(mockEmailSender)
	ctx := context.Background()

	buildingRepo.On("List", ctx).Return([]domain.Building{
		{ID: 1, Name: "ЖК Солнечный", ApartmentCount: 50},
	}, nil)
	buildingRepo.On("Create", ctx, mock.AnythingOfType("*domain.Building")).
		Run(func(args mock.Arguments) {
			b := args.Get(1).(*domain.Building)
			b.ID = 2
		}).Return(nil)
	ruleRepo.On("GetByBuildingID", ctx, int64(2)).Return(nil, nil)
	ruleRepo.On("Create", ctx, mock.AnythingOfType("*domain.Rule")).Return(nil)

	// Username uniqueness check used by createUserForRole — every random suffix is free.
	userRepo.On("GetByUsername", ctx, mock.AnythingOfType("string")).Return(nil, nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	emailer.On("Send", "admin@example.com", mock.Anything, mock.Anything).Return(nil)

	svc := NewSubscriptionService(buildingRepo, ruleRepo, userRepo, emailer)
	req := validPurchaseRequest("ЖК Лунный", "admin@example.com")

	resp, err := svc.Purchase(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(2), resp.BuildingID)
	assert.Equal(t, "ЖК Лунный", resp.BuildingName)
	assert.Len(t, resp.Accounts, 2)

	// Two user accounts must be created (admin + guard) — same email, different usernames.
	userRepo.AssertNumberOfCalls(t, "Create", 2)
	// No email-uniqueness probing — the constraint was lifted.
	userRepo.AssertNotCalled(t, "GetByNormalizedEmail", mock.Anything, mock.Anything)
}
