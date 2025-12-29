package handlers

import (
	"AI_recruit/repository"
	"net/http"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt" // Импортируем bcrypt
)

type AuthHandler struct {
	Repo *repository.AuthRepository
}

// Хеширование пароля
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (h *AuthHandler) RecruiterSignup(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		CompanyName string `json:"company_name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 1. Хешируем пароль
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 2. Отправляем в БД уже ХЕШ
	id, err := h.Repo.CreateRecruiter(c, input.Email, hashedPassword, input.CompanyName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": "jwt_token_here", "recruiter_id": id})
}

func (h *AuthHandler) CandidateSignup(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TG       string `json:"telegram_username"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 1. Хешируем пароль
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 2. Отправляем в БД ХЕШ
	id, err := h.Repo.CreateCandidate(c, input.Email, hashedPassword, input.TG)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": "jwt_token_here", "candidate_id": id})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// В методе Login мы передаем ЧИСТЫЙ пароль, 
	// так как проверка хеша будет внутри репозитория
	user, err := h.Repo.Authenticate(c, input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, user)
}