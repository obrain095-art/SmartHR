package handlers

import (
	"AI_recruit/models"
	"AI_recruit/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VacancyHandler struct {
	Repo *repository.VacancyRepository
}

func NewVacancyHandler(RecRepo *repository.VacancyRepository) *VacancyHandler {
	return &VacancyHandler{RecRepo}
}

// CreateVacancy godoc
// @Summary Создание вакансии
// @Description Создает новую вакансию и автоматически генерирует short_link
// @Tags vacancies
// @Accept json
// @Produce json
// @Param vacancy body models.VacancyCreateRequest true "Объект вакансии"
// @Success 201 {string} string "Vacancy created"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Router /vacancies [post]
func (h *VacancyHandler) CreateVacancy(c *gin.Context) {
	var v models.VacancyCreateRequest
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	v.ShortLink = "hire.me/" + v.Title

	if err := h.Repo.Create(c, v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

// ArchiveVacancy godoc
// @Summary Архивация вакансии
// @Description Переводит вакансию в статус "архивная" по ID
// @Tags vacancies
// @Param id path string true "ID вакансии"
// @Success 200 {string} string "Vacancy archived"
// @Failure 404 {object} map[string]string "Вакансия не найдена"
// @Router /vacancies/{id}/archive [patch]
func (h *VacancyHandler) ArchiveVacancy(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Archive(c, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.String(http.StatusOK, "Vacancy archived")
}

// DeArchiveVacancy godoc
// @Summary Разархивация вакансии
// @Description Восстанавливает вакансию из архива (делает активной)
// @Tags vacancies
// @Param id path string true "ID вакансии"
// @Success 200 {string} string "Vacancy de-archived"
// @Failure 404 {object} map[string]string "Вакансия не найдена"
// @Router /vacancies/{id}/dearchive [patch]
func (h *VacancyHandler) DeArchiveVacancy(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.DeArchive(c, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.String(http.StatusOK, "Vacancy de-archived")
}

// GetApplications godoc
// @Summary Список откликов на вакансию
// @Description Получает все заявки кандидатов для конкретной вакансии
// @Tags vacancies
// @Param id path string true "ID вакансии"
// @Success 200 {array} models.Application
// @Router /vacancies/{id}/applications [get]
func (h *VacancyHandler) GetApplications(c *gin.Context) {
	vacancyID := c.Param("id")

	apps, err := h.Repo.GetApplications(c, vacancyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching applications"})
		return
	}
	// Если nil, возвращаем пустой массив
	if apps == nil {
		apps = []models.Application{}
	}
	c.JSON(http.StatusOK, apps)
}

// ListActiveVacancies godoc
// @Summary Список активных вакансий
// @Description Получает все незаархивированные вакансии
// @Tags vacancies
// @Success 200 {array} models.Vacancy
// @Router /vacancies/active [get]
func (h *VacancyHandler) ListActiveVacancies(c *gin.Context) {
	vacancies, err := h.Repo.GetActive(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching vacancies"})
		return
	}
	c.JSON(http.StatusOK, vacancies)
}

// ListAllVacancies godoc
// @Summary Список всех вакансий
// @Description Получает список всех вакансий рекрутера (и активных, и архивных)
// @Param id path string true "ID рекрутера у которого вытаскиваются все вакансии"
// @Tags vacancies
// @Success 200 {array} models.Vacancy
// @Router /vacancies/all [get]
func (h *VacancyHandler) ListAllVacancies(c *gin.Context) {
	recruiter_id := c.Param("recruiter_id")

	vacancies, err := h.Repo.GetAll(c, recruiter_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching vacancies"})
		return
	}
	c.JSON(http.StatusOK, vacancies)
}

// ListInactiveVacancies godoc
// @Summary Список архивных вакансий
// @Description Получает список только неактивных (заархивированных) вакансий
// @Tags vacancies
// @Success 200 {array} models.Vacancy
// @Router /vacancies/inactive [get]
func (h *VacancyHandler) ListInactiveVacancies(c *gin.Context) {
	vacancies, err := h.Repo.GetInactive(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching vacancies"})
		return
	}
	c.JSON(http.StatusOK, vacancies)
}

// ListOneVacancy godoc
// @Summary Получить информацию о вакансии
// @Description Возвращает полную информацию о выбранной вакансии по ID
// @Tags vacancies
// @Param id path string true "ID вакансии"
// @Success 200 {object} models.Vacancy
// @Failure 404 {object} map[string]string "Вакансия не найдена"
// @Router /vacancies/{id} [get]
func (h *VacancyHandler) ListOneVacancy(c *gin.Context) {
	id := c.Param("id")
	vacancy, err := h.Repo.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vacancy not found"})
		return
	}
	c.JSON(http.StatusOK, vacancy)
}