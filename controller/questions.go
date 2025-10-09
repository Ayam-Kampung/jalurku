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
// QUESTION HANDLERS
// ============================================

// GetQuestions get all questions (optionally filtered by category)
func GetQuestions(c *fiber.Ctx) error {
	categoryID := c.Query("category_id")
	db := database.DB
	var questions []model.Question

	query := db.Preload("Category").Preload("Options").Where("is_active = ?", true)

	if categoryID != "" {
		catID, err := uuid.Parse(categoryID)
		if err == nil {
			query = query.Where("category_id = ?", catID)
		}
	}

	if err := query.Order("category_id, \"order\"").Find(&questions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching questions",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Questions retrieved successfully",
		"data":    questions,
	})
}

// GetQuestion get single question by ID
func GetQuestion(c *fiber.Ctx) error {
	id := c.Params("id")
	questionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid question ID",
			"data":    nil,
		})
	}

	db := database.DB
	var question model.Question

	if err := db.Preload("Category").Preload("Options").First(&question, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Question not found",
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
		"message": "Question found",
		"data":    question,
	})
}

// CreateQuestion create new question (Admin only)
func CreateQuestion(c *fiber.Ctx) error {
	type OptionInput struct {
		Text  string `json:"text"`
		Score int    `json:"score"`
		Order int    `json:"order"`
	}

	type CreateQuestionInput struct {
		CategoryID string        `json:"category_id"`
		Text       string        `json:"text"`
		Type       string        `json:"type"`
		Order      int           `json:"order"`
		Options    []OptionInput `json:"options"`
	}

	var input CreateQuestionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"data":    err.Error(),
		})
	}

	categoryID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid category ID",
			"data":    nil,
		})
	}

	if input.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Question text is required",
			"data":    nil,
		})
	}

	db := database.DB
	question := model.Question{
		CategoryID: categoryID,
		Text:       input.Text,
		Type:       input.Type,
		Order:      input.Order,
		IsActive:   true,
	}

	// Start transaction
	tx := db.Begin()
	if err := tx.Create(&question).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't create question",
			"data":    err.Error(),
		})
	}

	// Create options if type is multiple_choice
	if input.Type == "multiple_choice" && len(input.Options) > 0 {
		for _, opt := range input.Options {
			option := model.Option{
				QuestionID: question.ID,
				Text:       opt.Text,
				Score:      opt.Score,
				Order:      opt.Order,
			}
			if err := tx.Create(&option).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Couldn't create options",
					"data":    err.Error(),
				})
			}
		}
	}

	tx.Commit()

	// Load relations
	db.Preload("Category").Preload("Options").First(&question, question.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Question created successfully",
		"data":    question,
	})
}

// UpdateQuestion update existing question (Admin only)
func UpdateQuestion(c *fiber.Ctx) error {
	type UpdateQuestionInput struct {
		Text     string `json:"text"`
		Type     string `json:"type"`
		Order    int    `json:"order"`
		IsActive *bool  `json:"is_active"`
	}

	id := c.Params("id")
	questionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid question ID",
			"data":    nil,
		})
	}

	var input UpdateQuestionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"data":    err.Error(),
		})
	}

	db := database.DB
	var question model.Question

	if err := db.First(&question, questionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Question not found",
			"data":    nil,
		})
	}

	if input.Text != "" {
		question.Text = input.Text
	}
	if input.Type != "" {
		question.Type = input.Type
	}
	if input.Order > 0 {
		question.Order = input.Order
	}
	if input.IsActive != nil {
		question.IsActive = *input.IsActive
	}

	if err := db.Save(&question).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't update question",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Question updated successfully",
		"data":    question,
	})
}

// DeleteQuestion soft delete question (Admin only)
func DeleteQuestion(c *fiber.Ctx) error {
	id := c.Params("id")
	questionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid question ID",
			"data":    nil,
		})
	}

	db := database.DB
	var question model.Question

	if err := db.First(&question, questionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Question not found",
			"data":    nil,
		})
	}

	if err := db.Delete(&question).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't delete question",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Question deleted successfully",
		"data":    nil,
	})
}