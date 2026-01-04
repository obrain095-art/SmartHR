package repository

import (
	"AI_recruit/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CandidateRepository struct {
	DB *pgxpool.Pool
}

func NewCandidateRepository(conn *pgxpool.Pool) *CandidateRepository {
	return &CandidateRepository{conn}
}

// UC-3: Получение списка талантов с AI-скорингом
func (r *CandidateRepository) GetApplicationsByVacancy(c context.Context, vacancyID string) ([]models.Application, error) {
	query := `SELECT id, vacancy_id, candidate_id, status, ai_score, applied_at 
              FROM applications WHERE vacancy_id = $1 ORDER BY ai_score DESC`
	rows, err := r.DB.Query(c, query, vacancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var a models.Application
		rows.Scan(&a.ID, &a.VacancyID, &a.CandidateID, &a.Status, &a.AIScore, &a.AppliedAt)
		apps = append(apps, a)
	}
	return apps, nil
}

// GetResumeAnalysis собирает полный отчет ИИ по ID отклика
func (r *CandidateRepository) GetResumeAnalysis(c context.Context, appID string) (models.ResumeData, error) {
	var res models.ResumeData

	// Используем JOIN, так как скоринг хранится в applications, а текст в resume_data
	query := `
		SELECT r.id, r.application_id, r.ai_verdict, r.parsed_text, r.skills_detected, a.ai_score
		FROM resume_data r
		JOIN applications a ON r.application_id = a.id
		WHERE r.application_id = $1`

	err := r.DB.QueryRow(c, query, appID).Scan(
		&res.ID,
		&res.ApplicationID,
		&res.AIVerdict,
		&res.ParsedText,
		&res.SkillsDetected,
		&res.AIScore,
	)
	return res, err
}

// UC-5: Работа с шаблонами
func (r *CandidateRepository) GetTemplates(ctx context.Context, recruiterID string) ([]models.MessageTemplate, error) {
    // 1. Проверяем, не пустой ли ID, если это критично
    if recruiterID == "" {
        return nil, errors.New("recruiterID is empty")
    }

    query := `SELECT id, title, body_text FROM message_templates WHERE recruiter_id = $1`
    
    rows, err := r.DB.Query(ctx, query, recruiterID) // Используем Pool (pgx)
    if err != nil {
        return nil, err
    }
    defer rows.Close() // ОБЯЗАТЕЛЬНО закрываем

    var templates []models.MessageTemplate
    for rows.Next() {
        var t models.MessageTemplate
        // ОБЯЗАТЕЛЬНО проверяем ошибку Scan
        if err := rows.Scan(&t.ID, &t.Title, &t.BodyText); err != nil {
            return nil, err
        }
        templates = append(templates, t)
    }

    // Проверка на ошибки, возникшие во время итерации
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return templates, nil
}
// UC-3 & Stage 3: Смена статуса кандидата рекрутером
func (r *CandidateRepository) UpdateApplicationStatus(c context.Context, appID string, newStatus string) error {
	query := `UPDATE applications SET status = $1 WHERE id = $2`
	_, err := r.DB.Exec(c, query, newStatus, appID)
	return err
}

// GET /applications/{id}/ai-data
func (r *CandidateRepository) GetAIData(c context.Context, appID string) (map[string]interface{}, error) {
	var verdict, parsedText string
	var score int
	query := `SELECT ai_verdict, parsed_text, ai_score FROM resume_data 
              JOIN applications ON applications.id = resume_data.application_id 
              WHERE application_id = $1`
	err := r.DB.QueryRow(c, query, appID).Scan(&verdict, &parsedText, &score)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"ai_verdict":  verdict,
		"ai_score":    score,
		"parsed_text": parsedText,
	}, nil
}

func (r *CandidateRepository) GetTemplateByID(id string) (models.MessageTemplate, error) {
    var t models.MessageTemplate
    query := `SELECT id, title, body_text FROM message_templates WHERE id = $1`
    
    err := r.DB.QueryRow(context.Background(), query, id).Scan(
        &t.ID, &t.Title, &t.BodyText,
    )
    return t, err
}