package handlers

import (
	"AI_recruit/models"
	"AI_recruit/repository"
	"net/http"
	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	Repo *repository.TemplateRepository
}

// CreateTemplate godoc
// @Summary      Создать шаблон сообщения
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        input    body      object{recruiter_id=string,title=string,body_text=string}  true  "Данные шаблона"
// @Success      201      {string}  string "Created"
// @Router       /templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var t models.MessageTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.Repo.Create(c, t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

// ListTemplates godoc
// @Summary      Список всех шаблонов рекрутера
// @Tags         templates
// @Produce      json
// @Param        recruiter_id    query     string  true  "ID рекрутера"
// @Success      200      {array}   models.MessageTemplate
// @Router       /templates [get]
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	rID := c.Query("recruiter_id")
	templates, err := h.Repo.GetAllByRecruiter(c, rID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

// UpdateTemplate godoc
// @Summary      Обновить шаблон
// @Tags         templates
// @Accept       json
// @Param        id       path      string  true  "ID шаблона"
// @Param        input    body      object{title=string,body_text=string}  true  "Новые данные"
// @Success      200      {string}  string "Updated"
// @Router       /templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	var t models.MessageTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.Repo.Update(c, id, t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// DeleteTemplate godoc
// @Summary      Удалить шаблон
// @Tags         templates
// @Param        id       path      string  true  "ID шаблона"
// @Success      204      {string}  string "Deleted"
// @Router       /templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}