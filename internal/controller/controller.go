package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/AsmrS4/certificates-plugin-go/internal/dto"
	apperrors "github.com/AsmrS4/certificates-plugin-go/internal/errors"
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

	ctx.JSON(200, map[string]interface{}{
		"data":   results,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
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

	if strings.TrimSpace(payload.Reason) == "" {
		ctx.JSON(400, map[string]string{"error": "rejection reason is required"})
		return
	}

	if err := h.m.RejectOrder(id, payload.Reason); err != nil {
		h.handleError(ctx, err)
		return
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

	if err := h.m.PrepareOrder(id); err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(200, map[string]string{"status": "processing"})
}

func (h *HttpController) Upload(ctx *wasmplugin.EventContext) {
	idParam := ctx.HTTP.Query["id"]
	if idParam == "" {
		ctx.JSON(400, map[string]string{"error": "missing id parameter"})
		return
	}
	_, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid id format"})
		return
	}

	var payload struct {
		FileID   string `json:"file_id"`
		FileName string `json:"file_name"`
	}
	if err := json.Unmarshal([]byte(ctx.HTTP.Body), &payload); err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid JSON body"})
		return
	}

	if payload.FileID == "" || payload.FileName == "" {
		ctx.JSON(400, map[string]string{"error": "file_id and file_name are required"})
		return
	}

	// здесь должна быть логика сохранения файла через s.managementRepo.UploadCertificate
	// например:
	// _, err = h.m.UploadCertificate(id, payload.FileID, payload.FileName)
	if err != nil {
		h.handleError(ctx, err)
		return
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
