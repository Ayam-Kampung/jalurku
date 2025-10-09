package controller

import (
	"errors"

	"jalurku/database"
	"jalurku/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// CATEGORY HANDLERS
// ============================================

// GetCategories get all active categories
func GetCategories(c *fiber.Ctx) error {
	db := database.DB
	var categories []model.Category

	if err := db.Preload("Questions", "is_active = ?", true).Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching categories",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Categories retrieved successfully",
		"data":    categories,
	})
}

// GetCategory get single category by ID
func GetCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid category ID",
			"data":    nil,
		})
	}

	db := database.DB
	var category model.Category

	if err := db.Preload("Questions.Options").First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Category not found",
				"data":    nil,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database error",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Category found",
		"data":    category,
	})
}

// CreateCategory create new category (Admin only)
func CreateCategory(c *fiber.Ctx) error {
	type CreateCategoryInput struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Weight      float64 `json:"weight"`
	}

	var input CreateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"data":    err.Error(),
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Category name is required",
			"data":    nil,
		})
	}

	db := database.DB
	category := model.Category{
		Name:        input.Name,
		Description: input.Description,
		Weight:      input.Weight,
	}

	if err := db.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't create category",
			"data":    err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Category created successfully",
		"data":    category,
	})
}

// UpdateCategory update existing category (Admin only)
func UpdateCategory(c *fiber.Ctx) error {
	type UpdateCategoryInput struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Weight      float64 `json:"weight"`
	}

	id := c.Params("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid category ID",
			"data":    nil,
		})
	}

	var input UpdateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"data":    err.Error(),
		})
	}

	db := database.DB
	var category model.Category

	if err := db.First(&category, categoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Category not found",
			"data":    nil,
		})
	}

	if input.Name != "" {
		category.Name = input.Name
	}
	if input.Description != "" {
		category.Description = input.Description
	}
	if input.Weight > 0 {
		category.Weight = input.Weight
	}

	if err := db.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't update category",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Category updated successfully",
		"data":    category,
	})
}

// DeleteCategory soft delete category (Admin only)
func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid category ID",
			"data":    nil,
		})
	}

	db := database.DB
	var category model.Category

	if err := db.First(&category, categoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Category not found",
			"data":    nil,
		})
	}

	if err := db.Delete(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't delete category",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Category deleted successfully",
		"data":    nil,
	})
}