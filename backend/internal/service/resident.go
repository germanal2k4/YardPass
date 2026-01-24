package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"yardpass/internal/domain"

	"go.uber.org/zap"
)

type ResidentService struct {
	residentRepo  domain.ResidentRepository
	apartmentRepo domain.ApartmentRepository
	logger        *zap.Logger
}

func NewResidentService(residentRepo domain.ResidentRepository, apartmentRepo domain.ApartmentRepository, logger *zap.Logger) *ResidentService {
	return &ResidentService{
		residentRepo:  residentRepo,
		apartmentRepo: apartmentRepo,
		logger:        logger,
	}
}

func (s *ResidentService) CreateResident(ctx context.Context, req domain.CreateResidentRequest) (*domain.Resident, error) {
	apartment, err := s.apartmentRepo.GetByID(ctx, req.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return nil, errors.New("apartment not found")
	}

	existing, err := s.residentRepo.GetByTelegramID(ctx, req.TelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to check telegram_id: %w", err)
	}

	chatID := req.TelegramID
	if req.ChatID != nil {
		chatID = *req.ChatID
	}

	if existing != nil {
		existing.ApartmentID = req.ApartmentID
		existing.ChatID = chatID
		existing.Name = req.Name
		existing.Phone = req.Phone
		if err := s.residentRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update resident: %w", err)
		}
		return existing, nil
	}

	resident := &domain.Resident{
		ApartmentID: req.ApartmentID,
		TelegramID:  req.TelegramID,
		ChatID:      chatID,
		Name:        req.Name,
		Phone:       req.Phone,
		Status:      "active",
	}

	if err := s.residentRepo.Create(ctx, resident); err != nil {
		return nil, fmt.Errorf("failed to create resident: %w", err)
	}

	return resident, nil
}

func (s *ResidentService) BulkCreateResidents(ctx context.Context, requests []domain.CreateResidentRequest) ([]*domain.Resident, []domain.BulkCreateError) {
	var residents []*domain.Resident
	var createErrors []domain.BulkCreateError

	for i, req := range requests {
		resident, err := s.CreateResident(ctx, req)
		if err != nil {
			createErrors = append(createErrors, domain.BulkCreateError{
				Row:   i + 1,
				Error: err.Error(),
			})
			continue
		}
		residents = append(residents, resident)
	}

	return residents, createErrors
}

func (s *ResidentService) ImportFromCSV(ctx context.Context, reader io.Reader, buildingID int64) (int, []error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return 0, []error{fmt.Errorf("failed to read CSV: %w", err)}
	}

	if len(records) < 2 {
		return 0, []error{errors.New("CSV must have header row and at least one data row")}
	}

	header := records[0]
	headerMap := make(map[string]int)
	for i, h := range header {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	requiredFields := []string{"apartment", "telegram_id"}
	for _, field := range requiredFields {
		if _, ok := headerMap[field]; !ok {
			return 0, []error{fmt.Errorf("missing required column: %s", field)}
		}
	}

	var requests []domain.CreateResidentRequest
	var parseErrors []error

	for i, record := range records[1:] {
		if len(record) < len(header) {
			parseErrors = append(parseErrors, fmt.Errorf("row %d: insufficient columns", i+2))
			continue
		}

		apartmentNumber := strings.TrimSpace(record[headerMap["apartment"]])
		if apartmentNumber == "" {
			parseErrors = append(parseErrors, fmt.Errorf("row %d: apartment is required", i+2))
			continue
		}

		apartments, err := s.apartmentRepo.GetByBuildingID(ctx, buildingID)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("row %d: failed to get apartments: %w", i+2, err))
			continue
		}

		var apartmentID int64
		found := false
		for _, apt := range apartments {
			if apt.Number == apartmentNumber {
				apartmentID = apt.ID
				found = true
				break
			}
		}

		if !found {
			parseErrors = append(parseErrors, fmt.Errorf("row %d: apartment %s not found", i+2, apartmentNumber))
			continue
		}

		telegramIDStr := strings.TrimSpace(record[headerMap["telegram_id"]])
		telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("row %d: invalid telegram_id: %s", i+2, telegramIDStr))
			continue
		}

		chatID := telegramID
		if chatIDIdx, ok := headerMap["chat_id"]; ok {
			chatIDStr := strings.TrimSpace(record[chatIDIdx])
			if chatIDStr != "" {
				if parsedChatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
					chatID = parsedChatID
				}
			}
		}

		req := domain.CreateResidentRequest{
			ApartmentID: apartmentID,
			TelegramID:  telegramID,
			ChatID:      &chatID,
		}

		if nameIdx, ok := headerMap["name"]; ok {
			name := strings.TrimSpace(record[nameIdx])
			if name != "" {
				req.Name = &name
			}
		}

		if phoneIdx, ok := headerMap["phone"]; ok {
			phone := strings.TrimSpace(record[phoneIdx])
			if phone != "" {
				req.Phone = &phone
			}
		}

		requests = append(requests, req)
	}

	if len(parseErrors) > 0 {
		return 0, parseErrors
	}

	residents, createErrors := s.BulkCreateResidents(ctx, requests)
	s.logger.Info("bulk import completed",
		zap.Int("total", len(requests)),
		zap.Int("success", len(residents)),
		zap.Int("errors", len(createErrors)),
	)

	var errorList []error
	for _, err := range createErrors {
		errorList = append(errorList, fmt.Errorf("row %d: %s", err.Row, err.Error))
	}

	return len(residents), errorList
}

func (s *ResidentService) ListResidents(ctx context.Context, filters domain.ResidentFilters) ([]*domain.Resident, error) {
	return s.residentRepo.List(ctx, filters)
}

func (s *ResidentService) DeleteResident(ctx context.Context, id int64) error {
	resident, err := s.residentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get resident: %w", err)
	}
	if resident == nil {
		return errors.New("resident not found")
	}

	if err := s.residentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete resident: %w", err)
	}

	return nil
}
