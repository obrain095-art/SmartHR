package services

import (
	"AI_recruit/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/sashabaranov/go-openai"
)

type AIService struct {
	Client *openai.Client
}

func NewAIService() *AIService {
	return &AIService{}
}


func (s *AIService) AnalyzeResume(filePath string, vacancyDesc string) (*models.ResumeData, error) {
	// 1. Извлечение текста (Pure Go)
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия PDF: %v", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга текста: %v", err)
	}
	buf.ReadFrom(b)
	resumeText := buf.String()

	if len(resumeText) < 10 {
		return nil, fmt.Errorf("текстовый слой не найден или файл пуст")
	}

	// 2. Запрос к OpenAI
	prompt := fmt.Sprintf(`Сравни резюме и вакансию. Верни ответ СТРОГО в JSON.
    ВАКАНСИЯ: %s
    РЕЗЮМЕ: %s
    ФОРМАТ: {"score": int, "skills": ["string"], "summary": "string"}
	score - процентное соотношение того насколько человек подходит для должности
	skills - навыки именно полезные для вакансии
	summary - общая оценка наваыков кандидата
	`, vacancyDesc, resumeText)

	resp, err := s.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	var aiResult struct {
		Score   int      `json:"score"`
		Skills  []string `json:"skills"`
		Summary string   `json:"summary"`
	}
	json.Unmarshal([]byte(resp.Choices[0].Message.Content), &aiResult)

	return &models.ResumeData{
		AIVerdict:      aiResult.Summary,
		ParsedText:     resumeText,
		SkillsDetected: strings.Join(aiResult.Skills, ", "),
		AIScore:        aiResult.Score,
	}, nil
}