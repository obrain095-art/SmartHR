package handlers

import (
	"AI_recruit/repository"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type CandidateHandler struct {
	Repo *repository.CandidateRepository
}

func NewCandidateHandler(CandRepo *repository.CandidateRepository) *CandidateHandler {
	return &CandidateHandler{CandRepo}
}

// GetListByVacancy godoc
// @Summary Список заявок на вакансию
// @Description Получает список всех откликов, привязанных к конкретному ID вакансии. (Маршрут: /:id/vacancy, где id - это ID вакансии)
// @Tags applications
// @Produce json
// @Param id path string true "ID вакансии"
// @Success 200 {array} models.Application "Список заявок"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /applications/{id} [get]
func (h *CandidateHandler) GetListByVacancy(c *gin.Context) {
	vacancyID := c.Param("id")

	apps, err := h.Repo.GetApplicationsByVacancy(c, vacancyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching applications"})
		return
	}
	c.JSON(http.StatusOK, apps)
}

// GetApplicationDetails godoc
// @Summary Полная информация о заявке
// @Description Возвращает подробную информацию о конкретном отклике по его ID. (Маршрут: /:id, где id - это ID заявки)
// @Tags applications
// @Produce json
// @Param id path string true "ID заявки (application_id)"
// @Success 200 {object} models.Application "Объект заявки"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Router /applications/{id} [get]
func (h *CandidateHandler) GetApplicationDetails(c *gin.Context) {
	appID := c.Param("id")

	app, err := h.Repo.GetApplicationByID(c, appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}
	c.JSON(http.StatusOK, app)
}

// UpdateStatus godoc
// @Summary      Изменение статуса заявки
// @Description  Обновляет статус конкретного отклика (например: 'Интервью', 'Оффер', 'Отказ') по его ID.
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "ID заявки (application_id)"
// @Param        input   body      object{status=string}  true  "Новый статус"
// @Success      200     {string}  string  "Status updated"
// @Failure      400     {object}  map[string]string "Ошибка валидации"
// @Failure      500     {object}  map[string]string "Ошибка БД"
// @Router       /applications/{id}/status [patch]
func (h *CandidateHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.Repo.UpdateApplicationStatus(c, id, input.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// GetAIData godoc
// @Summary Результаты анализа ИИ
// @Description Получает вердикт, скоринг и распознанный текст резюме
// @Tags applications
// @Param id path string true "ID заявки (application_id)"
// @Produce json
// @Success 200 {object} map[string]interface{} "Данные анализа"
// @Failure 404 {object} map[string]string "Данные не найдены"
// @Router /applications/{id}/ai-data [get]
func (h *CandidateHandler) GetAIData(c *gin.Context) {
	appID := c.Param("id")

	aiData, err := h.Repo.GetResumeAnalysis(c, appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI data not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ai_verdict":  aiData.AIVerdict,
		"ai_score":    aiData.AIScore,
		"parsed_text": aiData.ParsedText,
	})
}

// GenerateTelegramText godoc
// @Summary      Генерация сообщения для Telegram
// @Description  Генерирует текст сообщения и ссылку на чат по шаблону
// @Tags         recruiter_tools
// @Accept       json
// @Produce      json
// @Param        id       path      string  true  "ID шаблона"
// @Param        input    body      object{candidate_name=string,vacancy_title=string,telegram_username=string}  true  "Данные кандидата"
// @Success      200      {object}  map[string]string "Текст и ссылка"
// @Failure      400      {object}  map[string]string "Ошибка входных данных"
// @Failure      404      {object}  map[string]string "Шаблон не найден"
// @Router       /templates/{id}/generate [post]
func (h *CandidateHandler) GenerateTelegramText(c *gin.Context) {
	templateID := c.Param("id")

	var input struct {
		CandidateName string `json:"candidate_name"`
		VacancyTitle  string `json:"vacancy_title"`
		TGUsername    string `json:"telegram_username"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные входные данные"})
		return
	}

	template, err := h.Repo.GetTemplateByID(templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Шаблон не найден"})
		return
	}

	msg := template.BodyText
	msg = strings.ReplaceAll(msg, "{ИМЯ}", input.CandidateName)
	msg = strings.ReplaceAll(msg, "{ВАКАНСИЯ}", input.VacancyTitle)

	link := fmt.Sprintf("https://t.me/%s?text=%s",
		input.TGUsername,
		url.QueryEscape(msg),
	)

	c.JSON(http.StatusOK, gin.H{
		"generated_text": msg,
		"telegram_link":  link,
	})
}

// // ListCandidates godoc
// // @Summary Список кандидатов для вакансии
// // @Description Получает список всех откликов, привязанных к конкретному ID вакансии
// // @Tags recruiter_tools
// // @Produce json
// // @Param vacancy_id query string true "ID вакансии"
// // @Success 200 {array} models.Application "Массив заявок"
// // @Failure 400 {object} map[string]string "Отсутствует vacancy_id"
// // @Failure 500 {object} map[string]string "Ошибка при получении списка из БД"
// // @Router /candidates [get]
// func (h *CandidateHandler) ListCandidates(c *gin.Context) {
// 	vacancyID := c.Query("vacancy_id")
// 	if vacancyID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "vacancy_id is required"})
// 		return
// 	}

// 	apps, err := h.Repo.GetApplicationsByVacancy(c, vacancyID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching candidates: " + err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, apps)
// }

// // GetCandidateSummary godoc
// // @Summary Краткий отчет по кандидату
// // @Description Возвращает результаты AI-анализа резюме для конкретного отклика
// // @Tags recruiter_tools
// // @Produce json
// // @Param application_id query string true "ID отклика"
// // @Success 200 {object} models.ResumeData "Объект с результатами анализа"
// // @Failure 400 {object} map[string]string "Отсутствует application_id"
// // @Failure 404 {object} map[string]string "Анализ для данной заявки не найден"
// // @Router /candidates/summary [get]
// func (h *CandidateHandler) GetCandidateSummary(c *gin.Context) {
// 	appID := c.Query("application_id")
// 	if appID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "application_id is required"})
// 		return
// 	}

// 	summary, err := h.Repo.GetResumeAnalysis(c, appID)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Summary not found"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, summary)
// }

// // ContactCandidate godoc
// // @Summary Подготовка ссылки для связи в Telegram
// // @Description Генерирует прямую ссылку t.me с текстом сообщения. Если указан template_id, текст берется из шаблона.
// // @Tags recruiter_tools
// // @Produce json
// // @Param tg query string true "Telegram username кандидата"
// // @Param name query string true "Имя кандидата для подстановки"
// // @Param vacancy query string true "Название вакансии для подстановки"
// // @Param template_id query string false "ID шаблона сообщения (необязательно)"
// // @Success 200 {object} map[string]string "tg_link (готовая ссылка) и preview (текст сообщения)"
// // @Router /candidates/contact [get]
// func (h *CandidateHandler) ContactCandidate(c *gin.Context) {
// 	tgUsername := c.Query("tg")
// 	candidateName := c.Query("name")
// 	vacancyTitle := c.Query("vacancy")
// 	templateID := c.Query("template_id")

// 	messageBody := "Здравствуйте, %s! Мы рассмотрели ваш отклик на вакансию %s."

// 	if templateID != "" {
// 		templates, err := h.Repo.GetTemplates(c, templateID)
// 		if err == nil {
// 			for _, t := range templates {
// 				if t.ID == templateID {
// 					messageBody = t.BodyText
// 					break
// 				}
// 			}
// 		}
// 	}

// 	finalMessage := fmt.Sprintf(messageBody, candidateName, vacancyTitle)
// 	tgURL := fmt.Sprintf("https://t.me/%s?text=%s",
// 		url.PathEscape(tgUsername),
// 		url.QueryEscape(finalMessage),
// 	)

// 	c.JSON(http.StatusOK, gin.H{
// 		"tg_link": tgURL,
// 		"preview": finalMessage,
// 	})
// }