package handlers

import (
	"Kasir-API/database"
	"Kasir-API/models"
	"Kasir-API/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllProducts(c *gin.Context) {
	var products []models.Product

	// Optimized: Select only necessary fields and preload category efficiently
	if err := database.GetDB().
		Select("id", "name", "price", "stock", "category_id", "created_at", "updated_at").
		Preload("Category", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Find(&products).Error; err != nil {
		utils.InternalServerError(c, "Failed to fetch products", err.Error())
		return
	}

	if len(products) == 0 {
		utils.Success(c, "No products found", []interface{}{})
		return
	}

	// Bersihkan category yang kosong
	for i := range products {
		if products[i].Category != nil && products[i].Category.ID == 0 {
			products[i].Category = nil
		}
	}

	utils.Success(c, "Products retrieved successfully", products)
}

func GetProductByID(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.GetDB().Preload("Category").First(&product, id).Error; err != nil {
		utils.NotFound(c, "Product")
		return
	}

	utils.Success(c, "Product retrieved successfully", product)
}

// handlers/product_handler.go
func CreateProduct(c *gin.Context) {
	var input struct {
		Name       string  `json:"name" binding:"required,min=3,max=100"`
		Price      float64 `json:"price" binding:"required,gt=0"`
		Stock      int     `json:"stock" binding:"required,gte=0"`
		CategoryID uint    `json:"category_id" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, "Validation error", err.Error())
		return
	}

	// Cek apakah kategori ada
	var category models.Category
	if err := database.GetDB().First(&category, input.CategoryID).Error; err != nil {
		utils.BadRequest(c, "Invalid category_id", nil)
		return
	}

	// Buat pointer untuk CategoryID
	//categoryIDPtr := &input.CategoryID

	product := models.Product{
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
		CategoryID: input.CategoryID, // <-- SEKARANG pakai pointer
	}

	if err := database.GetDB().Create(&product).Error; err != nil {
		utils.InternalServerError(c, "Failed to create product", err.Error())
		return
	}

	// Preload category untuk response
	if err := database.GetDB().Preload("Category").First(&product, product.ID).Error; err != nil {
		utils.InternalServerError(c, "Failed to fetch created product", err.Error())
		return
	}

	// Bersihkan category jika NULL
	if product.Category != nil && product.Category.ID == 0 {
		product.Category = nil
	}

	utils.Created(c, "Product created successfully", product)
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.GetDB().First(&product, id).Error; err != nil {
		utils.NotFound(c, "Product")
		return
	}

	var input struct {
		Name       string  `json:"name" binding:"omitempty,min=3,max=100"`
		Price      float64 `json:"price" binding:"omitempty,gt=0"`
		Stock      *int    `json:"stock" binding:"omitempty,gte=0"`
		CategoryID *uint   `json:"category_id" binding:"omitempty,gt=0"` // <-- MASIH pointer
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, "Validation error", err.Error())
		return
	}

	// Update fields
	updates := make(map[string]interface{})

	if input.Name != "" {
		updates["name"] = input.Name
	}

	if input.Price != 0 {
		updates["price"] = input.Price
	}

	if input.Stock != nil {
		updates["stock"] = *input.Stock
	}

	if input.CategoryID != nil {
		// Cek apakah kategori ada
		var category models.Category
		if err := database.GetDB().First(&category, *input.CategoryID).Error; err != nil {
			utils.BadRequest(c, "Invalid category_id", nil)
			return
		}
		// Simpan sebagai pointer
		updates["category_id"] = input.CategoryID
	}

	// Update product
	if err := database.GetDB().Model(&product).Updates(updates).Error; err != nil {
		utils.InternalServerError(c, "Failed to update product", err.Error())
		return
	}

	// Reload dengan category
	if err := database.GetDB().Preload("Category").First(&product, product.ID).Error; err != nil {
		utils.InternalServerError(c, "Failed to fetch updated product", err.Error())
		return
	}

	// Bersihkan category jika NULL
	if product.Category != nil && product.Category.ID == 0 {
		product.Category = nil
	}

	utils.Success(c, "Product updated successfully", product)
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.GetDB().First(&product, id).Error; err != nil {
		utils.NotFound(c, "Product")
		return
	}

	if err := database.GetDB().Delete(&product).Error; err != nil {
		utils.InternalServerError(c, "Failed to delete product", err.Error())
		return
	}

	utils.Success(c, "Product deleted successfully", gin.H{
		"id": product.ID,
	})
}
