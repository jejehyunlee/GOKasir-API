package repositories

import (
	"Kasir-API/models"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetAll(nameFilter string) ([]models.Product, error) {
	var products []models.Product

	query := r.db.Select("id", "name", "price", "stock", "category_id", "created_at", "updated_at").
		Preload("Category", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		})

	if nameFilter != "" {
		query = query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}

	err := query.Find(&products).Error
	return products, err
}
