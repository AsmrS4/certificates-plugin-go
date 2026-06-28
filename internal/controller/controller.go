package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/AsmrS4/certificates-plugin-go/internal/dto"
	apperrors "github.com/AsmrS4/certificates-plugin-go/internal/errors"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"
	"github.com/AsmrS4/certificates-plugin-go/internal/service"
	wasmplugin "github.com/SuperBotForge/sdk/go-sdk"
)

type HttpController struct {
	m   *service.ManagementService
	cat *wasmplugin.Catalog
}

func NewHttpController(m *service.ManagementService, cat *wasmplugin.Catalog) *HttpController {
	return &HttpController{m: m, cat: cat}
}

func (h *HttpController) GetAll(ctx *wasmplugin.EventContext) {
	filter := dto.FindRequestsFilter{
		FullName:        ctx.HTTP.Query["full_name"],
		NationalityType: ctx.HTTP.Query["nationality_type"],
		FacultyName:     ctx.HTTP.Query["faculty_name"],
		GroupCode:       ctx.HTTP.Query["group_code"],
		Type:            ctx.HTTP.Query["type"],
		Status:          ctx.HTTP.Query["status"],
	}

	limit, err := strconv.Atoi(ctx.HTTP.Query["limit"])
	if err != nil || limit <= 0 {
		filter.Limit = 10
	} else {
		filter.Limit = limit
	}

	if strings.TrimSpace(filter.Status) != "" && filter.Status != "Pending" && filter.Status != "Prepare" {
		ctx.JSON(400, map[string]string{"error": "Only Pending and Prepare statuses are allowed"})
		return
	}

	offset, err := strconv.Atoi(ctx.HTTP.Query["offset"])
	if err != nil || offset < 0 {
		filter.Offset = 0
	} else {
		filter.Offset = offset
	}

	results, total, err := h.m.FindRequests(filter)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	pagination := map[string]interface{}{
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	ctx.JSON(200, map[string]interface{}{
		"data":       results,
		"pagination": pagination,
	})
}

func (h *HttpController) GetAllForeign(ctx *wasmplugin.EventContext) {
	filter := dto.FindRequestsFilter{
		FullName:        ctx.HTTP.Query["full_name"],
		NationalityType: "foreign",
		FacultyName:     ctx.HTTP.Query["faculty_name"],
		GroupCode:       ctx.HTTP.Query["group_code"],
		Type:            ctx.HTTP.Query["type"],
		Status:          ctx.HTTP.Query["status"],
	}

	limit, err := strconv.Atoi(ctx.HTTP.Query["limit"])
	if err != nil || limit <= 0 {
		filter.Limit = 10
	} else {
		filter.Limit = limit
	}

	if strings.TrimSpace(filter.Status) != "" && filter.Status != "Pending" && filter.Status != "Prepare" {
		ctx.JSON(400, map[string]string{"error": "Only Pending and Prepare statuses are allowed"})
		return
	}

	offset, err := strconv.Atoi(ctx.HTTP.Query["offset"])
	if err != nil || offset < 0 {
		filter.Offset = 0
	} else {
		filter.Offset = offset
	}

	results, total, err := h.m.FindRequests(filter)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	pagination := map[string]interface{}{
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	ctx.JSON(200, map[string]interface{}{
		"data":       results,
		"pagination": pagination,
	})
}

func (h *HttpController) GetDetails(ctx *wasmplugin.EventContext) {
	idParam := ctx.HTTP.Query["id"]
	if idParam == "" {
		ctx.JSON(400, map[string]string{"error": "missing id parameter"})
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid id format"})
		return
	}

	details, err := h.m.FindWithUserDetails(ctx, id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	for idx, file := range details.Attachments {
		url, err := ctx.FileURL(file.FileID)
		if err != nil {
			continue
		}
		details.Attachments[idx].File_URL = url
	}
	ctx.JSON(200, details)
}

func (h *HttpController) GetHistory(ctx *wasmplugin.EventContext) {
	filter := dto.FindRequestsFilter{
		FullName:        ctx.HTTP.Query["full_name"],
		NationalityType: ctx.HTTP.Query["nationality_type"],
		FacultyName:     ctx.HTTP.Query["faculty_name"],
		GroupCode:       ctx.HTTP.Query["group_code"],
		Type:            ctx.HTTP.Query["type"],
		Status:          ctx.HTTP.Query["status"],
	}

	limit, err := strconv.Atoi(ctx.HTTP.Query["limit"])
	if err != nil || limit <= 0 {
		filter.Limit = 10
	} else {
		filter.Limit = limit
	}

	if strings.TrimSpace(filter.Status) != "" && filter.Status != "Rejected" && filter.Status != "Done" {
		ctx.JSON(400, map[string]string{"error": "Only Rejected and Done statuses are allowed"})
		return
	}

	offset, err := strconv.Atoi(ctx.HTTP.Query["offset"])
	if err != nil || offset < 0 {
		filter.Offset = 0
	} else {
		filter.Offset = offset
	}

	results, total, err := h.m.GetHistory(filter)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	pagination := map[string]interface{}{
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	ctx.JSON(200, map[string]interface{}{
		"data":       results,
		"pagination": pagination,
	})
}

func (h *HttpController) Reject(ctx *wasmplugin.EventContext) {
	idParam := ctx.HTTP.Query["id"]
	if idParam == "" {
		ctx.JSON(400, map[string]string{"error": "missing id parameter"})
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid id format"})
		return
	}

	var payload struct {
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(ctx.HTTP.Body), &payload); err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid JSON body"})
		return
	}

	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		ctx.JSON(400, map[string]string{"error": "rejection reason is required"})
		return
	}

	studentID, err := h.m.RejectOrder(id, reason)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	var event = &dto.OrderEvent{
		UserID:      studentID,
		OrderID:     id,
		OrderStatus: string(entity.Rejected),
		MessageKey:  dto.KeyRejected,
		File:        nil,
		Locale:      "ru",
	}

	err = wasmplugin.PublishEvent("certificates.change", event)
	if err != nil {
		ctx.LogError(fmt.Sprintf("failed send notification after rejection: %s", err.Error()))
	}

	ctx.JSON(200, map[string]string{"status": "rejected"})
}

func (h *HttpController) Process(ctx *wasmplugin.EventContext) {
	idParam := ctx.HTTP.Query["id"]
	if idParam == "" {
		ctx.JSON(400, map[string]string{"error": "missing id parameter"})
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid id format"})
		return
	}

	studentID, err := h.m.PrepareOrder(id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	var event = &dto.OrderEvent{
		UserID:      studentID,
		OrderID:     id,
		OrderStatus: string(entity.Prepare),
		MessageKey:  dto.KeyPrepare,
		File:        nil,
		Locale:      "ru",
	}

	err = wasmplugin.PublishEvent("certificates.change", event)
	if err != nil {
		ctx.LogError(fmt.Sprintf("failed send notification after prepare: %s", err.Error()))
	}

	ctx.JSON(200, map[string]string{"status": "processing"})
}

func (h *HttpController) Upload(ctx *wasmplugin.EventContext) {
	idParam := ctx.HTTP.Query["id"]
	if idParam == "" {
		ctx.JSON(400, map[string]string{"error": "missing id parameter"})
		return
	}
	orderID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid id format"})
		return
	}

	var payload struct {
		ID       string
		Name     string
		MIMEType string
		FileType string
	}

	if err := json.Unmarshal([]byte(ctx.HTTP.Body), &payload); err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid JSON body"})
		return
	}

	if payload.ID == "" || payload.Name == "" {
		ctx.JSON(400, map[string]string{"error": "file_id and file_name are required"})
		return
	}

	// здесь должна быть логика сохранения файла через s.managementRepo.UploadCertificate
	// например:
	// _, err = h.m.UploadCertificate(id, payload.FileID, payload.FileName)
	studentID, err := h.m.UploadCertificateFile(orderID, payload)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	var event = &dto.OrderEvent{
		UserID:      studentID,
		OrderID:     orderID,
		OrderStatus: string(entity.Done),
		MessageKey:  dto.KeyDone,
		Locale:      "ru",
	}

	err = wasmplugin.PublishEvent("certificates.change", event)
	if err != nil {
		ctx.LogError(fmt.Sprintf("failed send notification after prepare: %s", err.Error()))
	}

	ctx.JSON(200, map[string]string{"status": "uploaded"})
}

func (h *HttpController) handleError(ctx *wasmplugin.EventContext, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		status := appErr.Status
		if status == 0 {
			status = 500
		}
		ctx.JSON(status, map[string]string{"error": appErr.Key})
		return
	}
	ctx.LogError(fmt.Sprintf("unexpected error: %v", err))
	ctx.JSON(500, map[string]string{"error": "internal server error"})
}
