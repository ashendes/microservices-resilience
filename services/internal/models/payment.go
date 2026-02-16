package models

import "time"

// Transaction represents a payment transaction
type Transaction struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// TransactionStatus constants
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
)

// ChargeRequest represents a payment charge request
type ChargeRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

// ChargeResponse represents a payment charge response
type ChargeResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
}
