package handlers

import (
	"AI_recruit/repository"
	"AI_recruit/models"
	"net/http"
	"github.com/gin-gonic/gin"
)

type CandidatePortalHandler struct {
	Repo *repository.CandidatePortalRepository
}

func (h *CandidatePortalHandler) Auth(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Telegram string `json:"telegram_username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	candidate := models.Candidate{
		Email:            input.Email,
		TelegramUsername: input.Telegram,
		PasswordHash:     input.Password,
	}

	id, err := h.Repo.UpsertCandidate(c, candidate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"candidate_id": id, "status": "authorized"})
}

func (h *CandidatePortalHandler) Apply(c *gin.Context) {
	candidateID := c.PostForm("candidate_id")
	vacancyID := c.PostForm("vacancy_id")

	// 1. Создаем основной отклик
	appID, err := h.Repo.CreateApplication(c, models.Application{
		VacancyID:   vacancyID,
		CandidateID: candidateID,
		AIScore:     85, // Пример оценки от ИИ
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		return
	}

	// 2. ЗАПОЛНЯЕМ ТАБЛИЦУ RESUME_DATA
	resumeInfo := models.ResumeData{
		ApplicationID:  appID,
		AIVerdict:      "Кандидат идеально подходит по стеку технологий.",
		ParsedText:     "Опыт работы 5 лет, знание Go, PostgreSQL, Docker.",
		SkillsDetected: "Go, SQL, Docker, Gin",
	}

	if err := h.Repo.SaveResumeData(c, resumeInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save AI analysis"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "app_id": appID})
}

func (h *CandidatePortalHandler) MyApplications(c *gin.Context) {
	candidateID := c.Query("candidate_id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate_id is required"})
		return
	}

	apps, err := h.Repo.GetCandidateApplications(c, candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching applications: " + err.Error()})
		return
	}

	if apps == nil {
		apps = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, apps)
}

// GET /vacancies/:short_link
func (h *CandidatePortalHandler) ViewVacancy(c *gin.Context) {
    shortLink := c.Param("short_link")
    
    // В базе ищем по полю short_link
    vacancy, err := h.Repo.GetByShortLink(shortLink)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Вакансия не найдена или перемещена в архив"})
        return
    }

    c.JSON(http.StatusOK, vacancy)
}