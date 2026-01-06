package repository

import (
	"AI_recruit/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CandidateRepository struct {
	DB *pgxpool.Pool
}

func NewCandidateRepository(conn *pgxpool.Pool) *CandidateRepository {
	return &CandidateRepository{conn}
}

// UC-3: Получение списка заявок по ID вакансии
func (r *CandidateRepository) GetApplicationsByVacancy(c context.Context, vacancyID string) ([]models.Application, error) {
	query := `
		SELECT a.id, a.vacancy_id, a.candidate_id, c.telegram_username, a.status, a.ai_score, a.applied_at 
        FROM applications a
        JOIN candidates c ON a.candidate_id = c.id
        WHERE a.vacancy_id = $1 
        ORDER BY a.ai_score DESC`

	rows, err := r.DB.Query(c, query, vacancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var a models.Application
		// Сканируем также telegram_username в CandidateName для отображения
		if err := rows.Scan(&a.ID, &a.VacancyID, &a.CandidateID, &a.CandidateName, &a.Status, &a.AIScore, &a.AppliedAt); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	// Возвращаем пустой слайс вместо nil, если записей нет
	if apps == nil {
		apps = []models.Application{}
	}
	return apps, nil
}

// Получение полной информации по одной заявке (ID заявки)
func (r *CandidateRepository) GetApplicationByID(c context.Context, appID string) (*models.Application, error) {
	query := `
		SELECT a.id, a.vacancy_id, a.candidate_id, c.telegram_username, a.status, a.ai_score, a.applied_at 
        FROM applications a
        JOIN candidates c ON a.candidate_id = c.id
        WHERE a.id = $1`

	var a models.Application
	err := r.DB.QueryRow(c, query, appID).Scan(
		&a.ID, &a.VacancyID, &a.CandidateID, &a.CandidateName, &a.Status, &a.AIScore, &a.AppliedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("application not found")
		}
		return nil, err
	}
	return &a, nil
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
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ResumeData{}, errors.New("analysis not found")
		}
		return models.ResumeData{}, err
	}

	return res, err
}

// UC-5: Работа с шаблонами
func (r *CandidateRepository) GetTemplates(ctx context.Context, recruiterID string) ([]models.MessageTemplate, error) {
	if recruiterID == "" {
		return nil, errors.New("recruiterID is empty")
	}

	query := `SELECT id, title, body_text FROM message_templates WHERE recruiter_id = $1`

	rows, err := r.DB.Query(ctx, query, recruiterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.MessageTemplate
	for rows.Next() {
		var t models.MessageTemplate
		if err := rows.Scan(&t.ID, &t.Title, &t.BodyText); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

// UC-3 & Stage 3: Смена статуса кандидата рекрутером
func (r *CandidateRepository) UpdateApplicationStatus(c context.Context, appID string, newStatus string) error {
	query := `UPDATE applications SET status = $1 WHERE id = $2`
	result, err := r.DB.Exec(c, query, newStatus, appID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("application not found")
	}
	return nil
}

func (r *CandidateRepository) GetTemplateByID(id string) (models.MessageTemplate, error) {
	var t models.MessageTemplate
	query := `SELECT id, title, body_text FROM message_templates WHERE id = $1`

	err := r.DB.QueryRow(context.Background(), query, id).Scan(
		&t.ID, &t.Title, &t.BodyText,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return t, errors.New("template not found")
		}
		return t, err
	}
	return t, nil
}