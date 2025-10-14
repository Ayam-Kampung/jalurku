package controller

import (
	"errors"
	"net/mail"
	"time"

	"jalurku/config"
	"jalurku/database"
	"jalurku/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)
// ============================================
// HELPER FUNCTIONS
// ============================================

// Hash password dengan bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Membandingkan password dengan hash password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Pengecek validasi pengetikan format email
func isEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Validasi JWT token dengan user ID
func validToken(t *jwt.Token, id string) bool {
	userID, err := uuid.Parse(id)
	if err != nil {
		return false
	}

	claims := t.Claims.(jwt.MapClaims)
	tokenUserID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return false
	}

	return userID == tokenUserID
}

// Dapatkan pengguna dari email
func getUserByEmail(e string) (*model.User, error) {
	db := database.DB
	var user model.User
	if err := db.Where(&model.User{Email: e}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Dapatkan pengguna dari username
func getUserByUsername(u string) (*model.User, error) {
	db := database.DB
	var user model.User
	if err := db.Where(&model.User{Name: u}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Autentikasi keaslian pengguna, dicek dari kesamaan password dan hash password.
func validUser(id string, p string) bool {
	db := database.DB
	var user model.User
	userID, err := uuid.Parse(id)
	if err != nil {
		return false
	}
	
	db.First(&user, userID)
	if user.Name == "" {
		return false
	}
	if !CheckPasswordHash(p, user.Password) {
		return false
	}
	return true
}

// ============================================
// AUTH HANDLERS
// ============================================

// Log Masuk dan autentikasi pengguna, dan berikan token JWT
func Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}
	
	type UserData struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Email    string    `json:"email"`
		Role     string    `json:"role"`
	}

	// Memastikan format data adalah application/json
	if string(c.Request().Header.ContentType()) != "application/json" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_type_failed",
			"data":    nil,
		})
	}

	input := new(LoginInput)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_type_json_failed",
			"data":    err.Error(),
		})
	}

	// Memvalidasi kecukupan data untuk autentikasi
	if input.Identity == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_input_insufficient",
			"data":    nil,
		})
	}

	identity := input.Identity
	pass := input.Password
	var userModel *model.User
	var err error

	if isEmail(identity) {
		userModel, err = getUserByEmail(identity)
	} else {
		userModel, err = getUserByUsername(identity)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "internal_server_error",
			"data":    err.Error(),
		})
	}

	if userModel == nil {
		// Menghindari penyerangan timing
		CheckPasswordHash(pass, "")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_invalid",
			"data":    nil,
		})
	}

	if !CheckPasswordHash(pass, userModel.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_invalid",
			"data":    nil,
		})
	}

	// Buat token JWT
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = userModel.Name
	claims["user_id"] = userModel.ID.String()
	claims["role"] = userModel.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(config.Config("SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_jwt_token_failed",
			"data":    nil,
		})
	}

	userData := UserData{
		ID:       userModel.ID,
		Username: userModel.Name,
		Email:    userModel.Email,
		Role:     userModel.Role,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "login",
		"data": fiber.Map{
			"token": t,
			"user":  userData,
		},
	})
}

// Mendaftarkan pengguna baru
func Register(c *fiber.Ctx) error {
	type RegisterInput struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	type NewUser struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Email    string    `json:"email"`
		Role     string    `json:"role"`
	}

	// Memastikan format data adalah application/json
	if string(c.Request().Header.ContentType()) != "application/json" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_type_invalid",
			"data":    nil,
		})
	}

	// Mengambil isi data body
	input := new(RegisterInput)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_type_json_invalid",
			"data":    err.Error(),
		})
	}

	// Memvalidasi kecukupan data untuk autentikasi
	if input.Name == "" || input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_input_insufficient",
			"data":    nil,
		})
	}

	if len(input.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_input_password_insufficient",
			"data":    nil,
		})
	}

	if !isEmail(input.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_input_email_insufficient",
			"data":    nil,
		})
	}

	// Apakah email sudah ada di database?
	db := database.DB
	var existingUser model.User
	if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_conflict",
			"data":    nil,
		})
	}

	// Hash password
	hash, err := hashPassword(input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "internal_password_failed",
			"data":    err.Error(),
		})
	}

	// Buat pengguna
	user := model.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hash,
		Role:     "user",
	}

	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "user_created_failed",
			"data":    err.Error(),
		})
	}

	newUser := NewUser{
		ID:       user.ID,
		Username: user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "user_created",
		"data":    newUser,
	})
}

// ============================================
// USER HANDLERS
// ============================================

// Dapatkan pengguna sekarang yang ter-autentikasi (Log Masuk)
func GetCurrentUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_token_failed",
			"data":    nil,
		})
	}

	var user model.User
	db := database.DB

	if err := db.
		Preload("HasilAngket", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "identity_notfound",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "identity",
		"data":    user,
	})
}


// Memperbarui pengguna
func UpdateUser(c *fiber.Ctx) error {
	type UpdateUserInput struct {
		Name string `json:"name"`
	}

	id := c.Params("id")
	token := c.Locals("user").(*jwt.Token)

	if !validToken(token, id) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "internal_unauthorized",
			"data":    nil,
		})
	}

	var input UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "upduser_input_invalid",
			"data":    err.Error(),
		})
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "upduser_id_invalid",
			"data":    nil,
		})
	}

	db := database.DB
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "upduser_notfound",
			"data":    nil,
		})
	}

	if input.Name != "" {
		user.Name = input.Name
	}

	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "upduser_failed",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "upduser",
		"data":    user,
	})
}

// Menghapus pengguna
func DeleteUser(c *fiber.Ctx) error {
	type PasswordInput struct {
		Password string `json:"password"`
	}

	id := c.Params("id")
	token := c.Locals("user").(*jwt.Token)

	if !validToken(token, id) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "internal_unauthorized",
			"data":    nil,
		})
	}

	var input PasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "deluser_input_invalid",
			"data":    err.Error(),
		})
	}

	if !validUser(id, input.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "deluser_input_password_invalid",
			"data":    nil,
		})
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "deluser_id_invalid",
			"data":    nil,
		})
	}

	db := database.DB
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "deluser_notfound",
			"data":    nil,
		})
	}

	if err := db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "deluser_failed",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "deluser",
		"data":    nil,
	})
}

func IsAdmin(c *fiber.Ctx) error {
	// Ambil token JWT dari context (middleware Protected)
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	role, ok := claims["role"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "internal_unauthorized",
		})
	}

	isAdmin := role == "admin"

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"is_admin": isAdmin,
	})
}