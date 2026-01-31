package handlers

import (
	"Kasir-API/database"
	"Kasir-API/models"
	"Kasir-API/utils"
	"github.com/gin-gonic/gin"
	_ "strconv"
)

// GetAllCategories - GET /categories
func GetAllCategories(c *gin.Context) {
	var categories []models.Category

	// Optimized: Select only necessary fields
	if err := database.GetDB().Select("id", "name", "description", "created_at", "updated_at").Find(&categories).Error; err != nil {
		utils.InternalServerError(c, "Failed to fetch categories", err.Error())
		return
	}

	if len(categories) == 0 {
		utils.Success(c, "No categories found", []interface{}{})
		return
	}

	utils.Success(c, "Categories retrieved successfully", categories)
}

// GetCategoryByID - GET /categories/{id}
func GetCategoryByID(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.GetDB().First(&category, id).Error; err != nil {
		utils.NotFound(c, "Category")
		return
	}

	utils.Success(c, "Category retrieved successfully", category)
}

// CreateCategory - POST /categories
func CreateCategory(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required,min=3,max=100"`
		Description string `json:"description" binding:"max=500"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, "Validation error", err.Error())
		return
	}

	// Check if category name already exists
	var existingCategory models.Category
	if err := database.GetDB().Where("name = ?", input.Name).First(&existingCategory).Error; err == nil {
		utils.BadRequest(c, "Category name already exists", nil)
		return
	}

	category := models.Category{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := database.GetDB().Create(&category).Error; err != nil {
		utils.InternalServerError(c, "Failed to create category", err.Error())
		return
	}

	utils.Created(c, "Category created successfully", category)
}

// UpdateCategory - PUT /categories/{id}
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.GetDB().First(&category, id).Error; err != nil {
		utils.NotFound(c, "Category")
		return
	}

	var input struct {
		Name        string `json:"name" binding:"omitempty,min=3,max=100"`
		Description string `json:"description" binding:"omitempty,max=500"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, "Validation error", err.Error())
		return
	}

	// Check if new name already exists (if provided and different from current)
	if input.Name != "" && input.Name != category.Name {
		var existingCategory models.Category
		if err := database.GetDB().Where("name = ? AND id != ?", input.Name, id).First(&existingCategory).Error; err == nil {
			utils.BadRequest(c, "Category name already exists", nil)
			return
		}
		category.Name = input.Name
	}

	if input.Description != "" {
		category.Description = input.Description
	}

	if err := database.GetDB().Save(&category).Error; err != nil {
		utils.InternalServerError(c, "Failed to update category", err.Error())
		return
	}

	utils.Success(c, "Category updated successfully", category)
}

// DeleteCategory - DELETE /categories/{id}
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.GetDB().First(&category, id).Error; err != nil {
		utils.NotFound(c, "Category")
		return
	}

	if err := database.GetDB().Delete(&category).Error; err != nil {
		utils.InternalServerError(c, "Failed to delete category", err.Error())
		return
	}

	utils.Success(c, "Category deleted successfully", gin.H{
		"id": category.ID,
	})
}
