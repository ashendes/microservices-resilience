package models

import "time"

// OrderItem represents an item in an order
type OrderItem struct {
	ItemID   string  `json:"item_id" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price" binding:"required,gt=0"`
}

// Order represents a customer order
type Order struct {
	ID          string      `json:"id"`
	Items       []OrderItem `json:"items" binding:"required,dive"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"`
	Timestamp   time.Time   `json:"timestamp"`
}

// OrderStatus constants
const (
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"
	OrderStatusFailed    = "failed"
	OrderStatusCancelled = "cancelled"
)

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	Items []OrderItem `json:"items" binding:"required,dive"`
}

// CreateOrderResponse represents the response after creating an order
type CreateOrderResponse struct {
	OrderID string  `json:"order_id"`
	Status  string  `json:"status"`
	Message string  `json:"message,omitempty"`
	Total   float64 `json:"total"`
}
