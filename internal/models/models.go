package models

import (
	"time"
)

type User struct {
	Username string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Balance struct {
	Current   float64 `json:"current" binding:"required"`
	Withdrawn float64 `json:"withdrawn" binding:"required"`
}

type Withdraw struct {
	Order string  `json:"order" binding:"required"`
	Sum   float64 `json:"sum" binding:"required"`
}

type Withdrawals struct {
	Order       string    `json:"order" binding:"required"`
	Sum         float64   `json:"sum" binding:"required"`
	ProcessedAt time.Time `json:"processed_at" binding:"required"`
}

const (
	OrderNEW        = "NEW"
	OrderPROCESSED  = "PROCESSED"
	OrderPROCESSING = "PROCESSING"
	OrderINVALID    = "INVALID"
)

type Order struct {
	Number     string    `json:"number"`
	Username   string    `json:"username"`
	UploadedAt time.Time `json:"uploaded_at"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
}

func NewOrder(username string, number string) *Order {
	return &Order{
		Number:     number,
		Username:   username,
		UploadedAt: time.Now(),
		Status:     OrderNEW,
	}
}
