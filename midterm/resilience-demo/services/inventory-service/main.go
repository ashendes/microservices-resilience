package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/ashendes/resilience-demo/internal/metrics"
	"github.com/ashendes/resilience-demo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// InventoryService manages inventory operations
type InventoryService struct {
	items         map[string]*models.Item
	mutex         sync.RWMutex
	chaosEnabled  bool
	chaosSlowMode bool
	chaosMutex    sync.RWMutex
}

var inventoryService *InventoryService

func init() {
	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize inventory with sample data
	inventoryService = &InventoryService{
		items:        make(map[string]*models.Item),
		chaosEnabled: false,
	}

	// Add sample items
	sampleItems := []*models.Item{
		{ID: "item-1", Name: "Laptop", Quantity: 10000, Price: 999.99},
		{ID: "item-2", Name: "Mouse", Quantity: 50000, Price: 29.99},
		{ID: "item-3", Name: "Keyboard", Quantity: 30000, Price: 79.99},
		{ID: "item-4", Name: "Monitor", Quantity: 15000, Price: 299.99},
		{ID: "item-5", Name: "Headphones", Quantity: 2000, Price: 149.99},
	}

	for _, item := range sampleItems {
		inventoryService.items[item.ID] = item
		// Initialize inventory level metric
		metrics.InventoryLevel.WithLabelValues(item.ID).Set(float64(item.Quantity))
	}
}

func main() {
	router := gin.Default()

	// Add Prometheus middleware
	router.Use(metrics.PrometheusMiddleware("inventory-service"))

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/inventory/status", getStatus)

	// Inventory endpoints
	router.GET("/inventory/check/:itemId", checkInventory)
	router.POST("/inventory/reserve", reserveItems)
	router.POST("/inventory/release", releaseItems)

	// Chaos engineering endpoints
	router.POST("/chaos/inventory/enable", enableChaos)
	router.POST("/chaos/inventory/disable", disableChaos)
	router.POST("/chaos/inventory/slow", enableSlowMode)
	router.POST("/chaos/inventory/slow/disable", disableSlowMode)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Info("Inventory Service starting on port 8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":         "inventory-service",
		"status":          "healthy",
		"chaos_enabled":   inventoryService.getChaosEnabled(),
		"chaos_slow_mode": inventoryService.getSlowMode(),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

func checkInventory(c *gin.Context) {
	itemID := c.Param("itemId")

	// Simulate chaos
	if err := simulateChaos(); err != nil {
		log.WithField("item_id", itemID).Warn("Chaos: Simulated failure")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service temporarily unavailable",
			"message": err.Error(),
		})
		return
	}

	inventoryService.mutex.RLock()
	item, exists := inventoryService.items[itemID]
	inventoryService.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"available": false,
			"quantity":  0,
			"message":   "Item not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.CheckInventoryResponse{
		Available: item.Quantity > 0,
		Quantity:  item.Quantity,
	})
}

func reserveItems(c *gin.Context) {
	var req models.ReserveItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Simulate chaos
	if err := simulateChaos(); err != nil {
		log.WithField("order_id", req.OrderID).Warn("Chaos: Simulated failure during reserve")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Service temporarily unavailable: " + err.Error(),
		})
		return
	}

	inventoryService.mutex.Lock()
	defer inventoryService.mutex.Unlock()

	// Check if all items are available
	for _, orderItem := range req.Items {
		item, exists := inventoryService.items[orderItem.ItemID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Item not found: " + orderItem.ItemID,
			})
			return
		}

		if item.Quantity < orderItem.Quantity {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "Insufficient inventory for item: " + orderItem.ItemID,
			})
			return
		}
	}

	// Reserve items (deduct from inventory)
	for _, orderItem := range req.Items {
		item := inventoryService.items[orderItem.ItemID]
		item.Quantity -= orderItem.Quantity
		// Update metric
		metrics.InventoryLevel.WithLabelValues(item.ID).Set(float64(item.Quantity))
	}

	log.WithFields(log.Fields{
		"order_id": req.OrderID,
		"items":    len(req.Items),
	}).Info("Items reserved successfully")

	c.JSON(http.StatusOK, models.ReserveItemsResponse{
		Success: true,
		Message: "Items reserved successfully",
	})
}

func releaseItems(c *gin.Context) {
	var req models.ReleaseItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	inventoryService.mutex.Lock()
	defer inventoryService.mutex.Unlock()

	// Release items (add back to inventory)
	for _, orderItem := range req.Items {
		item, exists := inventoryService.items[orderItem.ItemID]
		if exists {
			item.Quantity += orderItem.Quantity
			// Update metric
			metrics.InventoryLevel.WithLabelValues(item.ID).Set(float64(item.Quantity))
		}
	}

	log.WithFields(log.Fields{
		"order_id": req.OrderID,
		"items":    len(req.Items),
	}).Info("Items released successfully")

	c.JSON(http.StatusOK, models.ReleaseItemsResponse{
		Success: true,
		Message: "Items released successfully",
	})
}

func enableChaos(c *gin.Context) {
	inventoryService.setChaosEnabled(true)
	metrics.ChaosFailureRate.WithLabelValues("inventory-service").Set(1)

	log.Info("Chaos mode ENABLED for inventory service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Chaos mode enabled",
		"info":    "30% of requests will fail randomly",
	})
}

func disableChaos(c *gin.Context) {
	inventoryService.setChaosEnabled(false)
	inventoryService.setSlowMode(false)
	metrics.ChaosFailureRate.WithLabelValues("inventory-service").Set(0)
	metrics.ChaosSlowMode.WithLabelValues("inventory-service").Set(0)

	log.Info("Chaos mode DISABLED for inventory service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Chaos mode disabled",
	})
}

func enableSlowMode(c *gin.Context) {
	inventoryService.setSlowMode(true)
	metrics.ChaosSlowMode.WithLabelValues("inventory-service").Set(1)

	log.Info("Slow mode ENABLED for inventory service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Slow mode enabled",
		"info":    "Requests will have 2-5 second delays",
	})
}

func disableSlowMode(c *gin.Context) {
	inventoryService.setSlowMode(false)
	metrics.ChaosSlowMode.WithLabelValues("inventory-service").Set(0)

	log.Info("Slow mode DISABLED for inventory service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Slow mode disabled",
	})
}

// Helper methods
func (is *InventoryService) setChaosEnabled(enabled bool) {
	is.chaosMutex.Lock()
	defer is.chaosMutex.Unlock()
	is.chaosEnabled = enabled
}

func (is *InventoryService) getChaosEnabled() bool {
	is.chaosMutex.RLock()
	defer is.chaosMutex.RUnlock()
	return is.chaosEnabled
}

func (is *InventoryService) setSlowMode(enabled bool) {
	is.chaosMutex.Lock()
	defer is.chaosMutex.Unlock()
	is.chaosSlowMode = enabled
}

func (is *InventoryService) getSlowMode() bool {
	is.chaosMutex.RLock()
	defer is.chaosMutex.RUnlock()
	return is.chaosSlowMode
}

func simulateChaos() error {
	// Check if slow mode is enabled
	if inventoryService.getSlowMode() {
		delay := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
		log.WithField("delay_ms", delay.Milliseconds()).Debug("Chaos: Simulating slow response")
		time.Sleep(delay)
	}

	// Check if failure mode is enabled
	if inventoryService.getChaosEnabled() {
		// 30% failure rate
		if rand.Float32() < 0.3 {
			return gin.Error{Err: http.ErrAbortHandler, Type: gin.ErrorTypePublic}
		}
	}

	return nil
}
