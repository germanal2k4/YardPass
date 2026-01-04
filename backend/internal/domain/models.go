package domain

import (
	"time"

	"github.com/google/uuid"
)

type Building struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Apartment struct {
	ID         int64     `json:"id"`
	BuildingID int64     `json:"building_id"`
	Number     string    `json:"number"`
	Floor      *int      `json:"floor,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Resident struct {
	ID          int64     `json:"id"`
	ApartmentID int64     `json:"apartment_id"`
	TelegramID  int64     `json:"telegram_id"`
	ChatID      int64     `json:"chat_id"`
	Name        *string   `json:"name,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Pass struct {
	ID          uuid.UUID `json:"id"`
	ApartmentID int64     `json:"apartment_id"`
	CarPlate    string    `json:"car_plate"`
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
	Meta        *string   `json:"meta,omitempty"`
}

type Rule struct {
	ID                         int64     `json:"id"`
	BuildingID                 int64     `json:"building_id"`
	QuietHoursStart            *string   `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd              *string   `json:"quiet_hours_end,omitempty"`
	DailyPassLimitPerApartment int       `json:"daily_pass_limit_per_apartment"`
	MaxPassDurationHours       int       `json:"max_pass_duration_hours"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        *string   `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PassValidationResult struct {
	Valid     bool       `json:"valid"`
	Reason    string     `json:"reason,omitempty"`
	Pass      *Pass      `json:"pass,omitempty"`
	CarPlate  string     `json:"car_plate,omitempty"`
	Apartment string     `json:"apartment,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}
