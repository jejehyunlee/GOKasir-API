package handlers

import (
	"Kasir-API/database"
	"Kasir-API/models"
	"Kasir-API/utils"
	"sync"

	"github.com/gin-gonic/gin"
)

// Cache untuk kategori (opsional, jika kategori jarang berubah)
var (
	categoryCache     = make(map[uint]bool)
	categoryCacheLock sync.RWMutex
)

func GetAllProducts(c *gin.Context) {
	var products []models.Product

	if err := database.GetDB().Preload("Category").Find(&products).Error; err != nil {
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

// Fungsi helper untuk validasi kategori dengan cache
func validateCategory(categoryID uint) bool {
	categoryCacheLock.RLock()
	exists, found := categoryCache[categoryID]
	categoryCacheLock.RUnlock()

	if found {
		return exists
	}

	// Cek di database
	var count int64
	database.GetDB().Model(&models.Category{}).Where("id = ?", categoryID).Count(&count)

	exists = count > 0

	// Update cache
	categoryCacheLock.Lock()
	categoryCache[categoryID] = exists
	categoryCacheLock.Unlock()

	return exists
}

// Clear cache kategori (bisa dipanggil saat ada perubahan kategori)
func ClearCategoryCache() {
	categoryCacheLock.Lock()
	categoryCache = make(map[uint]bool)
	categoryCacheLock.Unlock()
}

// CreateProduct - Versi Optimized
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

	// OPSI 1: Validasi kategori dengan cache (lebih cepat)
	if !validateCategory(input.CategoryID) {
		utils.BadRequest(c, "Invalid category_id", nil)
		return
	}

	// OPSI 2: Atau langsung insert tanpa validasi terlebih dahulu
	// (lebih cepat, tapi bisa error foreign key constraint dari database)
	// Jika yakin category_id valid dari client, bisa lewati validasi

	// Mulai transaction
	tx := database.GetDB().Begin()
	if tx.Error != nil {
		utils.InternalServerError(c, "Failed to start transaction", tx.Error.Error())
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create product dalam transaction
	product := models.Product{
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
		CategoryID: input.CategoryID,
	}

	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		// Jika error foreign key, beri pesan khusus
		if utils.IsForeignKeyError(err) {
			utils.BadRequest(c, "Invalid category_id", "Category does not exist")
			return
		}
		utils.InternalServerError(c, "Failed to create product", err.Error())
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.InternalServerError(c, "Failed to commit transaction", err.Error())
		return
	}

	// Response minimal - tanpa preload category (lebih cepat)
	response := gin.H{
		"id":          product.ID,
		"name":        product.Name,
		"price":       product.Price,
		"stock":       product.Stock,
		"category_id": product.CategoryID,
		"created_at":  product.CreatedAt,
		"updated_at":  product.UpdatedAt,
	}

	utils.Created(c, "Product created successfully", response)
}

// CreateProductWithCategoryResponse - Versi dengan category di response (sedikit lebih lambat)
func CreateProductWithCategoryResponse(c *gin.Context) {
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

	// Mulai transaction
	tx := database.GetDB().Begin()
	if tx.Error != nil {
		utils.InternalServerError(c, "Failed to start transaction", tx.Error.Error())
		return
	}

	// 1. Cek kategori dalam transaction
	var category models.Category
	if err := tx.First(&category, input.CategoryID).Error; err != nil {
		tx.Rollback()
		utils.BadRequest(c, "Invalid category_id", nil)
		return
	}

	// 2. Create product
	product := models.Product{
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
		CategoryID: input.CategoryID,
		Category:   &category, // Set langsung dari memory
	}

	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		utils.InternalServerError(c, "Failed to create product", err.Error())
		return
	}

	// 3. Commit
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.InternalServerError(c, "Failed to commit transaction", err.Error())
		return
	}

	utils.Created(c, "Product created successfully", product)
}

// BatchCreateProducts - Untuk load testing yang lebih efisien
func BatchCreateProducts(c *gin.Context) {
	var inputs []struct {
		Name       string  `json:"name" binding:"required,min=3,max=100"`
		Price      float64 `json:"price" binding:"required,gt=0"`
		Stock      int     `json:"stock" binding:"required,gte=0"`
		CategoryID uint    `json:"category_id" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&inputs); err != nil {
		utils.ValidationError(c, "Validation error", err.Error())
		return
	}

	if len(inputs) == 0 {
		utils.BadRequest(c, "No products to create", nil)
		return
	}

	// Batasi batch size
	if len(inputs) > 100 {
		utils.BadRequest(c, "Batch too large. Max 100 products per request", nil)
		return
	}

	tx := database.GetDB().Begin()
	if tx.Error != nil {
		utils.InternalServerError(c, "Failed to start transaction", tx.Error.Error())
		return
	}

	// Group by category untuk validasi lebih efisien
	categoryIDs := make(map[uint]bool)
	for _, input := range inputs {
		categoryIDs[input.CategoryID] = true
	}

	// Validasi semua kategori sekaligus
	var existingCategories []uint
	tx.Model(&models.Category{}).Where("id IN ?", keys(categoryIDs)).Pluck("id", &existingCategories)

	existingMap := make(map[uint]bool)
	for _, id := range existingCategories {
		existingMap[id] = true
	}

	// Cek semua kategori valid
	for _, input := range inputs {
		if !existingMap[input.CategoryID] {
			tx.Rollback()
			utils.BadRequest(c, "Invalid category_id", gin.H{"category_id": input.CategoryID})
			return
		}
	}

	// Insert semua produk
	products := make([]models.Product, len(inputs))
	for i, input := range inputs {
		products[i] = models.Product{
			Name:       input.Name,
			Price:      input.Price,
			Stock:      input.Stock,
			CategoryID: input.CategoryID,
		}
	}

	if err := tx.CreateInBatches(&products, 50).Error; err != nil {
		tx.Rollback()
		utils.InternalServerError(c, "Failed to create products", err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.InternalServerError(c, "Failed to commit transaction", err.Error())
		return
	}

	// Response IDs saja untuk hemat bandwidth
	var ids []uint
	for _, p := range products {
		ids = append(ids, p.ID)
	}

	utils.Created(c, "Products created successfully", gin.H{
		"count": len(products),
		"ids":   ids,
	})
}

// Helper function untuk keys map
func keys(m map[uint]bool) []uint {
	keys := make([]uint, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
		CategoryID *uint   `json:"category_id" binding:"omitempty,gt=0"`
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
		// Validasi kategori dengan cache
		if !validateCategory(*input.CategoryID) {
			utils.BadRequest(c, "Invalid category_id", nil)
			return
		}
		updates["category_id"] = input.CategoryID
	}

	// Update product
	if err := database.GetDB().Model(&product).Updates(updates).Error; err != nil {
		utils.InternalServerError(c, "Failed to update product", err.Error())
		return
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
