package handlers

import (
	"Kasir-API/database"
	"Kasir-API/models"
	"Kasir-API/utils"
	"github.com/gin-gonic/gin"
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

	var category models.Category
	if err := database.GetDB().First(&category, input.CategoryID).Error; err != nil {
		utils.BadRequest(c, "Invalid category_id", nil)
		return
	}

	product := models.Product{
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
		CategoryID: input.CategoryID,
	}

	if err := database.GetDB().Create(&product).Error; err != nil {
		utils.InternalServerError(c, "Failed to create product", err.Error())
		return
	}

	if err := database.GetDB().Preload("Category").First(&product, product.ID).Error; err != nil {
		utils.InternalServerError(c, "Failed to fetch created product", err.Error())
		return
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
		CategoryID *uint   `json:"category_id" binding:"omitempty,gt=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, "Validation error", err.Error())
		return
	}

	if input.Name != "" {
		product.Name = input.Name
	}

	if input.Price != 0 {
		product.Price = input.Price
	}

	if input.Stock != nil {
		product.Stock = *input.Stock
	}

	if input.CategoryID != nil {
		var category models.Category
		if err := database.GetDB().First(&category, *input.CategoryID).Error; err != nil {
			utils.BadRequest(c, "Invalid category_id", nil)
			return
		}
		product.CategoryID = *input.CategoryID
	}

	if err := database.GetDB().Save(&product).Error; err != nil {
		utils.InternalServerError(c, "Failed to update product", err.Error())
		return
	}

	if err := database.GetDB().Preload("Category").First(&product, product.ID).Error; err != nil {
		utils.InternalServerError(c, "Failed to fetch updated product", err.Error())
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
