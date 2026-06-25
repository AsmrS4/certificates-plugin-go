package main

import (
	"database/sql"
	"embed"
	"log"

	"github.com/AsmrS4/certificates-plugin-go/internal/handler"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/repo"
	"github.com/AsmrS4/certificates-plugin-go/internal/service"
	"github.com/AsmrS4/certificates-plugin-go/internal/utils"
	wasmplugin "github.com/SuperBotForge/sdk/go-sdk"
)

//go:embed i18n/*.toml
var i18nFS embed.FS

//go:embed config/certificate_types.yaml
var configFS embed.FS

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	cat = wasmplugin.NewCatalog("en").
		LoadFS(i18nFS, "i18n")
	registry map[string]utils.CertificateTypeConfig
)

var (
	db           *sql.DB
	certRepo     persistence.CertificateRepo
	messengerSvc *service.MessengerService
	certHandler  *handler.CertificateHandler
)

func initDependencies() error {
	var err error
	db, err = persistence.OpenDBConnection()
	if err != nil {
		return err
	}
	certRepo = repo.NewRepo(db)
	messengerSvc = service.NewMessengerService(certRepo, db)
	certHandler = handler.NewCertificateHandler(messengerSvc, cat)
	return nil
}

func main() {
	if err := initDependencies(); err != nil {
		log.Fatalf("init failed: %v", err)
	}
	data, err := configFS.ReadFile("config/certificate_types.yaml")
	if err != nil {
		panic(err)
	}

	registry, err = utils.ParseCertificateTypes(data)
	if err != nil {
		panic(err)
	}
	_ = registry

	wasmplugin.Run(wasmplugin.Plugin{
		ID:      "certificates_plugin",
		Name:    "Заказ справки из деканата",
		Version: "1.0.9",
		Requirements: []wasmplugin.Requirement{
			wasmplugin.Database("Store applications for a certificate plugin").Build(),
			wasmplugin.File("Store and serve uploaded documents appendix to the certificate plugin").Build(),
			wasmplugin.NotifyReq("Send notifications to users").Build(),
			wasmplugin.EventsReq("Public events to users").Build(),
			wasmplugin.UserInfoReq("Student info").Build(),
		},
		Migrations: wasmplugin.MigrationsFromFS(migrationsFS, "migrations"),
		Triggers: []wasmplugin.Trigger{
			order_command(),
			cancel_command(),
			find_command(),
			find_all_command(),
		},
	})
}

func order_command() wasmplugin.Trigger {
	var nodes []wasmplugin.Node
	nodes = append(nodes, wasmplugin.NewStep("type").
		LocalizedText(cat.L("select_certificate_type"), wasmplugin.StylePlain).
		DynamicOptions("",
			func(cbCtx *wasmplugin.CallbackContext) []wasmplugin.Option {
				var opts []wasmplugin.Option
				for id, cfg := range registry {
					label := cfg.DisplayName[cbCtx.Locale]
					if label == "" {
						label = cfg.DisplayName["en"]
					}
					opts = append(opts, wasmplugin.Opt(label, id))
				}
				return opts
			},
		),
	)

	for id, cfg := range registry {
		for _, field := range cfg.Fields {
			inSteps := false
			for _, step := range cfg.Steps {
				if step.Field == field.Name {
					inSteps = true
					break
				}
			}
			if !inSteps {
				continue
			}

			node := wasmplugin.NewStep(field.Name).
				LocalizedText(field.Label, wasmplugin.StylePlain).
				VisibleWhenFunc(func(ctx *wasmplugin.CallbackContext) bool {
					return ctx.Params["type"] == id
				})

			if field.Required {
				node = node.Validate(`.+`)
			}

			nodes = append(nodes, node)
		}
	}

	nodes = append(nodes, wasmplugin.NewStep("attachments").
		LocalizedText(cat.L("upload_attachments"), wasmplugin.StylePlain).
		VisibleWhenFunc(func(ctx *wasmplugin.CallbackContext) bool {
			typ := ctx.Params["type"]
			cfg, ok := registry[typ]
			return ok && cfg.RequiresAttachments
		}),
	)

	nodes = append(nodes, wasmplugin.NewStep("obtain_method").
		LocalizedText(cat.L("select_certificate_obtain"), wasmplugin.StylePlain).
		DynamicOptions("",
			func(cbCtx *wasmplugin.CallbackContext) []wasmplugin.Option {
				return []wasmplugin.Option{
					wasmplugin.Opt(cat.L("paper")[cbCtx.Locale], "Paper"),
					wasmplugin.Opt(cat.L("electronic")[cbCtx.Locale], "Electronic"),
				}
			},
		),
	)

	return wasmplugin.Trigger{
		Name: "order_cert",
		Type: wasmplugin.TriggerMessenger,
		Descriptions: map[string]string{
			"ru": "Заказать справку",
			"en": "Order certificate",
		},
		Nodes: nodes,
		Handler: func(ctx *wasmplugin.EventContext) error {
			certType := ctx.Param("type")
			formData := make(map[string]string)
			cfg, ok := registry[certType]
			if ok {
				for _, field := range cfg.Fields {
					if val := ctx.Param(field.Name); val != "" {
						formData[field.Name] = val
					}
				}
			}
			certHandler.CreateOrder(ctx)
			return nil
		},
	}
}

func cancel_command() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name: "cancel",
		Type: wasmplugin.TriggerMessenger,
		Descriptions: map[string]string{
			"ru": "Отменить заказ",
			"en": "Cancel order",
		},
		Nodes: []wasmplugin.Node{
			wasmplugin.NewStep("enter_id").
				LocalizedText(cat.L("enter_order_id"), wasmplugin.StyleHeader).
				Validate(`^\d+$`),
			wasmplugin.NewStep("confirm_cancellation").
				LocalizedText(cat.L("confirm_cancel_order"), wasmplugin.StyleHeader).
				DynamicOptions("",
					func(cbCtx *wasmplugin.CallbackContext) []wasmplugin.Option {
						return []wasmplugin.Option{
							wasmplugin.Opt(cat.L("yes")[cbCtx.Locale], "yes"),
							wasmplugin.Opt(cat.L("no")[cbCtx.Locale], "no"),
						}
					},
				),
		},
		Handler: func(ctx *wasmplugin.EventContext) error {
			certHandler.CancelOrder(ctx)
			return nil
		},
	}
}

func find_command() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name: "find",
		Type: wasmplugin.TriggerMessenger,
		Descriptions: map[string]string{
			"ru": "Найти справку",
			"en": "Find certificate",
		},
		Nodes: []wasmplugin.Node{
			wasmplugin.NewStep("enter_id").
				LocalizedText(cat.L("enter_order_id"), wasmplugin.StyleHeader).
				Validate(`^\d+$`),
		},
		Handler: func(ctx *wasmplugin.EventContext) error {
			certHandler.FindOrderByID(ctx)
			return nil
		},
	}
}

func find_all_command() wasmplugin.Trigger {
	var nodes []wasmplugin.Node
	nodes = append(nodes, wasmplugin.NewStep("status").
		LocalizedText(cat.L("filter_by"), wasmplugin.StyleHeader).
		DynamicOptions("",
			func(cbCtx *wasmplugin.CallbackContext) []wasmplugin.Option {
				return []wasmplugin.Option{
					wasmplugin.Opt(cat.L("pending")[cbCtx.Locale], "Pending"),
					wasmplugin.Opt(cat.L("prepare")[cbCtx.Locale], "Prepare"),
					wasmplugin.Opt(cat.L("done")[cbCtx.Locale], "Done"),
					wasmplugin.Opt(cat.L("skip")[cbCtx.Locale], "Skip"),
				}
			},
		))
	nodes = append(nodes,
		wasmplugin.ConditionalBranch(
			wasmplugin.When(
				wasmplugin.ParamNeq("status", "Skip"),
				wasmplugin.NewStep("type").
					LocalizedText(cat.L("select_certificate_type"), wasmplugin.StylePlain).
					DynamicOptions("",
						func(cbCtx *wasmplugin.CallbackContext) []wasmplugin.Option {
							var opts []wasmplugin.Option
							for id, cfg := range registry {
								label := cfg.DisplayName[cbCtx.Locale]
								if label == "" {
									label = cfg.DisplayName["en"]
								}
								opts = append(opts, wasmplugin.Opt(label, id))
							}
							return opts
						},
					),
			),
		),
	)

	return wasmplugin.Trigger{
		Name: "all",
		Type: wasmplugin.TriggerMessenger,
		Descriptions: map[string]string{
			"ru": "Последние заказы",
			"en": "Show recently orders",
		},
		Nodes: nodes,
		Handler: func(ctx *wasmplugin.EventContext) error {
			certHandler.FindAllOrders(ctx)
			return nil
		},
	}
}

func get_all_http() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name:        "Find all certificate orders",
		Type:        wasmplugin.TriggerHTTP,
		Description: "Find all ordered certificate requests from users.",
		Path:        "/api/certificates/all",
		Methods:     []string{"GET"},
		Handler: func(ctx *wasmplugin.EventContext) error {
			return nil
		},
	}
}

func get_details_http() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name:        "Certificate order details",
		Type:        wasmplugin.TriggerHTTP,
		Description: "Find concrete certificate order details",
		Path:        "/api/certificates",
		Methods:     []string{"GET"},
		Handler: func(ctx *wasmplugin.EventContext) error {
			return nil
		},
	}
}

func process_http() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name:        "Start process certificate order",
		Type:        wasmplugin.TriggerHTTP,
		Description: "Start process certificate order",
		Path:        "/api/certificates/process",
		Methods:     []string{"POST"},
		Handler: func(ctx *wasmplugin.EventContext) error {
			return nil
		},
	}
}

func reject_http() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name:        "Reject certificate order",
		Type:        wasmplugin.TriggerHTTP,
		Description: "Rejection process certificate order",
		Path:        "/api/certificates/reject",
		Methods:     []string{"DELETE"},
		Handler: func(ctx *wasmplugin.EventContext) error {
			return nil
		},
	}
}

func upload_http() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name:        "Upload certificate document",
		Description: "Upload created document to certificate order",
		Type:        wasmplugin.TriggerHTTP,
		Path:        "/api/certificates/upload",
		Methods:     []string{"POST"},
		Handler: func(ctx *wasmplugin.EventContext) error {
			return nil
		},
	}
}

func notify() wasmplugin.Trigger {
	return wasmplugin.Trigger{
		Name:  "on_change_order_status",
		Type:  wasmplugin.TriggerEvent,
		Topic: "certificate_order.updated",
		Handler: func(ctx *wasmplugin.EventContext) error {
			// tr := cat.Tr(ctx.Locale())
			// var payload models.OrderEvent
			// raw := ctx.Event.Payload
			// if err := json.Unmarshal(raw, &payload); err != nil {
			// 	ctx.LogError(fmt.Sprintf("notification listener error: %s", err.Error()))
			// 	return nil
			// }
			// message := fmt.Sprintf(tr("change_status_event"), payload.OrderID, payload.OrderStatus)
			// return ctx.NotifyUser(payload.UserID, message, wasmplugin.PriorityNormal)
			return nil
		},
	}
}
