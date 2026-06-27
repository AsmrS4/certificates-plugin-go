package controller

import (
	"github.com/AsmrS4/certificates-plugin-go/internal/service"
	wasmplugin "github.com/SuperBotForge/sdk/go-sdk"
)

type HttpController struct {
	m   *service.ManagementService
	cat *wasmplugin.Catalog
}

func (h *HttpController) NewController(m *service.ManagementService, cat *wasmplugin.Catalog) *HttpController {
	return &HttpController{m: m, cat: cat}
}

func (h *HttpController) GetDetails(ctx *wasmplugin.EventContext) {}

func (h *HttpController) GetAll(ctx *wasmplugin.EventContext) {}

func (h *HttpController) Reject(ctx *wasmplugin.EventContext) {}

func (h *HttpController) Process(ctx *wasmplugin.EventContext) {}

func (h *HttpController) Upload(ctx *wasmplugin.EventContext) {}
