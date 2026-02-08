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
