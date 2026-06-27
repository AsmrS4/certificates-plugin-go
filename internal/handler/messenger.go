package handler

import (
	"errors"
	"fmt"
	"strconv"

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

	var files []service.FileInfo
	if ctx.HasFiles() {
		for _, f := range ctx.Files() {
			files = append(files, service.FileInfo{
				ID:       f.ID,
				Name:     f.Name,
				FileType: f.FileType,
				MimeType: f.MIMEType,
				Size:     f.Size,
			})
		}
	}

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
	order, err := h.service.FindOrder(req)
	if err != nil {
		handleMessengerError(ctx, tr, err)
		return nil
	}
	if order == nil {
		ctx.Reply(wasmplugin.NewMessage(fmt.Sprintf(tr("order_not_found"), id)))
		return nil
	}

	msg := h.formatOrderMessage(order, tr)
	ctx.Reply(wasmplugin.NewMessage(msg))
	ctx.Log(fmt.Sprintf("user %d viewed order %d", ctx.Messenger.UserID, id))
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
	typeName := tr("certificate_type_" + string(order.Type))
	methodName := tr("obtain_method_" + string(order.ObtainMethod))
	statusName := tr("status_" + string(order.Status))

	return fmt.Sprintf(
		"%s: %d\n%s: %s\n%s: %s\n%s: %s",
		tr("order_info_id"), order.ID,
		tr("order_info_type"), typeName,
		tr("order_info_obtain_method"), methodName,
		tr("order_info_status"), statusName,
	)
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
