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


// Apply godoc
// @Summary Подача отклика
// @Description Создает отклик и запись в RESUME_DATA (ИИ-анализ)
// @Tags portal
// @Accept x-www-form-urlencoded
// @Produce json
// @Param candidate_id formData string true "ID кандидата"
// @Param vacancy_id formData string true "ID вакансии"
// @Success 201 {object} map[string]interface{} "Успешный отклик и ID заявки"
// @Router /applications [post]
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


// MyApplications godoc
// @Summary Мои отклики
// @Description Получение истории всех откликов текущего кандидата
// @Tags portal
// @Param candidate_id query string true "ID кандидата"
// @Success 200 {array} map[string]interface{}
// @Router /my-applications [get]
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

// ViewVacancy godoc
// @Summary Просмотр вакансии по ссылке
// @Description Получение деталей вакансии через short_link
// @Tags portal
// @Param short_link path string true "Короткая ссылка вакансии"
// @Success 200 {object} models.Vacancy
// @Failure 404 {object} map[string]string "Вакансия не найдена"
// @Router /vacancies/link/{short_link} [get]
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