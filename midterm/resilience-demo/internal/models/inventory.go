package models

// Item represents an inventory item
type Item struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// CheckInventoryRequest represents a request to check inventory
type CheckInventoryRequest struct {
	ItemID string `json:"item_id" binding:"required"`
}

// CheckInventoryResponse represents inventory check response
type CheckInventoryResponse struct {
	Available bool   `json:"available"`
	Quantity  int    `json:"quantity"`
	Message   string `json:"message,omitempty"`
}

// ReserveItemsRequest represents a request to reserve inventory
type ReserveItemsRequest struct {
	OrderID string      `json:"order_id" binding:"required"`
	Items   []OrderItem `json:"items" binding:"required,dive"`
}

// ReserveItemsResponse represents the response after reserving items
type ReserveItemsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ReleaseItemsRequest represents a request to release reserved inventory
type ReleaseItemsRequest struct {
	OrderID string      `json:"order_id" binding:"required"`
	Items   []OrderItem `json:"items" binding:"required,dive"`
}

// ReleaseItemsResponse represents the response after releasing items
type ReleaseItemsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
