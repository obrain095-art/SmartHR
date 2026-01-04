package handlers

import (
	"AI_recruit/repository"
	"AI_recruit/models"
	"net/http"
	"github.com/gin-gonic/gin"
)

type RecruiterHandler struct {
	Repo *repository.RecruiterRepository
}

func NewRecruiterHandler(RecRepo *repository.RecruiterRepository) *RecruiterHandler {
	return &RecruiterHandler{RecRepo}
}

// CreateVacancy godoc
// @Summary Создание вакансии
// @Description Создает вакансию и автоматически генерирует short_link
// @Tags vacancies
// @Accept json
// @Produce json
// @Param vacancy body models.VacancyCreateRequest true "Объект вакансии"
// @Success 201 "Вакансия успешно создана"
// @Failure 400 {object} map[string]string "Ошибка парсинга JSON"
// @Router /vacancies [post]
func (h *RecruiterHandler) CreateVacancy(c *gin.Context) {
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
// @Description Переводит вакансию в статус архивной по ID
// @Tags vacancies
// @Param id path string true "ID вакансии"
// @Success 200 {string} string "Vacancy archived"
// @Failure 404 {object} map[string]string "Вакансия не найдена"
// @Router /vacancies/{id}/archive [patch]
func (h *RecruiterHandler) ArchiveVacancy(c *gin.Context) {
	id := c.Param("id") // Или c.Query("id") в зависимости от роутинга
	if err := h.Repo.Archive(c, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.String(http.StatusOK, "Vacancy archived")
}

// GetApplications godoc
// @Summary Список откликов на вакансию
// @Description Получает все заявки кандидатов для конкретной вакансии
// @Tags vacancies
// @Param id path string true "ID вакансии"
// @Success 200 {array} models.Application
// @Router /vacancies/{id}/applications [get]
func (h *RecruiterHandler) GetApplications(c *gin.Context) {
	vacancyID := c.Param("id")

	apps, _ := h.Repo.GetApplications(c, vacancyID)
	c.JSON(http.StatusOK, apps)
}

// GetApplications godoc
// @Summary Список активных вакансий
// @Description Получает все активные вакансии
// @Tags vacancies
// @Success 200 {array} []models.Vacancy
// @Router /vacancies [get]
func (h *RecruiterHandler) ListActiveVacancies(c *gin.Context) {
	vacancies, err := h.Repo.GetActive(c) 
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching vacancies"})
		return
	}

	c.JSON(http.StatusOK, vacancies)
}