package handler

import (
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

type CertificateHandler struct {
	service *service.MessengerService
	cat     *wasmplugin.Catalog
}

func NewCertificateHandler(service *service.MessengerService, cat *wasmplugin.Catalog) *CertificateHandler {
	return &CertificateHandler{service: service, cat: cat}
}

func (h *CertificateHandler) CreateOrder(ctx *wasmplugin.EventContext) error {
	tr := h.cat.Tr(ctx.Locale())

	certType := ctx.Param("type")
	obtainMethod := ctx.Param("obtain_method")
	studentID := ctx.Messenger.UserID

	formData := make(map[string]interface{})
	var comment string
	ctx.Log(fmt.Sprintf("DEBUG: all params: %+v", ctx.Messenger.Params))
	for key, value := range ctx.Messenger.Params {
		if key == "type" || key == "obtain_method" {
			continue
		}
		if key == "comment" {
			comment = value
			continue
		}
		formData[key] = value
	}

	var files []*dto.File
	if ctx.HasFiles() {
		for _, f := range ctx.Files() {
			data, err := ctx.FileReadAll(f.ID)
			if err != nil {
				ctx.LogError(fmt.Sprintf("failed to read file %s: %v", f.Name, err))
				continue
			}
			stored, err := ctx.FileStore(f.Name, f.MIMEType, f.FileType, data)
			if err != nil {
				ctx.LogError(fmt.Sprintf("failed to store file %s: %v", f.Name, err))
				continue
			}
			var file = &dto.File{
				ID:       stored.ID,
				Name:     stored.Name,
				MIMEType: stored.MIMEType,
				FileType: stored.FileType,
			}
			files = append(files, file)
		}
	} else {
		ctx.LogError(fmt.Sprintf("No file attachment"))
	}
	ctx.Log(fmt.Sprintf("DEBUG: c.FormData in Handler call CreateOrder: %+v", formData))
	req := service.CreateOrderRequest{
		StudentID:       studentID,
		CertificateType: certType,
		ObtainMethod:    obtainMethod,
		Comment:         comment,
		FormData:        formData,
		Files:           files,
	}

	orderID, err := h.service.SaveOrder(ctx, req)
	if err != nil {
		handleMessengerError(ctx, tr, err)
		return nil
	}

	ctx.Reply(wasmplugin.NewMessage(fmt.Sprintf(tr("order_created"), orderID)))
	ctx.Log(fmt.Sprintf("order %d created by user %d", orderID, studentID))
	return nil
}

func (h *CertificateHandler) CancelOrder(ctx *wasmplugin.EventContext) error {
	tr := h.cat.Tr(ctx.Locale())

	confirm := ctx.Param("confirm_cancellation")
	if confirm != "yes" {
		ctx.Reply(wasmplugin.NewMessage(tr("cancel_cancelled")))
		return nil
	}

	idParam := ctx.Param("enter_id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.Reply(wasmplugin.NewMessage(tr("incorrect_id")))
		return nil
	}

	req := service.RejectOrderRequest{OrderID: id}
	if err := h.service.RejectOrder(req); err != nil {
		handleMessengerError(ctx, tr, err)
		return nil
	}

	ctx.Reply(wasmplugin.NewMessage(tr("order_cancelled_successfully", id)))
	ctx.Log(fmt.Sprintf("order %d cancelled by user %d", id, ctx.Messenger.UserID))
	return nil
}

func (h *CertificateHandler) FindOrderByID(ctx *wasmplugin.EventContext) error {
	tr := h.cat.Tr(ctx.Locale())

	idParam := ctx.Param("enter_id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.Reply(wasmplugin.NewMessage(tr("incorrect_id")))
		return nil
	}

	req := service.FindOrderRequest{OrderID: id}
	details, err := h.service.GetOrderDetails(ctx, req)
	if err != nil {
		handleMessengerError(ctx, tr, err)
		return nil
	}
	if details.Application == nil {
		ctx.Reply(wasmplugin.NewMessage(fmt.Sprintf(tr("order_not_found"), id)))
		return nil
	}

	msgText := h.formatOrderMessage(details.Application, tr)
	msg := wasmplugin.NewMessage(msgText)

	if len(details.FileIDs) == 0 {
		ctx.Log(fmt.Sprintf("user %d viewed order %d (no files)", ctx.Messenger.UserID, id))
	}

	ctx.Reply(msg)
	ctx.Log(fmt.Sprintf("user %d viewed order %d with %d file(s)", ctx.Messenger.UserID, id))
	return nil
}

func (h *CertificateHandler) FindAllOrders(ctx *wasmplugin.EventContext) error {
	tr := h.cat.Tr(ctx.Locale())
	status := ctx.Param("status")
	certType := ctx.Param("type")
	userID := ctx.Messenger.UserID

	req := service.FindAllRequest{
		StudentID: userID,
		Status:    status,
		Type:      certType,
	}
	orders, err := h.service.FindAll(req)
	if err != nil {
		handleMessengerError(ctx, tr, err)
		return nil
	}

	if len(orders) == 0 {
		ctx.Reply(wasmplugin.NewMessage(tr("no_orders")))
		return nil
	}

	var msg string
	msg = fmt.Sprintf(tr("orders_header"), len(orders)) + "\n\n"
	for _, o := range orders {
		msg += h.formatOrderMessage(o, tr) + "\n_______________\n"
	}
	ctx.Reply(wasmplugin.NewMessage(msg))
	return nil
}

func (h *CertificateHandler) formatOrderMessage(order *entity.CertificateApplication, tr func(key string, args ...any) string) string {
	var parts []string

	typeName := tr("certificate_type_" + string(order.Type))
	methodName := tr("obtain_method_" + string(order.ObtainMethod))
	statusName := tr("status_" + string(order.Status))

	parts = append(parts, fmt.Sprintf("%s: %d", tr("order_info_id"), order.ID))
	parts = append(parts, fmt.Sprintf("%s: %s", tr("order_info_type"), typeName))
	parts = append(parts, fmt.Sprintf("%s: %s", tr("order_info_obtain_method"), methodName))
	parts = append(parts, fmt.Sprintf("%s: %s", tr("order_info_status"), statusName))
	if order.Status == "Rejected" && order.RejectionReason != "" {
		parts = append(parts, fmt.Sprintf("\n%s: %s", tr("order_info_rejection_reason"), order.RejectionReason))
	}
	if order.Status == "Done" && order.ObtainMethod == "Paper" {
		parts = append(parts, fmt.Sprintf("\n%s", tr("order_paper_done_comment")))
	}
	return strings.Join(parts, "\n")
}

func handleMessengerError(ctx *wasmplugin.EventContext, tr func(key string, args ...any) string, err error) {
	var appErr *apperrors.AppError

	if errors.As(err, &appErr) {
		msg := tr(appErr.Key, appErr.Args...)
		ctx.Reply(wasmplugin.NewMessage(msg))
		return
	}

	ctx.LogError(err.Error())
	ctx.Reply(wasmplugin.NewMessage(tr("error_default")))
}
