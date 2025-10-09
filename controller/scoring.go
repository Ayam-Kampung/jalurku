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

// CompleteReflection complete reflection and calculate scores
func CompleteReflection(c *fiber.Ctx) error {
	id := c.Params("id")
	reflectionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid reflection ID",
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
			"message": "Reflection already completed",
			"data":    nil,
		})
	}

	// Get all answers with question details
	var answers []model.Answer
	if err := db.Where("reflection_id = ?", reflectionID).
		Preload("Question.Category").
		Find(&answers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error fetching answers",
			"data":    err.Error(),
		})
	}

	if len(answers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "No answers found. Please answer at least one question",
			"data":    nil,
		})
	}

	// Calculate scores per category
	categoryScores := make(map[uuid.UUID]struct {
		Score    float64
		MaxScore float64
		Category model.Category
	})

	var totalScore float64
	var maxScore float64

	// Get all active questions to calculate max possible score
	var allQuestions []model.Question
	db.Where("is_active = ?", true).Find(&allQuestions)

	// Calculate max score per category
	for _, q := range allQuestions {
		cat := categoryScores[q.CategoryID]
		if q.Type == "scale" {
			cat.MaxScore += 5.0 // Assuming scale 1-5
		} else if q.Type == "multiple_choice" {
			var maxOptionScore int
			db.Model(&model.Option{}).Where("question_id = ?", q.ID).Select("MAX(score)").Scan(&maxOptionScore)
			cat.MaxScore += float64(maxOptionScore)
		}
		categoryScores[q.CategoryID] = cat
	}

	// Calculate actual scores from answers
	for _, answer := range answers {
		cat := categoryScores[answer.Question.CategoryID]
		cat.Score += answer.Score
		cat.Category = answer.Question.Category
		categoryScores[answer.Question.CategoryID] = cat
		totalScore += answer.Score
	}

	// Calculate total max score
	for _, cat := range categoryScores {
		maxScore += cat.MaxScore
	}

	// Start transaction to save all scores
	tx := db.Begin()

	// Delete existing category scores
	tx.Where("reflection_id = ?", reflectionID).Delete(&model.CategoryScore{})

	// Save category scores
	for categoryID, cat := range categoryScores {
		percentage := 0.0
		if cat.MaxScore > 0 {
			percentage = (cat.Score / cat.MaxScore) * 100
		}

		categoryScore := model.CategoryScore{
			ReflectionID: reflectionID,
			CategoryID:   categoryID,
			Score:        cat.Score,
			MaxScore:     cat.MaxScore,
			Percentage:   percentage,
		}
		if err := tx.Create(&categoryScore).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Error saving category scores",
				"data":    err.Error(),
			})
		}
	}

	// Generate recommendations based on scores
	for categoryID, cat := range categoryScores {
		percentage := 0.0
		if cat.MaxScore > 0 {
			percentage = (cat.Score / cat.MaxScore) * 100
		}

		var recommendation model.Recommendation
		if percentage < 40 {
			recommendation = model.Recommendation{
				ReflectionID: reflectionID,
				CategoryID:   categoryID,
				Title:        "Perlu Peningkatan di " + cat.Category.Name,
				Description:  "Skor Anda pada dimensi ini masih rendah. Disarankan untuk lebih fokus meningkatkan aspek-aspek dalam dimensi " + cat.Category.Name + ".",
				Priority:     "high",
			}
		} else if percentage < 70 {
			recommendation = model.Recommendation{
				ReflectionID: reflectionID,
				CategoryID:   categoryID,
				Title:        "Terus Tingkatkan " + cat.Category.Name,
				Description:  "Anda sudah cukup baik, namun masih ada ruang untuk berkembang di dimensi " + cat.Category.Name + ".",
				Priority:     "medium",
			}
		} else {
			recommendation = model.Recommendation{
				ReflectionID: reflectionID,
				CategoryID:   categoryID,
				Title:        "Pertahankan Prestasi di " + cat.Category.Name,
				Description:  "Selamat! Anda sudah sangat baik di dimensi " + cat.Category.Name + ". Terus pertahankan dan kembangkan lebih lanjut.",
				Priority:     "low",
			}
		}
		
		if err := tx.Create(&recommendation).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Error saving recommendations",
				"data":    err.Error(),
			})
		}
	}

	// Update reflection with final scores
	now := time.Now()
	reflection.TotalScore = totalScore
	reflection.MaxScore = maxScore
	if maxScore > 0 {
		reflection.Percentage = (totalScore / maxScore) * 100
	}
	reflection.Status = "completed"
	reflection.CompletedAt = &now

	if err := tx.Save(&reflection).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error updating reflection",
			"data":    err.Error(),
		})
	}

	tx.Commit()

	// Load complete reflection with all relations
	db.Where("id = ?", reflectionID).
		Preload("CategoryScores.Category").
		Preload("Recommendations.Category").
		First(&reflection)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Reflection completed successfully",
		"data":    reflection,
	})
}

// GetReflectionReport get detailed report of completed reflection
func GetReflectionReport(c *fiber.Ctx) error {
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
		Preload("User").
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

	if reflection.Status != "completed" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Reflection not yet completed",
			"data":    nil,
		})
	}

	// Get score threshold labels
	var thresholds []model.ScoreThreshold
	db.Where("category_id IS NULL").Order("min_score").Find(&thresholds)

	var overallLabel string
	for _, threshold := range thresholds {
		if reflection.Percentage >= threshold.MinScore && reflection.Percentage <= threshold.MaxScore {
			overallLabel = threshold.Label
			break
		}
	}

	report := fiber.Map{
		"reflection": reflection,
		"summary": fiber.Map{
			"total_score":      reflection.TotalScore,
			"max_score":        reflection.MaxScore,
			"percentage":       reflection.Percentage,
			"label":            overallLabel,
			"completed_at":     reflection.CompletedAt,
		},
		"category_scores":  reflection.CategoryScores,
		"recommendations":  reflection.Recommendations,
		"total_answers":    len(reflection.Answers),
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Report generated successfully",
		"data":    report,
	})
}

// DeleteReflection soft delete reflection
func DeleteReflection(c *fiber.Ctx) error {
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

	if err := db.Where("id = ? AND user_id = ?", reflectionID, userID).First(&reflection).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Reflection not found or unauthorized",
			"data":    nil,
		})
	}

	if err := db.Delete(&reflection).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't delete reflection",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Reflection deleted successfully",
		"data":    nil,
	})
}

// ============================================
// ANALYTICS & STATISTICS
// ============================================

// GetUserStatistics get user reflection statistics
func GetUserStatistics(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))

	db := database.DB

	// Count total reflections
	var totalReflections int64
	db.Model(&model.Reflection{}).Where("user_id = ?", userID).Count(&totalReflections)

	// Count completed reflections
	var completedReflections int64
	db.Model(&model.Reflection{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completedReflections)

	// Get average score
	var avgScore float64
	db.Model(&model.Reflection{}).
		Where("user_id = ? AND status = ?", userID, "completed").
		Select("AVG(percentage)").
		Scan(&avgScore)

	// Get latest reflection
	var latestReflection model.Reflection
	db.Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&latestReflection)

	// Get score trend (last 5 completed reflections)
	var scoreTrend []model.Reflection
	db.Where("user_id = ? AND status = ?", userID, "completed").
		Select("id, title, percentage, completed_at").
		Order("completed_at DESC").
		Limit(5).
		Find(&scoreTrend)

	// Get category performance (average per category from all completed reflections)
	type CategoryPerformance struct {
		CategoryID   uuid.UUID `json:"category_id"`
		CategoryName string    `json:"category_name"`
		AvgScore     float64   `json:"avg_score"`
		AvgPercentage float64  `json:"avg_percentage"`
	}

	var categoryPerformance []CategoryPerformance
	db.Table("category_scores").
		Select("category_scores.category_id, categories.name as category_name, AVG(category_scores.score) as avg_score, AVG(category_scores.percentage) as avg_percentage").
		Joins("JOIN reflections ON reflections.id = category_scores.reflection_id").
		Joins("JOIN categories ON categories.id = category_scores.category_id").
		Where("reflections.user_id = ? AND reflections.status = ?", userID, "completed").
		Group("category_scores.category_id, categories.name").
		Scan(&categoryPerformance)

	statistics := fiber.Map{
		"total_reflections":     totalReflections,
		"completed_reflections": completedReflections,
		"draft_reflections":     totalReflections - completedReflections,
		"average_score":         avgScore,
		"latest_reflection": fiber.Map{
			"id":         latestReflection.ID,
			"title":      latestReflection.Title,
			"status":     latestReflection.Status,
			"created_at": latestReflection.CreatedAt,
		},
		"score_trend":          scoreTrend,
		"category_performance": categoryPerformance,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Statistics retrieved successfully",
		"data":    statistics,
	})
}

// GetLeaderboard get leaderboard (optional: for gamification)
func GetLeaderboard(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit > 100 {
		limit = 100
	}

	db := database.DB

	type LeaderboardEntry struct {
		UserID       uuid.UUID `json:"user_id"`
		UserName     string    `json:"user_name"`
		AvgScore     float64   `json:"avg_score"`
		TotalReflections int64 `json:"total_reflections"`
	}

	var leaderboard []LeaderboardEntry
	db.Table("reflections").
		Select("users.id as user_id, users.name as user_name, AVG(reflections.percentage) as avg_score, COUNT(reflections.id) as total_reflections").
		Joins("JOIN users ON users.id = reflections.user_id").
		Where("reflections.status = ?", "completed").
		Group("users.id, users.name").
		Order("avg_score DESC").
		Limit(limit).
		Scan(&leaderboard)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Leaderboard retrieved successfully",
		"data":    leaderboard,
	})
}

// ============================================
// ADMIN ANALYTICS
// ============================================

// GetAdminDashboard get admin dashboard statistics
func GetAdminDashboard(c *fiber.Ctx) error {
	db := database.DB

	// Total users
	var totalUsers int64
	db.Model(&model.User{}).Count(&totalUsers)

	// Total reflections
	var totalReflections int64
	db.Model(&model.Reflection{}).Count(&totalReflections)

	// Completed reflections
	var completedReflections int64
	db.Model(&model.Reflection{}).Where("status = ?", "completed").Count(&completedReflections)

	// Total categories
	var totalCategories int64
	db.Model(&model.Category{}).Count(&totalCategories)

	// Total questions
	var totalQuestions int64
	db.Model(&model.Question{}).Where("is_active = ?", true).Count(&totalQuestions)

	// Recent activities (last 10 completed reflections)
	type RecentActivity struct {
		ReflectionID uuid.UUID  `json:"reflection_id"`
		UserName     string     `json:"user_name"`
		Title        string     `json:"title"`
		Percentage   float64    `json:"percentage"`
		CompletedAt  *time.Time `json:"completed_at"`
	}

	var recentActivities []RecentActivity
	db.Table("reflections").
		Select("reflections.id as reflection_id, users.name as user_name, reflections.title, reflections.percentage, reflections.completed_at").
		Joins("JOIN users ON users.id = reflections.user_id").
		Where("reflections.status = ?", "completed").
		Order("reflections.completed_at DESC").
		Limit(10).
		Scan(&recentActivities)

	// Category usage statistics
	type CategoryUsage struct {
		CategoryID   uuid.UUID `json:"category_id"`
		CategoryName string    `json:"category_name"`
		QuestionCount int64    `json:"question_count"`
		AvgScore     float64   `json:"avg_score"`
	}

	var categoryUsage []CategoryUsage
	db.Table("categories").
		Select("categories.id as category_id, categories.name as category_name, COUNT(DISTINCT questions.id) as question_count, AVG(category_scores.percentage) as avg_score").
		Joins("LEFT JOIN questions ON questions.category_id = categories.id").
		Joins("LEFT JOIN category_scores ON category_scores.category_id = categories.id").
		Group("categories.id, categories.name").
		Scan(&categoryUsage)

	dashboard := fiber.Map{
		"total_users":            totalUsers,
		"total_reflections":      totalReflections,
		"completed_reflections":  completedReflections,
		"draft_reflections":      totalReflections - completedReflections,
		"total_categories":       totalCategories,
		"total_questions":        totalQuestions,
		"completion_rate":        float64(completedReflections) / float64(totalReflections) * 100,
		"recent_activities":      recentActivities,
		"category_usage":         categoryUsage,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Dashboard data retrieved successfully",
		"data":    dashboard,
	})
}