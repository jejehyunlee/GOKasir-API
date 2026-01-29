package database

import (
	"Kasir-API/models"
	"database/sql"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var sqlDB *sql.DB // Tambahkan ini

func ConnectDatabase() {
	// Prioritize DATABASE_URL from environment
	databaseURL := viper.GetString("DATABASE_URL")

	if databaseURL == "" {
		// Fallback to individual variables
		dbHost := viper.GetString("DB_HOST")
		dbPort := viper.GetString("DB_PORT")
		dbUser := viper.GetString("DB_USER")
		dbPassword := viper.GetString("DB_PASSWORD")
		dbName := viper.GetString("DB_NAME")
		dbSSLMode := viper.GetString("DB_SSLMODE")

		if dbSSLMode == "" {
			dbSSLMode = "disable"
		}

		databaseURL = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
			dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
		)
	}

	log.Printf("Connecting to database...")

	var err error

	// ==================== GORM CONFIGURATION ====================
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Silent), // Production
		Logger:                 logger.Default.LogMode(logger.Info), // Development
		PrepareStmt:            true,                                // Enable prepared statement cache
		SkipDefaultTransaction: true,                                // Improve performance
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// ==================== DATABASE POOLING CONFIGURATION ====================
	// Get underlying sql.DB
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}

	// Untuk beban ~95-200 req/detik & 200 VUs:
	sqlDB.SetMaxOpenConns(100)                 // Naikkan sesuai kapasitas DB Railway (bisa 100-500)
	sqlDB.SetMaxIdleConns(50)                  // 50% dari MaxOpenConns
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Lebih panjang untuk hindari turnover koneksi
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Oke
	log.Println("✅ Database connected successfully!")
	log.Printf("📊 Connection Pool Stats: MaxOpen=%d, MaxIdle=%d", 25, 10)

	// ==================== AUTO MIGRATE ====================
	err = DB.AutoMigrate(&models.Category{}, &models.Product{})
	if err != nil {
		log.Printf("⚠️ Warning: AutoMigrate failed: %v", err)
	} else {
		log.Println("✅ Database migration completed")
	}

	// ==================== TEST CONNECTION & POOL ====================
	testConnection()

	// Monitor pool stats secara periodic (opsional)
	go monitorConnectionPool()
}

func testConnection() {
	var result int
	if err := DB.Raw("SELECT 1").Scan(&result).Error; err != nil {
		log.Printf("❌ Database test query failed: %v", err)
	} else {
		log.Printf("✅ Database test query successful: %d", result)
	}

	// Test pool stats
	if sqlDB != nil {
		stats := sqlDB.Stats()
		log.Printf("📈 Initial Pool Stats: InUse=%d, Idle=%d, OpenConnections=%d",
			stats.InUse, stats.Idle, stats.OpenConnections)
	}
}

// ==================== MONITORING FUNCTION ====================
func monitorConnectionPool() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if sqlDB != nil {
			stats := sqlDB.Stats()
			log.Printf("📊 Pool Monitor: InUse=%d, Idle=%d, WaitCount=%d, WaitDuration=%v",
				stats.InUse, stats.Idle, stats.WaitCount, stats.WaitDuration)

			// Warning jika terlalu banyak wait
			if stats.WaitCount > 100 {
				log.Printf("⚠️ HIGH WAIT COUNT: %d - Consider increasing MaxOpenConns", stats.WaitCount)
			}
		}
	}
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// GetSQLDB returns the underlying sql.DB for manual pooling control
func GetSQLDB() (*sql.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return DB.DB()
}

// GetPoolStats returns current connection pool statistics
func GetPoolStats() string {
	if sqlDB == nil {
		return "Pool not initialized"
	}

	stats := sqlDB.Stats()
	return fmt.Sprintf(
		"MaxOpen: %d, InUse: %d, Idle: %d, Open: %d, WaitCount: %d",
		25, stats.InUse, stats.Idle, stats.OpenConnections, stats.WaitCount,
	)
}

// HealthCheck untuk monitoring
func HealthCheck() bool {
	if sqlDB == nil {
		return false
	}

	// Ping database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Printf("❌ Database health check failed: %v", err)
		return false
	}

	// Cek pool health
	stats := sqlDB.Stats()
	if stats.OpenConnections >= 20 { // 80% dari MaxOpenConns
		log.Printf("⚠️ Pool near capacity: %d/%d connections",
			stats.OpenConnections, 25)
	}

	return true
}
