package handlers

import (
	"AI_recruit/models"
	"AI_recruit/repository"
	"AI_recruit/services"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CandidatePortalHandler struct {
	Repo *repository.CandidatePortalRepository
	AIService *services.AIService
}

func NewCandidatePortalHandler(CandPortRepo *repository.CandidatePortalRepository, AIService *services.AIService) *CandidatePortalHandler {
	return &CandidatePortalHandler{CandPortRepo, AIService}
}

// @Summary Подать отклик на вакансию
// @Description Загружает PDF, получает описание вакансии из БД, анализирует ИИ и сохраняет результат
// @Tags portal
// @Accept multipart/form-data
// @Produce json
// @Param candidate_id formData string true "ID кандидата"
// @Param vacancy_id formData string true "ID вакансии"
// @Param resume formData file true "Резюме PDF"
// @Router /applications [post]
func (h *CandidatePortalHandler)  Apply(c *gin.Context) {
	candidateID := c.PostForm("candidate_id")
	vacancyID := c.PostForm("vacancy_id")

	// 1. Проверяем наличие файла
	file, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не прикреплен"})
		return
	}

	// 2. РЕАЛЬНОЕ ПОЛУЧЕНИЕ ВАКАНСИИ
	// Мы должны знать требования, чтобы ИИ мог сравнить их с резюме
	vacancy, err := h.Repo.GetVacancyByID(c.Request.Context(), vacancyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Вакансия не найдена"})
		return
	}
	// Используем реальные требования из БД (например, поле Description или Requirements)
	vacDesc := vacancy.AIFilters

	// 3. Временное сохранение для парсинга
	tempPath := fmt.Sprintf("./temp_%s.pdf", uuid.New().String())
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла на сервере"})
		return
	}
	defer os.Remove(tempPath)

	// 4. Реальный анализ ИИ
	// Сервис извлекает текст из PDF и отправляет его в OpenAI вместе с vacDesc
	aiData, err := h.AIService.AnalyzeResume(tempPath, vacDesc)
	if err != nil {
		// Если ИИ не смог прочитать текст, возвращаем ошибку соискателю (как договорились ранее)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": fmt.Sprintf("Анализ не удался: %v", err)})
		return
	}

	// 5. Сохранение в БД (Атомарная транзакция)
	appID := uuid.New().String()
	app := models.Application{
		ID:          appID,
		VacancyID:   vacancyID,
		CandidateID: candidateID,
		Status:      "New",
	}
	
	// Привязываем вердикт к ID заявки
	aiData.ApplicationID = appID

	createdID, err := h.Repo.CreateApplication(c.Request.Context(), app, *aiData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения в БД"})
		return
	}

	// 6. Успешный ответ
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"app_id": createdID,
		"ai": gin.H{
			"score":   aiData.AIScore,
			"verdict": aiData.AIVerdict,
			"skills":  aiData.SkillsDetected,
		},
	})
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
