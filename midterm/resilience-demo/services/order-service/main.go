package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ashendes/resilience-demo/internal/metrics"
	"github.com/ashendes/resilience-demo/internal/models"
	"github.com/ashendes/resilience-demo/internal/patterns"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// OrderService manages order operations
type OrderService struct {
	orders              map[string]*models.Order
	mutex               sync.RWMutex
	inventoryClient     *resty.Client
	paymentClient       *resty.Client
	inventoryCircuit    *patterns.CircuitBreakerWrapper
	paymentCircuit      *patterns.CircuitBreakerWrapper
	inventoryBulkhead   *patterns.Bulkhead
	paymentBulkhead     *patterns.Bulkhead
	inventoryServiceURL string
	paymentServiceURL   string
}

var orderService *OrderService

func init() {
	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func main() {
	// Get service URLs from environment or use defaults
	inventoryServiceURL := getEnv("INVENTORY_SERVICE_URL", "http://localhost:8081")
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8082")

	// Initialize order service with resilience patterns
	orderService = &OrderService{
		orders: make(map[string]*models.Order),
		inventoryClient: resty.New().
			SetTimeout(patterns.DefaultTimeout).
			SetRetryCount(0), // No automatic retries, we handle via circuit breaker
		paymentClient: resty.New().
			SetTimeout(patterns.DefaultTimeout).
			SetRetryCount(0),
		inventoryCircuit:    patterns.NewCircuitBreaker("Inventory", "order-service"),
		paymentCircuit:      patterns.NewCircuitBreaker("Payment", "order-service"),
		inventoryBulkhead:   patterns.NewBulkhead(10, "inventory", "order-service"),
		paymentBulkhead:     patterns.NewBulkhead(10, "payment", "order-service"),
		inventoryServiceURL: inventoryServiceURL,
		paymentServiceURL:   paymentServiceURL,
	}

	router := gin.Default()

	// Add Prometheus middleware
	router.Use(metrics.PrometheusMiddleware("order-service"))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Order endpoints
	router.POST("/order/create", createOrder)
	router.GET("/order/:orderId", getOrder)
	router.GET("/order/circuit-status", getCircuitStatus)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.WithFields(log.Fields{
		"inventory_url": inventoryServiceURL,
		"payment_url":   paymentServiceURL,
	}).Info("Order Service starting on port 8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

// createOrder handles order creation with full resilience patterns
func createOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.OrdersTotal.WithLabelValues("validation_failed").Inc()
		c.JSON(http.StatusBadRequest, models.CreateOrderResponse{
			Status:  models.OrderStatusFailed,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// FAIL FAST: Validate order
	if err := validateOrder(&req); err != nil {
		metrics.OrdersTotal.WithLabelValues("validation_failed").Inc()
		c.JSON(http.StatusBadRequest, models.CreateOrderResponse{
			Status:  models.OrderStatusFailed,
			Message: "Validation failed: " + err.Error(),
		})
		return
	}

	// Create order
	orderID := uuid.New().String()
	order := &models.Order{
		ID:        orderID,
		Items:     req.Items,
		Status:    models.OrderStatusPending,
		Timestamp: time.Now(),
	}

	// Calculate total
	totalAmount := 0.0
	for _, item := range req.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = totalAmount

	// Store order
	orderService.mutex.Lock()
	orderService.orders[orderID] = order
	orderService.mutex.Unlock()

	log.WithFields(log.Fields{
		"order_id": orderID,
		"items":    len(req.Items),
		"total":    totalAmount,
	}).Info("Processing new order")

	// Process order with resilience patterns
	if err := orderService.processOrder(order); err != nil {
		order.Status = models.OrderStatusFailed
		metrics.OrdersTotal.WithLabelValues("failed").Inc()

		c.JSON(http.StatusInternalServerError, models.CreateOrderResponse{
			OrderID: orderID,
			Status:  models.OrderStatusFailed,
			Message: fmt.Sprintf("Order processing failed: %v", err),
			Total:   totalAmount,
		})
		return
	}

	order.Status = models.OrderStatusCompleted
	metrics.OrdersTotal.WithLabelValues("completed").Inc()

	log.WithField("order_id", orderID).Info("Order completed successfully")

	c.JSON(http.StatusOK, models.CreateOrderResponse{
		OrderID: orderID,
		Status:  models.OrderStatusCompleted,
		Message: "Order processed successfully",
		Total:   totalAmount,
	})
}

// getOrder retrieves order details
func getOrder(c *gin.Context) {
	orderID := c.Param("orderId")

	orderService.mutex.RLock()
	order, exists := orderService.orders[orderID]
	orderService.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":    "Order not found",
			"order_id": orderID,
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// getCircuitStatus returns the status of circuit breakers
func getCircuitStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"inventory_circuit": gin.H{
			"name":  "Inventory",
			"state": orderService.inventoryCircuit.GetState(),
			"value": orderService.inventoryCircuit.GetStateValue(),
		},
		"payment_circuit": gin.H{
			"name":  "Payment",
			"state": orderService.paymentCircuit.GetState(),
			"value": orderService.paymentCircuit.GetStateValue(),
		},
	})
}

// processOrder orchestrates the order processing with resilience patterns
func (os *OrderService) processOrder(order *models.Order) error {
	// Step 1: Reserve inventory with Circuit Breaker + Bulkhead
	if err := os.reserveInventory(order); err != nil {
		log.WithField("order_id", order.ID).Error("Failed to reserve inventory: ", err)
		return fmt.Errorf("inventory reservation failed: %w", err)
	}

	// Step 2: Process payment with Circuit Breaker + Bulkhead
	if err := os.processPayment(order); err != nil {
		log.WithField("order_id", order.ID).Error("Payment failed, releasing inventory: ", err)

		// Rollback: Release inventory
		if releaseErr := os.releaseInventory(order); releaseErr != nil {
			log.WithField("order_id", order.ID).Error("Failed to release inventory during rollback: ", releaseErr)
		}

		return fmt.Errorf("payment processing failed: %w", err)
	}

	return nil
}

// reserveInventory reserves items with circuit breaker and bulkhead patterns
func (os *OrderService) reserveInventory(order *models.Order) error {
	reserveRequest := models.ReserveItemsRequest{
		OrderID: order.ID,
		Items:   order.Items,
	}

	// Execute with bulkhead pattern
	err := os.inventoryBulkhead.Execute(func() error {
		// Execute with circuit breaker pattern
		_, cbErr := os.inventoryCircuit.Execute(func() (interface{}, error) {
			resp, httpErr := os.inventoryClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(reserveRequest).
				Post(os.inventoryServiceURL + "/inventory/reserve")

			if httpErr != nil {
				return nil, fmt.Errorf("HTTP error: %w", httpErr)
			}

			if resp.StatusCode() != http.StatusOK {
				return nil, fmt.Errorf("inventory service returned status %d: %s", resp.StatusCode(), resp.String())
			}

			var response models.ReserveItemsResponse
			if err := json.Unmarshal(resp.Body(), &response); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			if !response.Success {
				return nil, fmt.Errorf("reservation failed: %s", response.Message)
			}

			return response, nil
		})

		return patterns.FormatError("Inventory", cbErr)
	})

	return err
}

// processPayment processes payment with circuit breaker and bulkhead patterns
func (os *OrderService) processPayment(order *models.Order) error {
	chargeRequest := models.ChargeRequest{
		OrderID: order.ID,
		Amount:  order.TotalAmount,
	}

	// Execute with bulkhead pattern
	err := os.paymentBulkhead.Execute(func() error {
		// Execute with circuit breaker pattern
		_, cbErr := os.paymentCircuit.Execute(func() (interface{}, error) {
			resp, httpErr := os.paymentClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(chargeRequest).
				Post(os.paymentServiceURL + "/payment/charge")

			if httpErr != nil {
				return nil, fmt.Errorf("HTTP error: %w", httpErr)
			}

			if resp.StatusCode() != http.StatusOK {
				return nil, fmt.Errorf("payment service returned status %d: %s", resp.StatusCode(), resp.String())
			}

			var response models.ChargeResponse
			if err := json.Unmarshal(resp.Body(), &response); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			if response.Status != models.TransactionStatusCompleted {
				return nil, fmt.Errorf("payment failed: %s", response.Message)
			}

			return response, nil
		})

		return patterns.FormatError("Payment", cbErr)
	})

	return err
}

// releaseInventory releases reserved inventory (rollback operation)
func (os *OrderService) releaseInventory(order *models.Order) error {
	releaseRequest := models.ReleaseItemsRequest{
		OrderID: order.ID,
		Items:   order.Items,
	}

	resp, err := os.inventoryClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(releaseRequest).
		Post(os.inventoryServiceURL + "/inventory/release")

	if err != nil {
		return fmt.Errorf("HTTP error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("inventory service returned status %d", resp.StatusCode())
	}

	return nil
}

// validateOrder performs fail-fast validation
func validateOrder(req *models.CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	for i, item := range req.Items {
		if item.ItemID == "" {
			return fmt.Errorf("item %d: item_id is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be greater than 0", i)
		}
		if item.Price <= 0 {
			return fmt.Errorf("item %d: price must be greater than 0", i)
		}
	}

	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
