package repositories

import (
	"Kasir-API/models"
	"fmt"

	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(transaction *models.Transaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Save Transaction Header
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// 2. Process each detail (Save and Update Stock)
		for _, detail := range transaction.Details {
			// Update product stock
			var product models.Product
			if err := tx.First(&product, detail.ProductID).Error; err != nil {
				return fmt.Errorf("product with ID %d not found", detail.ProductID)
			}

			if product.Stock < detail.Quantity {
				return fmt.Errorf("insufficient stock for product: %s", product.Name)
			}

			// Reduce stock
			newStock := product.Stock - detail.Quantity
			if err := tx.Model(&product).Update("stock", newStock).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *TransactionRepository) GetAll() ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Preload("Details").Find(&transactions).Error
	return transactions, err
}

func (r *TransactionRepository) GetReport(startDate, endDate string) (models.ReportResponse, error) {
	var report models.ReportResponse

	// 1. Get Total Revenue and Total Transaksi
	row := r.db.Model(&models.Transaction{}).
		Select("SUM(total_amount) as total_revenue, COUNT(id) as total_transaksi").
		Where("created_at >= ? AND created_at <= ?", startDate+" 00:00:00", endDate+" 23:59:59").
		Row()

	var totalRevenue *int
	var totalTransaksi int
	if err := row.Scan(&totalRevenue, &totalTransaksi); err != nil {
		return report, err
	}

	if totalRevenue != nil {
		report.TotalRevenue = *totalRevenue
	}
	report.TotalTransaksi = totalTransaksi

	// 2. Get Produk Terlaris
	var bestProduct models.BestSellingProduct
	err := r.db.Model(&models.TransactionDetail{}).
		Select("product_name as name, SUM(quantity) as qty_sold").
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.created_at >= ? AND transactions.created_at <= ?", startDate+" 00:00:00", endDate+" 23:59:59").
		Group("product_name").
		Order("qty_sold DESC").
		Limit(1).
		Scan(&bestProduct).Error

	if err == nil {
		report.ProdukTerlaris = bestProduct
	}

	return report, nil
}
