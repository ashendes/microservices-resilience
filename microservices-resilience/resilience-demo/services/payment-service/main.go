package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/ashendes/resilience-demo/internal/metrics"
	"github.com/ashendes/resilience-demo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// PaymentService manages payment operations
type PaymentService struct {
	transactions  map[string]*models.Transaction
	mutex         sync.RWMutex
	chaosEnabled  bool
	chaosSlowMode bool
	chaosMutex    sync.RWMutex
}

var paymentService *PaymentService

func init() {
	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize payment service
	paymentService = &PaymentService{
		transactions:  make(map[string]*models.Transaction),
		chaosEnabled:  false,
		chaosSlowMode: false,
	}
}

func main() {
	router := gin.Default()

	// Add Prometheus middleware
	router.Use(metrics.PrometheusMiddleware("payment-service"))

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/payment/status", getStatus)

	// Payment endpoints
	router.POST("/payment/charge", chargePayment)

	// Chaos engineering endpoints
	router.POST("/chaos/payment/enable", enableChaos)
	router.POST("/chaos/payment/disable", disableChaos)
	router.POST("/chaos/payment/slow", enableSlowMode)
	router.POST("/chaos/payment/slow/disable", disableSlowMode)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Info("Payment Service starting on port 8082")
	if err := router.Run(":8082"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":         "payment-service",
		"status":          "healthy",
		"chaos_enabled":   paymentService.getChaosEnabled(),
		"chaos_slow_mode": paymentService.getSlowMode(),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

func chargePayment(c *gin.Context) {
	var req models.ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"transaction_id": "",
			"status":         models.TransactionStatusFailed,
			"message":        "Invalid request: " + err.Error(),
		})
		return
	}

	// Simulate chaos
	if err := simulateChaos(); err != nil {
		log.WithFields(log.Fields{
			"order_id": req.OrderID,
			"amount":   req.Amount,
		}).Warn("Chaos: Simulated payment failure")

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"transaction_id": "",
			"status":         models.TransactionStatusFailed,
			"message":        "Payment service temporarily unavailable: " + err.Error(),
		})
		return
	}

	// Create transaction
	transactionID := uuid.New().String()
	transaction := &models.Transaction{
		ID:        transactionID,
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Status:    models.TransactionStatusCompleted,
		Timestamp: time.Now(),
	}

	paymentService.mutex.Lock()
	paymentService.transactions[transactionID] = transaction
	paymentService.mutex.Unlock()

	// Record payment amount metric
	metrics.PaymentAmount.Observe(req.Amount)

	log.WithFields(log.Fields{
		"transaction_id": transactionID,
		"order_id":       req.OrderID,
		"amount":         req.Amount,
	}).Info("Payment processed successfully")

	c.JSON(http.StatusOK, models.ChargeResponse{
		TransactionID: transactionID,
		Status:        models.TransactionStatusCompleted,
		Message:       "Payment processed successfully",
	})
}

func enableChaos(c *gin.Context) {
	paymentService.setChaosEnabled(true)
	metrics.ChaosFailureRate.WithLabelValues("payment-service").Set(1)

	log.Info("Chaos mode ENABLED for payment service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Chaos mode enabled",
		"info":    "40% of requests will fail randomly",
	})
}

func disableChaos(c *gin.Context) {
	paymentService.setChaosEnabled(false)
	paymentService.setSlowMode(false)
	metrics.ChaosFailureRate.WithLabelValues("payment-service").Set(0)
	metrics.ChaosSlowMode.WithLabelValues("payment-service").Set(0)

	log.Info("Chaos mode DISABLED for payment service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Chaos mode disabled",
	})
}

func enableSlowMode(c *gin.Context) {
	paymentService.setSlowMode(true)
	metrics.ChaosSlowMode.WithLabelValues("payment-service").Set(1)

	log.Info("Slow mode ENABLED for payment service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Slow mode enabled",
		"info":    "Requests will have 5-10 second delays",
	})
}

func disableSlowMode(c *gin.Context) {
	paymentService.setSlowMode(false)
	metrics.ChaosSlowMode.WithLabelValues("payment-service").Set(0)

	log.Info("Slow mode DISABLED for payment service")
	c.JSON(http.StatusOK, gin.H{
		"message": "Slow mode disabled",
	})
}

// Helper methods
func (ps *PaymentService) setChaosEnabled(enabled bool) {
	ps.chaosMutex.Lock()
	defer ps.chaosMutex.Unlock()
	ps.chaosEnabled = enabled
}

func (ps *PaymentService) getChaosEnabled() bool {
	ps.chaosMutex.RLock()
	defer ps.chaosMutex.RUnlock()
	return ps.chaosEnabled
}

func (ps *PaymentService) setSlowMode(enabled bool) {
	ps.chaosMutex.Lock()
	defer ps.chaosMutex.Unlock()
	ps.chaosSlowMode = enabled
}

func (ps *PaymentService) getSlowMode() bool {
	ps.chaosMutex.RLock()
	defer ps.chaosMutex.RUnlock()
	return ps.chaosSlowMode
}

func simulateChaos() error {
	// Check if slow mode is enabled
	if paymentService.getSlowMode() {
		delay := time.Duration(5000+rand.Intn(5000)) * time.Millisecond
		log.WithField("delay_ms", delay.Milliseconds()).Debug("Chaos: Simulating slow response")
		time.Sleep(delay)
	}

	// Check if failure mode is enabled
	if paymentService.getChaosEnabled() {
		// 40% failure rate
		if rand.Float32() < 0.4 {
			return gin.Error{Err: http.ErrAbortHandler, Type: gin.ErrorTypePublic}
		}
	}

	return nil
}
