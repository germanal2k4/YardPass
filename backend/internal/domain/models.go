package domain

import (
	"time"

	"github.com/google/uuid"
)

type Building struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Address        *string   `json:"address,omitempty"`
	ApartmentCount int32     `json:"apartment_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Apartment struct {
	ID         int64     `json:"id"`
	BuildingID int64     `json:"building_id"`
	Number     string    `json:"number"`
	Floor      *int32    `json:"floor,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Resident struct {
	ID              int64     `json:"id"`
	ApartmentID     int64     `json:"apartment_id"`
	ApartmentNumber *string   `json:"apartment_number,omitempty"`
	TelegramID      int64     `json:"telegram_id"`
	ChatID          int64     `json:"chat_id"`
	Name            *string   `json:"name,omitempty"`
	Phone           *string   `json:"phone,omitempty"`
	CarPlate        *string   `json:"car_plate,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Pass struct {
	ID          uuid.UUID `json:"id"`
	ApartmentID int64     `json:"apartment_id"`
	ResidentID  *int64    `json:"resident_id,omitempty"`
	CarPlate    *string   `json:"car_plate,omitempty"`
	GuestName   *string   `json:"guest_name,omitempty"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ScanEvent struct {
	ID          int64     `json:"id"`
	PassID      uuid.UUID `json:"pass_id"`
	GuardUserID int64     `json:"guard_user_id"`
	ScannedAt   time.Time `json:"scanned_at"`
	Result      string    `json:"result"`
	Reason      *string   `json:"reason,omitempty"`
	Meta        []byte    `json:"meta,omitempty"`
}

type Rule struct {
	ID                         int64     `json:"id"`
	BuildingID                 int64     `json:"building_id"`
	QuietHoursStart            *string   `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd              *string   `json:"quiet_hours_end,omitempty"`
	DailyPassLimitPerApartment int32     `json:"daily_pass_limit_per_apartment"`
	MaxPassDurationHours       int32     `json:"max_pass_duration_hours"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

type User struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	Email           *string   `json:"email,omitempty"`
	PasswordHash    string    `json:"-"`
	Role            string    `json:"role"`
	BuildingID      *int64    `json:"building_id,omitempty"`
	ApartmentNumber *int32    `json:"apartment_number,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PassValidationResult struct {
	Valid     bool       `json:"valid"`
	Reason    string     `json:"reason,omitempty"`
	Pass      *Pass      `json:"pass,omitempty"`
	CarPlate  string     `json:"car_plate,omitempty"`
	Apartment string     `json:"apartment,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

// RegisterUserRequest is the request payload for user registration.
type RegisterUserRequest struct {
	Username        string  `json:"username" binding:"required"`
	Email           *string `json:"email,omitempty"`
	Password        string  `json:"password" binding:"required"`
	Role            string  `json:"role" binding:"required"`
	BuildingID      *int64  `json:"building_id,omitempty"`
	BuildingName    *string `json:"building_name,omitempty"`
	ApartmentNumber *int32  `json:"apartment_number,omitempty"`
}

type PurchaseSubscriptionRequest struct {
	Email          string `json:"email" binding:"required,email"`
	BuildingName   string `json:"building_name" binding:"required"`
	ApartmentCount int32  `json:"apartment_count" binding:"required"`
	CardNumber     string `json:"card_number" binding:"required"`
	CardHolder     string `json:"card_holder" binding:"required"`
	Expiry         string `json:"expiry" binding:"required"`
	CVV            string `json:"cvv" binding:"required"`
}

type AccountCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PurchaseSubscriptionResponse struct {
	BuildingID      int64                `json:"building_id"`
	BuildingName    string               `json:"building_name"`
	ApartmentCount  int32                `json:"apartment_count"`
	SubscriptionFee int64                `json:"subscription_fee"`
	Period          string               `json:"period"`
	Email           string               `json:"email"`
	Accounts        []AccountCredentials `json:"accounts"`
	Message         string               `json:"message"`
}

// CreateResidentRequest is the request payload for resident creation.
type CreateResidentRequest struct {
	ApartmentID     *int64  `json:"apartment_id,omitempty"`
	ApartmentNumber *string `json:"apartment_number,omitempty"`
	BuildingID      *int64  `json:"building_id,omitempty"`
	TelegramID      int64   `json:"telegram_id" binding:"required"`
	ChatID          *int64  `json:"chat_id,omitempty"`
	Name            *string `json:"name,omitempty"`
	Phone           *string `json:"phone,omitempty"`
}

// BulkCreateError represents an error during bulk creation.
type BulkCreateError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}
