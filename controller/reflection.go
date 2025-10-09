package controller

import (
	"errors"
	"time"

	"jalurku/database"
	"jalurku/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// REFLECTION HANDLERS
// ============================================

// GetReflections get all reflections for current user
func GetReflections(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user ID in token",
			"data":    nil,
		})
	}

	db := database.DB
	var reflections []model.Reflection

	if err := db.Where("user_id = ?", userID).
		Preload("CategoryScores.Category").
		Order("created_at DESC").
		Find(&reflections).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching reflections",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Reflections retrieved successfully",
		"data":    reflections,
	})
}

// GetReflection get single reflection by ID
func GetReflection(c *fiber.Ctx) error {
	id := c.Params("id")
	reflectionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid reflection ID",
			"data":    nil,
		})
	}

	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))

	db := database.DB
	var reflection model.Reflection

	if err := db.Where("id = ? AND user_id = ?", reflectionID, userID).
		Preload("Answers.Question.Category").
		Preload("Answers.Option").
		Preload("CategoryScores.Category").
		Preload("Recommendations.Category").
		First(&reflection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Reflection not found",
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
		"message": "Reflection found",
		"data":    reflection,
	})
}

// CreateReflection create new reflection session
func CreateReflection(c *fiber.Ctx) error {
	type CreateReflectionInput struct {
		Title string `json:"title"`
	}

	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user ID in token",
			"data":    nil,
		})
	}

	var input CreateReflectionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"data":    err.Error(),
		})
	}

	if input.Title == "" {
		input.Title = "Refleksi " + time.Now().Format("02 Jan 2006 15:04")
	}

	db := database.DB
	reflection := model.Reflection{
		UserID: userID,
		Title:  input.Title,
		Status: "draft",
	}

	if err := db.Create(&reflection).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't create reflection",
			"data":    err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Reflection created successfully",
		"data":    reflection,
	})
}

// SubmitAnswer submit answer for a question in reflection
func SubmitAnswer(c *fiber.Ctx) error {
	type SubmitAnswerInput struct {
		ReflectionID string     `json:"reflection_id"`
		QuestionID   string     `json:"question_id"`
		OptionID     *string    `json:"option_id"`
		ScaleValue   *int       `json:"scale_value"`
		TextValue    string     `json:"text_value"`
	}

	var input SubmitAnswerInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"data":    err.Error(),
		})
	}

	reflectionID, err := uuid.Parse(input.ReflectionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid reflection ID",
			"data":    nil,
		})
	}

	questionID, err := uuid.Parse(input.QuestionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid question ID",
			"data":    nil,
		})
	}

	// Verify user owns this reflection
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))

	db := database.DB
	var reflection model.Reflection
	if err := db.Where("id = ? AND user_id = ?", reflectionID, userID).First(&reflection).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Reflection not found or unauthorized",
			"data":    nil,
		})
	}

	if reflection.Status == "completed" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot modify completed reflection",
			"data":    nil,
		})
	}

	// Get question to determine type
	var question model.Question
	if err := db.Preload("Options").First(&question, questionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Question not found",
			"data":    nil,
		})
	}

	// Calculate score based on question type
	var score float64
	var optionIDUUID *uuid.UUID

	if question.Type == "scale" && input.ScaleValue != nil {
		score = float64(*input.ScaleValue)
	} else if question.Type == "multiple_choice" && input.OptionID != nil {
		optID, err := uuid.Parse(*input.OptionID)
		if err == nil {
			optionIDUUID = &optID
			var option model.Option
			if err := db.First(&option, optID).Error; err == nil {
				score = float64(option.Score)
			}
		}
	}

	// Check if answer already exists
	var answer model.Answer
	result := db.Where("reflection_id = ? AND question_id = ?", reflectionID, questionID).First(&answer)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database error",
			"data":    result.Error.Error(),
		})
	}

	// Update or create answer
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new answer
		answer = model.Answer{
			ReflectionID: reflectionID,
			QuestionID:   questionID,
			OptionID:     optionIDUUID,
			ScaleValue:   input.ScaleValue,
			TextValue:    input.TextValue,
			Score:        score,
		}
		if err := db.Create(&answer).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Couldn't save answer",
				"data":    err.Error(),
			})
		}
	} else {
		// Update existing answer
		answer.OptionID = optionIDUUID
		answer.ScaleValue = input.ScaleValue
		answer.TextValue = input.TextValue
		answer.Score = score
		if err := db.Save(&answer).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Couldn't update answer",
				"data":    err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Answer submitted successfully",
		"data":    answer,
	})
}