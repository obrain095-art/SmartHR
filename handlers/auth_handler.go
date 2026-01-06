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

func NewAuthHandler(AuthRepo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{AuthRepo}
}

// Хеширование пароля
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// RecruiterSignup godoc
// @Summary      Регистрация рекрутера
// @Description  Создает новый аккаунт рекрутера с безопасным хешированием пароля через bcrypt.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input    body      object{email=string,password=string,company_name=string}  true  "Данные для регистрации рекрутера"
// @Success      200      {object}  map[string]string "Пример: { 'token': '...', 'recruiter_id': '...' }"
// @Failure      400      {object}  map[string]string "Ошибка: Неверный формат входных данных"
// @Failure      409      {object}  map[string]string "Ошибка: Пользователь с таким Email уже существует"
// @Failure      500      {object}  map[string]string "Ошибка: Критическая ошибка при хешировании или сохранении"
// @Router       /auth/recruiter/signup [post]
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

// CandidateSignup godoc
// @Summary      Регистрация соискателя
// @Description  Создает аккаунт кандидата, сохраняя email и никнейм в Telegram для последующей связи.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input    body      object{email=string,password=string,telegram_username=string}  true  "Данные для регистрации соискателя"
// @Success      200      {object}  map[string]string "Пример: { 'token': '...', 'candidate_id': '...' }"
// @Failure      400      {object}  map[string]string "Ошибка: Некорректные данные"
// @Failure      409      {object}  map[string]string "Ошибка: Почта уже используется"
// @Router       /auth/candidate/signup [post]
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




// RecruiterLogin godoc
// @Summary      Вход для рекрутера
// @Description  Аутентификация пользователя как рекрутера. Проверяет email и пароль в таблице recruiters.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input    body      object{email=string,password=string}  true  "Учетные данные рекрутера"
// @Success      200      {object}  map[string]string "Данные: { 'id': '...', 'email': '...', 'role': 'recruiter', 'company_name': '...' }"
// @Failure      400      {object}  map[string]string "Ошибка: Неверный формат входных данных"
// @Failure      401      {object}  map[string]string "Ошибка: Неверный email или пароль"
// @Router       /auth/recruiter/login [post]
func (h *AuthHandler) RecruiterLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := h.Repo.AuthenticateRecruiter(c, input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// CandidateLogin godoc
// @Summary      Вход для соискателя
// @Description  Аутентификация пользователя как кандидата. Проверяет email и пароль в таблице candidates.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input    body      object{email=string,password=string}  true  "Учетные данные соискателя"
// @Success      200      {object}  map[string]string "Данные: { 'id': '...', 'email': '...', 'role': 'candidate', 'telegram_username': '...' }"
// @Failure      400      {object}  map[string]string "Ошибка: Неверный формат входных данных"
// @Failure      401      {object}  map[string]string "Ошибка: Неверный email или пароль"
// @Router       /auth/candidate/login [post]
func (h *AuthHandler) CandidateLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := h.Repo.AuthenticateCandidate(c, input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}