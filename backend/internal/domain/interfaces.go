package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BuildingRepository interface {
	GetByID(ctx context.Context, id int64) (*Building, error)
	List(ctx context.Context) ([]*Building, error)
}

type ApartmentRepository interface {
	GetByID(ctx context.Context, id int64) (*Apartment, error)
	GetByBuildingID(ctx context.Context, buildingID int64) ([]*Apartment, error)
	GetByResidentTelegramID(ctx context.Context, telegramID int64) (*Apartment, error)
}

type ResidentRepository interface {
	GetByID(ctx context.Context, id int64) (*Resident, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*Resident, error)
	Create(ctx context.Context, resident *Resident) error
	Update(ctx context.Context, resident *Resident) error
	Delete(ctx context.Context, id int64) error
	BulkCreate(ctx context.Context, residents []*Resident) error
	List(ctx context.Context, filters ResidentFilters) ([]*Resident, error)
}

type ResidentFilters struct {
	ApartmentID *int64
	BuildingID  *int64
	Status      *string
	Limit       int
	Offset      int
}

type PassRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Pass, error)
	GetByApartmentID(ctx context.Context, apartmentID int64, status string) ([]*Pass, error)
	GetActiveByApartmentID(ctx context.Context, apartmentID int64) ([]*Pass, error)
	GetActiveByResidentID(ctx context.Context, residentID int64) ([]*Pass, error)
	GetActiveByBuildingID(ctx context.Context, buildingID int64) ([]*Pass, error)
	GetActiveByCarPlate(ctx context.Context, normalizedCarPlate string, buildingID *int64) (*Pass, error)
	SearchByCarPlate(ctx context.Context, carPlate string, buildingID *int64, limit int) ([]*Pass, error)
	CountActiveTodayByApartmentID(ctx context.Context, apartmentID int64) (int, error)
	CountActiveTodayByResidentID(ctx context.Context, residentID int64) (int, error)
	Create(ctx context.Context, pass *Pass) error
	Update(ctx context.Context, pass *Pass) error
	Revoke(ctx context.Context, id uuid.UUID) error
}

type ScanEventRepository interface {
	Create(ctx context.Context, event *ScanEvent) error
	List(ctx context.Context, filters ScanEventFilters) ([]*ScanEvent, error)
	CountValidScansToday(ctx context.Context) (int, error)
	GetEventsWithDetails(ctx context.Context, filters ScanEventFilters, buildingID *int64) ([]*ScanEventWithDetails, error)
	GetStatistics(ctx context.Context, from *time.Time, to *time.Time, buildingID *int64) (*Statistics, error)
}

type Statistics struct {
	TotalScans   int
	ValidScans   int
	InvalidScans int
	UniquePasses int
	UniqueGuards int
}

type ScanEventWithDetails struct {
	ID              int64
	PassID          uuid.UUID
	GuardUserID     int64
	GuardUsername   string
	ScannedAt       time.Time
	Result          string
	Reason          *string
	Meta            *string
	CarPlate        string
	ApartmentNumber string
	BuildingID      int64
}

type ScanEventFilters struct {
	PassID      *uuid.UUID
	GuardUserID *int64
	Result      *string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

type RuleRepository interface {
	GetByBuildingID(ctx context.Context, buildingID int64) (*Rule, error)
	Create(ctx context.Context, rule *Rule) error
	Update(ctx context.Context, rule *Rule) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	List(ctx context.Context, filters UserFilters) ([]*User, error)
}

type UserFilters struct {
	Role       *string
	BuildingID *int64
	Status     *string
	Limit      int
	Offset     int
}

type CreatePassRequest struct {
	ApartmentID int64
	ResidentID  *int64
	CarPlate    *string
	GuestName   *string
	ValidFrom   time.Time
	ValidTo     time.Time
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type TokenClaims struct {
	UserID     int64
	Role       string
	BuildingID *int64
	Type       string
}
