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



// UpdateStatus godoc
// @Summary      Изменение статуса заявки
// @Description  Обновляет статус конкретного отклика (например: 'Интервью', 'Оффер', 'Отказ') по его ID в URL.
// @Tags         recruiter_tools
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "ID заявки (application_id)"
// @Param        input   body      object{status=string}  true  "Новый статус для установки"
// @Success      200     {string}  string  "Статус успешно обновлен"
// @Failure      400     {object}  map[string]string "Ошибка: Неверный формат JSON"
// @Failure      500     {object}  map[string]string "Ошибка: Проблема при обновлении в базе данных"
// @Router       /applications/{id}/status [patch]
func (h *CandidateHandler) UpdateStatus(c *gin.Context) {
    // 1. Берем ID из пути (согласно роуту /applications/:id/status)
    id := c.Param("id") 

    // 2. Из тела берем только новый статус
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
// @Description Получает вердикт, скоринг и распознанный текст из RESUME_DATA
// @Tags recruiter_tools
// @Param id path string true "ID заявки (application_id)"
// @Produce json
// @Success 200 {object} map[string]interface{} "Данные анализа"
// @Router /applications/{id}/ai-data [get]
func (h *CandidateHandler) GetAIData(c *gin.Context) {
	appID := c.Param("id") // Используем Param для путей типа /applications/:id/ai-data

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
// @Description  Получает шаблон сообщения по ID, заменяет в нем плейсхолдеры {name} и {job} данными кандидата и формирует прямую ссылку на чат.
// @Tags         recruiter_tools
// @Accept       json
// @Produce      json
// @Param        id       path      string  true  "ID шаблона сообщения в базе данных"
// @Param        input    body      object{candidate_name=string,vacancy_title=string,telegram_username=string}  true  "Данные для генерации сообщения"
// @Success      200      {object}  map[string]string "Пример ответа: { 'generated_text': '...', 'telegram_link': '...' }"
// @Failure      400      {object}  map[string]string "Ошибка: Неверные входные данные или пустые поля"
// @Failure      404      {object}  map[string]string "Ошибка: Шаблон с таким ID не найден"
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

	// 1. Получаем тело шаблона из БД
	template, err := h.Repo.GetTemplateByID(templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Шаблон не найден"})
		return
	}

	// 2. Делаем простую замену переменных в тексте
	// Предположим, в шаблоне текст: "Привет, {name}! Вакансия: {job}"
	msg := template.BodyText
	msg = strings.ReplaceAll(msg, "{name}", input.CandidateName)
	msg = strings.ReplaceAll(msg, "{job}", input.VacancyTitle)

	// 3. Формируем прямую ссылку для открытия чата в Telegram
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