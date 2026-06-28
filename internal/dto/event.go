package dto

import wasmplugin "github.com/SuperBotForge/sdk/go-sdk"

type OrderEvent struct {
	UserID      int64
	OrderID     int64
	OrderStatus string
	MessageKey  string
	Locale      string
	File        *wasmplugin.FileRef
}
