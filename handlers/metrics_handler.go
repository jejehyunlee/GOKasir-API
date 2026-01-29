package handlers

import (
	"Kasir-API/database"
	"Kasir-API/utils"
	"github.com/gin-gonic/gin"
	"time"
)

// HealthCheck - GET /health
func HealthCheck(c *gin.Context) {
	utils.Success(c, "Service is healthy", gin.H{
		"timestamp": time.Now().Unix(),
		"service":   "category-api",
		"version":   "1.0.0",
		"status":    "operational",
	})
}

// HealthCheckDB - GET /health/db
func HealthCheckDB(c *gin.Context) {
	if database.GetDB() == nil {
		utils.Error(c, 503, "Database not initialized", gin.H{
			"status": "disconnected",
		})
		return
	}

	// Test database connection
	var result int
	err := database.GetDB().Raw("SELECT 1").Scan(&result).Error

	if err != nil {
		utils.Error(c, 503, "Database connection failed", gin.H{
			"status":  "disconnected",
			"details": err.Error(),
		})
		return
	}

	// Get database stats
	var categoryCount int64
	var productCount int64
	database.GetDB().Table("categories").Count(&categoryCount)
	database.GetDB().Table("products").Count(&productCount)

	utils.Success(c, "Database is healthy", gin.H{
		"status":           "connected",
		"timestamp":        time.Now().Unix(),
		"connection_test":  "successful",
		"categories_count": categoryCount,
		"products_count":   productCount,
		"query_result":     result,
	})
}

// ... (getUptime function tetap sama) ...
