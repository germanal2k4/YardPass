package service

import (
	"context"
	"testing"

	"yardpass/internal/domain"
	"yardpass/internal/observability/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockUserRepo struct{ mock.Mock }
type mockBuildingRepo struct{ mock.Mock }
type mockRuleRepo struct{ mock.Mock }

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) List(ctx context.Context, filters domain.UserFilters) ([]domain.User, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *mockBuildingRepo) GetByID(ctx context.Context, id int64) (*domain.Building, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Building), args.Error(1)
}

func (m *mockBuildingRepo) List(ctx context.Context) ([]domain.Building, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Building), args.Error(1)
}

func (m *mockBuildingRepo) Create(ctx context.Context, building *domain.Building) error {
	args := m.Called(ctx, building)
	return args.Error(0)
}

func (m *mockBuildingRepo) UpdateApartmentCount(ctx context.Context, id int64, apartmentCount int32) (*domain.Building, error) {
	args := m.Called(ctx, id, apartmentCount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Building), args.Error(1)
}

func (m *mockRuleRepo) GetByBuildingID(ctx context.Context, buildingID int64) (*domain.Rule, error) {
	args := m.Called(ctx, buildingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *mockRuleRepo) Create(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockRuleRepo) Update(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func TestUserService_RegisterUser(t *testing.T) {
	logger := zap.NewNop()
	noopMetrics := &metrics.Metrics{}

	t.Run("success by admin", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)
		ctx := context.Background()

		buildingID := int64(1)
		apartmentNumber := int32(42)
		admin := &domain.User{ID: 1, Role: "admin", BuildingID: &buildingID}
		userRepo.On("GetByID", ctx, int64(1)).Return(admin, nil)
		buildingRepo.On("GetByID", ctx, buildingID).Return(&domain.Building{ID: buildingID}, nil)
		userRepo.On("GetByUsername", ctx, "guard1").Return(nil, nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		req := domain.RegisterUserRequest{
			Username:        "guard1",
			Password:        "secret",
			Role:            "guard",
			BuildingID:      &buildingID,
			ApartmentNumber: &apartmentNumber,
		}
		user, err := svc.RegisterUser(ctx, req, 1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "guard1", user.Username)
		assert.Equal(t, "guard", user.Role)
		userRepo.AssertExpectations(t)
		buildingRepo.AssertExpectations(t)
		ruleRepo.AssertExpectations(t)
	})

	t.Run("admin cannot create for other building", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)
		ctx := context.Background()

		adminBID := int64(1)
		reqBID := int64(2)
		apartmentNumber := int32(12)
		admin := &domain.User{ID: 1, Role: "admin", BuildingID: &adminBID}
		userRepo.On("GetByID", ctx, int64(1)).Return(admin, nil)
		buildingRepo.On("GetByID", ctx, reqBID).Return(&domain.Building{ID: reqBID}, nil)

		req := domain.RegisterUserRequest{
			Username:        "guard1",
			Password:        "secret",
			Role:            "guard",
			BuildingID:      &reqBID,
			ApartmentNumber: &apartmentNumber,
		}
		user, err := svc.RegisterUser(ctx, req, 1)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "только для своего здания")
		userRepo.AssertNotCalled(t, "Create")
	})

	t.Run("cannot create superuser", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)
		ctx := context.Background()

		admin := &domain.User{ID: 1, Role: "superuser"}
		userRepo.On("GetByID", ctx, int64(1)).Return(admin, nil)

		req := domain.RegisterUserRequest{
			Username: "su",
			Password: "secret",
			Role:     "superuser",
		}
		user, err := svc.RegisterUser(ctx, req, 1)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "суперпользователя")
		userRepo.AssertNotCalled(t, "Create")
	})

	t.Run("rejects short username and password", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)
		ctx := context.Background()

		buildingID := int64(1)
		admin := &domain.User{ID: 1, Role: "admin", BuildingID: &buildingID}
		userRepo.On("GetByID", ctx, int64(1)).Return(admin, nil)
		buildingRepo.On("GetByID", ctx, buildingID).Return(&domain.Building{ID: buildingID}, nil)

		req := domain.RegisterUserRequest{
			Username:   "abc",
			Password:   "12345",
			Role:       "guard",
			BuildingID: &buildingID,
		}
		user, err := svc.RegisterUser(ctx, req, 1)
		assert.Error(t, err)
		assert.Nil(t, user)
		userRepo.AssertNotCalled(t, "Create")
	})

	t.Run("creates default rule for new building by name", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)
		ctx := context.Background()

		superuser := &domain.User{ID: 1, Role: "superuser"}
		userRepo.On("GetByID", ctx, int64(1)).Return(superuser, nil)
		buildingRepo.On("List", ctx).Return([]domain.Building{}, nil)
		buildingRepo.On("Create", ctx, mock.AnythingOfType("*domain.Building")).Run(func(args mock.Arguments) {
			b := args.Get(1).(*domain.Building)
			b.ID = 99
		}).Return(nil)
		ruleRepo.On("GetByBuildingID", ctx, int64(99)).Return(nil, nil)
		ruleRepo.On("Create", ctx, mock.AnythingOfType("*domain.Rule")).Return(nil)
		buildingRepo.On("GetByID", ctx, int64(99)).Return(&domain.Building{ID: 99}, nil)
		userRepo.On("GetByUsername", ctx, "guard1").Return(nil, nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		buildingName := "Новый дом 7"
		apartmentNumber := int32(10)
		req := domain.RegisterUserRequest{
			Username:        "guard1",
			Password:        "secret",
			Role:            "guard",
			BuildingName:    &buildingName,
			ApartmentNumber: &apartmentNumber,
		}

		user, err := svc.RegisterUser(ctx, req, 1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		ruleRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateCredentials(t *testing.T) {
	logger := zap.NewNop()
	noopMetrics := &metrics.Metrics{}
	ctx := context.Background()

	t.Run("admin updates own username and password", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)

		bID := int64(1)
		admin := &domain.User{
			ID:           1,
			Username:     "admin_old",
			Role:         "admin",
			BuildingID:   &bID,
			PasswordHash: "old_hash",
		}

		userRepo.On("GetByID", ctx, int64(1)).Return(admin, nil).Twice()
		userRepo.On("GetByUsername", ctx, "admin_new").Return(nil, nil)
		userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		updated, err := svc.UpdateCredentials(ctx, 1, 1, "admin_new", "new_password")
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "admin_new", updated.Username)
		assert.Empty(t, updated.PasswordHash)
		userRepo.AssertExpectations(t)
	})

	t.Run("admin updates guard from own building", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)

		bID := int64(7)
		admin := &domain.User{ID: 10, Role: "admin", BuildingID: &bID}
		guard := &domain.User{ID: 20, Username: "guard_1", Role: "guard", BuildingID: &bID, PasswordHash: "old_hash"}

		userRepo.On("GetByID", ctx, int64(10)).Return(admin, nil)
		userRepo.On("GetByID", ctx, int64(20)).Return(guard, nil)
		userRepo.On("GetByUsername", ctx, "guard_new").Return(nil, nil)
		userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		updated, err := svc.UpdateCredentials(ctx, 10, 20, "guard_new", "")
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "guard_new", updated.Username)
		userRepo.AssertExpectations(t)
	})

	t.Run("admin cannot update guard from another building", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)

		adminBID := int64(1)
		guardBID := int64(2)
		admin := &domain.User{ID: 10, Role: "admin", BuildingID: &adminBID}
		guard := &domain.User{ID: 20, Username: "guard_1", Role: "guard", BuildingID: &guardBID}

		userRepo.On("GetByID", ctx, int64(10)).Return(admin, nil)
		userRepo.On("GetByID", ctx, int64(20)).Return(guard, nil)

		updated, err := svc.UpdateCredentials(ctx, 10, 20, "guard_new", "")
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "своего здания")
		userRepo.AssertNotCalled(t, "Update")
	})

	t.Run("rejects short credentials on update", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		buildingRepo := new(mockBuildingRepo)
		ruleRepo := new(mockRuleRepo)
		svc := NewUserService(userRepo, buildingRepo, ruleRepo, logger, noopMetrics)

		bID := int64(1)
		admin := &domain.User{ID: 1, Role: "admin", BuildingID: &bID, Username: "admin_old"}

		userRepo.On("GetByID", ctx, int64(1)).Return(admin, nil).Twice()

		updated, err := svc.UpdateCredentials(ctx, 1, 1, "abc", "12345")
		assert.Error(t, err)
		assert.Nil(t, updated)
		userRepo.AssertNotCalled(t, "Update")
	})
}

func TestNormalizeBuildingName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"trim_cyrillic", "  ЖК Солнечный 12  ", "ЖК Солнечный 12"},
		{"preserve_double_space", "Building A  7", "Building A  7"},
		{"punctuation", "ул. Ленина, д. 5", "ул. Ленина, д. 5"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeBuildingName(tt.in))
		})
	}
}
