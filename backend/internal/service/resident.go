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
	"yardpass/internal/observability/logger"
	"yardpass/internal/observability/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type ResidentService struct {
	residentRepo   domain.ResidentRepository
	apartmentRepo  domain.ApartmentRepository
	fallbackLogger *zap.Logger
	opsTotal       *prometheus.CounterVec
	importTotal    *prometheus.CounterVec
}

func NewResidentService(residentRepo domain.ResidentRepository, apartmentRepo domain.ApartmentRepository, logger *zap.Logger, m *metrics.Metrics) *ResidentService {
	opsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_resident",
			Name:      "operations_total",
			Help:      "Total number of resident operations",
		},
		[]string{"operation", "result"},
	)

	importTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_resident",
			Name:      "import_rows_total",
			Help:      "Total number of rows processed during CSV import",
		},
		[]string{"result"},
	)

	m.GetRegistry().MustRegister(opsTotal, importTotal)

	for _, op := range []string{"create", "delete"} {
		for _, res := range []string{"success", "error"} {
			opsTotal.WithLabelValues(op, res)
		}
	}
	for _, res := range []string{"success", "error"} {
		importTotal.WithLabelValues(res)
	}

	return &ResidentService{
		residentRepo:   residentRepo,
		apartmentRepo:  apartmentRepo,
		fallbackLogger: logger,
		opsTotal:       opsTotal,
		importTotal:    importTotal,
	}
}

func (s *ResidentService) CreateResident(ctx context.Context, req domain.CreateResidentRequest) (*domain.Resident, error) {
	apartmentID, err := s.resolveApartmentID(ctx, req)
	if err != nil {
		return nil, err
	}

	apartment, err := s.apartmentRepo.GetByID(ctx, apartmentID)
	if err != nil {
		return nil, errors.New("Не удалось проверить квартиру. Попробуйте позже.")
	}
	if apartment == nil {
		return nil, errors.New("Квартира не найдена.")
	}

	existing, err := s.residentRepo.GetByTelegramID(ctx, req.TelegramID)
	if err != nil {
		return nil, errors.New("Не удалось проверить Telegram. Попробуйте позже.")
	}
	if existing != nil {
		s.opsTotal.WithLabelValues("create", "error").Inc()
		return nil, errors.New("Житель с таким Telegram ID уже существует.")
	}

	chatID := req.TelegramID
	if req.ChatID != nil {
		chatID = *req.ChatID
	}

	resident := &domain.Resident{
		ApartmentID: apartmentID,
		TelegramID:  req.TelegramID,
		ChatID:      chatID,
		Name:        req.Name,
		Phone:       req.Phone,
		Status:      "active",
	}

	if err := s.residentRepo.Create(ctx, resident); err != nil {
		s.opsTotal.WithLabelValues("create", "error").Inc()
		return nil, errors.New("Не удалось сохранить жителя. Попробуйте позже.")
	}

	s.opsTotal.WithLabelValues("create", "success").Inc()
	return resident, nil
}

func (s *ResidentService) BulkCreateResidents(ctx context.Context, requests []domain.CreateResidentRequest) ([]domain.Resident, []domain.BulkCreateError) {
	var residents []domain.Resident
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
		residents = append(residents, *resident)
	}

	return residents, createErrors
}

func (s *ResidentService) ImportFromCSV(ctx context.Context, reader io.Reader, buildingID int64) (int, []error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		s.importTotal.WithLabelValues("error").Inc()
		return 0, []error{errors.New("Не удалось прочитать CSV. Проверьте кодировку и разделитель полей.")}
	}

	if len(records) < 2 {
		s.importTotal.WithLabelValues("error").Inc()
		return 0, []error{errors.New("В файле должны быть строка заголовков и хотя бы одна строка с данными.")}
	}

	header := records[0]
	headerMap := make(map[string]int)
	for i, h := range header {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	requiredFields := []string{"apartment", "telegram_id"}
	for _, field := range requiredFields {
		if _, ok := headerMap[field]; !ok {
			s.importTotal.WithLabelValues("error").Inc()
			return 0, []error{fmt.Errorf("В заголовке CSV отсутствует обязательная колонка: %s", field)}
		}
	}

	var requests []domain.CreateResidentRequest
	var parseErrors []error

	for i, record := range records[1:] {
		if len(record) < len(header) {
			parseErrors = append(parseErrors, fmt.Errorf("Строка %d: недостаточно столбцов.", i+2))
			continue
		}

		apartmentNumber := strings.TrimSpace(record[headerMap["apartment"]])
		if apartmentNumber == "" {
			parseErrors = append(parseErrors, fmt.Errorf("Строка %d: укажите номер квартиры (apartment).", i+2))
			continue
		}

		apartments, err := s.apartmentRepo.GetByBuildingID(ctx, buildingID)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("Строка %d: не удалось загрузить квартиры здания.", i+2))
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
			parseErrors = append(parseErrors, fmt.Errorf("Строка %d: квартира «%s» не найдена в этом здании.", i+2, apartmentNumber))
			continue
		}

		telegramIDStr := strings.TrimSpace(record[headerMap["telegram_id"]])
		telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("Строка %d: некорректный telegram_id «%s» (нужно целое число).", i+2, telegramIDStr))
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
			ApartmentID: &apartmentID,
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
		s.importTotal.WithLabelValues("error").Add(float64(len(parseErrors)))
		return 0, parseErrors
	}

	residents, createErrors := s.BulkCreateResidents(ctx, requests)

	s.importTotal.WithLabelValues("success").Add(float64(len(residents)))
	s.importTotal.WithLabelValues("error").Add(float64(len(createErrors)))

	lgr := logger.FromContext(ctx)
	if lgr == nil {
		lgr = s.fallbackLogger
	}
	lgr.Info("Bulk import completed",
		zap.Int("total", len(requests)),
		zap.Int("success", len(residents)),
		zap.Int("errors", len(createErrors)),
	)

	var errorList []error
	for _, err := range createErrors {
		errorList = append(errorList, fmt.Errorf("Строка %d: %s", err.Row, err.Error))
	}

	return len(residents), errorList
}

func (s *ResidentService) resolveApartmentID(ctx context.Context, req domain.CreateResidentRequest) (int64, error) {
	if req.ApartmentID != nil {
		return *req.ApartmentID, nil
	}
	if req.BuildingID == nil {
		return 0, errors.New("Укажите building_id, если не передан apartment_id.")
	}
	if req.ApartmentNumber == nil || strings.TrimSpace(*req.ApartmentNumber) == "" {
		return 0, errors.New("Укажите apartment_number, если не передан apartment_id.")
	}

	apartments, err := s.apartmentRepo.GetByBuildingID(ctx, *req.BuildingID)
	if err != nil {
		return 0, errors.New("Не удалось загрузить квартиры здания. Попробуйте позже.")
	}
	targetNumber := strings.TrimSpace(*req.ApartmentNumber)
	for _, apt := range apartments {
		if strings.EqualFold(strings.TrimSpace(apt.Number), targetNumber) {
			return apt.ID, nil
		}
	}

	return 0, fmt.Errorf("Квартира «%s» не найдена в указанном здании.", targetNumber)
}

func (s *ResidentService) ListResidents(ctx context.Context, filters domain.ResidentFilters) ([]domain.Resident, error) {
	return s.residentRepo.List(ctx, filters)
}

func (s *ResidentService) DeleteResident(ctx context.Context, id int64) error {
	resident, err := s.residentRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("Не удалось найти жителя. Попробуйте позже.")
	}
	if resident == nil {
		return errors.New("Житель не найден.")
	}

	if err := s.residentRepo.Delete(ctx, id); err != nil {
		s.opsTotal.WithLabelValues("delete", "error").Inc()
		return errors.New("Не удалось удалить жителя. Попробуйте позже.")
	}

	s.opsTotal.WithLabelValues("delete", "success").Inc()
	return nil
}
