package handlers

import (
	"AI_recruit/repository"
	"AI_recruit/models"
	"net/http"
	"github.com/gin-gonic/gin"
)

type RecruiterHandler struct {
	Repo *repository.VacancyRepository
}

func (h *RecruiterHandler) CreateVacancy(c *gin.Context) {
	var v models.Vacancy
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

func (h *RecruiterHandler) ArchiveVacancy(c *gin.Context) {
	id := c.Param("id") // Или c.Query("id") в зависимости от роутинга
	if err := h.Repo.Archive(c, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.String(http.StatusOK, "Vacancy archived")
}

func (h *RecruiterHandler) GetApplications(c *gin.Context) {
	vacancyID := c.Param("id")

	apps, _ := h.Repo.GetApplications(c, vacancyID)
	c.JSON(http.StatusOK, apps)
}

func (h *RecruiterHandler) ListActiveVacancies(c *gin.Context) {
	vacancies, err := h.Repo.GetActive(c) 
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching vacancies"})
		return
	}

	c.JSON(http.StatusOK, vacancies)
}