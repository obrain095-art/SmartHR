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

func (h *CandidateHandler) ListCandidates(c *gin.Context) {
	vacancyID := c.Query("vacancy_id")
	if vacancyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vacancy_id is required"})
		return
	}

	apps, err := h.Repo.GetApplicationsByVacancy(c, vacancyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching candidates: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, apps)
}

func (h *CandidateHandler) GetCandidateSummary(c *gin.Context) {
	appID := c.Query("application_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application_id is required"})
		return
	}

	summary, err := h.Repo.GetResumeAnalysis(c, appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Summary not found"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *CandidateHandler) ContactCandidate(c *gin.Context) {
	tgUsername := c.Query("tg")
	candidateName := c.Query("name")
	vacancyTitle := c.Query("vacancy")
	templateID := c.Query("template_id")

	messageBody := "Здравствуйте, %s! Мы рассмотрели ваш отклик на вакансию %s."

	if templateID != "" {
		templates, err := h.Repo.GetTemplates(c, templateID)
		if err == nil {
			for _, t := range templates {
				if t.ID == templateID {
					messageBody = t.BodyText
					break
				}
			}
		}
	}

	finalMessage := fmt.Sprintf(messageBody, candidateName, vacancyTitle)
	tgURL := fmt.Sprintf("https://t.me/%s?text=%s",
		url.PathEscape(tgUsername),
		url.QueryEscape(finalMessage),
	)

	c.JSON(http.StatusOK, gin.H{
		"tg_link": tgURL,
		"preview": finalMessage,
	})
}

func (h *CandidateHandler) UpdateStatus(c *gin.Context) {
	var input struct {
		ApplicationID string `json:"application_id"`
		Status        string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.Repo.UpdateApplicationStatus(c, input.ApplicationID, input.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

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

// POST /templates/:id/generate
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