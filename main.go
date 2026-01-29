package main

import (
	"Kasir-API/config"
	"Kasir-API/database"
	"Kasir-API/handlers"
	"Kasir-API/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
	"time"
)

func main() {
	config.Init()

	// Log Railway environment info
	log.Println("🚀 Starting Category API on Railway...")

	if url := viper.GetString("RAILWAY_PUBLIC_URL"); url != "" {
		log.Printf("🌐 Public URL: %s", url)
	}

	if env := viper.GetString("RAILWAY_ENVIRONMENT"); env != "" {
		log.Printf("🏭 Environment: %s", env)
	}

	ginMode := viper.GetString("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug" // default
	}

	gin.SetMode(ginMode)

	if ginMode == "release" {
		log.Println("PRODUCTION mode activated")
	} else {
		log.Println("DEBUG mode activated")
	}

	// Initialize database
	database.ConnectDatabase()

	// Create router
	router := gin.New()

	// Custom logger yang skip /metrics
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Skip logging for /metrics endpoint
		if param.Path == "/metrics" {
			return ""
		}

		// Custom log format untuk endpoint lainnya
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))

	router.SetTrustedProxies(nil)

	// Recovery middleware
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(middleware.CORS())

	// Health check routes
	router.GET("/health", handlers.HealthCheck)
	router.GET("/health/db", handlers.HealthCheckDB)

	// Metrics endpoint (simple version - no logs)
	router.GET("/metrics", func(c *gin.Context) {
		// Return minimal response tanpa log
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "category-api",
			"timestamp": time.Now().Unix(),
		})
	})

	// Root route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Category API is running",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"GET /":                  "API info",
				"GET /health":            "Basic health check",
				"GET /health/db":         "Database health check",
				"GET /metrics":           "Metrics endpoint",
				"GET /categories":        "Get all categories",
				"POST /categories":       "Create new category",
				"GET /categories/:id":    "Get category by ID",
				"PUT /categories/:id":    "Update category",
				"DELETE /categories/:id": "Delete category",
				"GET /products":          "Get all products",
				"POST /products":         "Create new product",
				"GET /products/:id":      "Get product by ID",
				"PUT /products/:id":      "Update product",
				"DELETE /products/:id":   "Delete product",
			},
		})
	})

	// Category routes
	categoryRoutes := router.Group("/categories")
	{
		categoryRoutes.GET("/", handlers.GetAllCategories)
		categoryRoutes.POST("/", handlers.CreateCategory)
		categoryRoutes.GET("/:id", handlers.GetCategoryByID)
		categoryRoutes.PUT("/:id", handlers.UpdateCategory)
		categoryRoutes.DELETE("/:id", handlers.DeleteCategory)
	}

	productRoutes := router.Group("/products")
	{
		productRoutes.GET("/", handlers.GetAllProducts)
		productRoutes.POST("/", handlers.CreateProduct)
		productRoutes.GET("/:id", handlers.GetProductByID)
		productRoutes.PUT("/:id", handlers.UpdateProduct)
		productRoutes.DELETE("/:id", handlers.DeleteProduct)
	}

	// Start server
	port := viper.GetString("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)

	router.Run(":" + port)
}
