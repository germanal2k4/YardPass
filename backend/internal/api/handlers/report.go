package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ReportHandler struct {
	scanEventRepo domain.ScanEventRepository
	passRepo      domain.PassRepository
}

func NewReportHandler(scanEventRepo domain.ScanEventRepository, passRepo domain.PassRepository) *ReportHandler {
	return &ReportHandler{
		scanEventRepo: scanEventRepo,
		passRepo:      passRepo,
	}
}

func (h *ReportHandler) GetStatistics(c *gin.Context) {
	var from, to *time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &parsed
		}
	}

	role, exists := c.Get("role")
	if !exists {
		errors.Unauthorized(c, "MISSING_ROLE", "User role not found")
		return
	}

	roleStr, ok := role.(string)
	if !ok {
		errors.InternalServerError(c, "INVALID_ROLE", "Invalid role type")
		return
	}

	buildingID, _ := c.Get("building_id")

	var bID *int64

	if roleStr == "superuser" {
		if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
			if id, err := strconv.ParseInt(buildingIDStr, 10, 64); err == nil {
				bID = &id
			}
		}
	} else if roleStr == "admin" || roleStr == "guard" {
		if buildingID != nil {
			if id, ok := buildingID.(int64); ok {
				bID = &id
			}
		}
	}

	stats, err := h.scanEventRepo.GetStatistics(c.Request.Context(), from, to, bID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_scans":   stats.TotalScans,
		"valid_scans":   stats.ValidScans,
		"invalid_scans": stats.InvalidScans,
		"unique_passes": stats.UniquePasses,
		"unique_guards": stats.UniqueGuards,
		"valid_percent": calculatePercent(stats.ValidScans, stats.TotalScans),
		"period_from":   from,
		"period_to":     to,
	})
}

func (h *ReportHandler) ExportToExcel(c *gin.Context) {
	format := c.Query("format")
	if format != "xlsx" {
		errors.BadRequest(c, "INVALID_FORMAT", "format must be xlsx")
		return
	}

	var from, to *time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &parsed
		}
	}

	role, exists := c.Get("role")
	if !exists {
		errors.Unauthorized(c, "MISSING_ROLE", "User role not found")
		return
	}

	roleStr, ok := role.(string)
	if !ok {
		errors.InternalServerError(c, "INVALID_ROLE", "Invalid role type")
		return
	}

	buildingID, _ := c.Get("building_id")

	var bID *int64

	if roleStr == "superuser" {
		if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
			if id, err := strconv.ParseInt(buildingIDStr, 10, 64); err == nil {
				bID = &id
			}
		}
	} else if roleStr == "admin" || roleStr == "guard" {
		if buildingID != nil {
			if id, ok := buildingID.(int64); ok {
				bID = &id
			}
		}
	}

	var filters domain.ScanEventFilters
	filters.From = from
	filters.To = to
	filters.Limit = 10000

	events, err := h.scanEventRepo.GetEventsWithDetails(c.Request.Context(), filters, bID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	file := excelize.NewFile()
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "События"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		errors.InternalServerError(c, "EXCEL_ERROR", err.Error())
		return
	}

	file.SetActiveSheet(index)

	headers := []string{"ID", "Дата/Время", "Результат", "Номер авто", "Квартира", "Охранник", "Причина"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		file.SetCellValue(sheetName, cell, header)
	}

	for i, event := range events {
		row := i + 2
		file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), event.ID)
		file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), event.ScannedAt.Format("2006-01-02 15:04:05"))
		file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), event.Result)
		file.SetCellValue(sheetName, fmt.Sprintf("D%d", row), event.CarPlate)
		file.SetCellValue(sheetName, fmt.Sprintf("E%d", row), event.ApartmentNumber)
		file.SetCellValue(sheetName, fmt.Sprintf("F%d", row), event.GuardUsername)
		if event.Reason != nil {
			file.SetCellValue(sheetName, fmt.Sprintf("G%d", row), *event.Reason)
		}
	}

	file.DeleteSheet("Sheet1")

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s.xlsx", time.Now().Format("20060102_150405")))

	if err := file.Write(c.Writer); err != nil {
		errors.InternalServerError(c, "EXCEL_ERROR", err.Error())
		return
	}
}

func calculatePercent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
