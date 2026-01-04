package repository

import (
	"AI_recruit/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecruiterRepository struct {
	DB *pgxpool.Pool
}

func NewRecruiterRepository(conn *pgxpool.Pool) *RecruiterRepository {
	return &RecruiterRepository{conn}
}

// GET /vacancies/{id}/applications
func (r *RecruiterRepository) GetApplications(c context.Context, vacancyID string) ([]models.Application, error) {
	query := `
		SELECT 
			a.id, 
			a.vacancy_id, 
			a.candidate_id, 
			c.telegram_username, 
			a.status, 
			a.ai_score, 
			a.applied_at
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
		// Сканируем данные, включая CandidateName (telegram_username)
		err := rows.Scan(&a.ID, &a.VacancyID, &a.CandidateID, &a.CandidateName, &a.Status, &a.AIScore, &a.AppliedAt)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

// PATCH /vacancies/{id}/archive
func (r *RecruiterRepository) Archive(c context.Context, id string) error {
	query := `UPDATE vacancies SET is_archived = true WHERE id = $1`
	result, err := r.DB.Exec(c, query, id)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return errors.New("not found")
	}
	return nil
}

// Создание вакансии (теперь соответствует вызову в хендлере)
func (r *RecruiterRepository) Create(c context.Context, v models.VacancyCreateRequest) error {
	query := `INSERT INTO vacancies (recruiter_id, title, ai_filters, short_link, is_archived) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(c, query, v.RecruiterID, v.Title, v.AIFilters, v.ShortLink, false)
	return err
}

// GetActive возвращает только неархивированные вакансии
func (r *RecruiterRepository) GetActive(c context.Context) ([]models.Vacancy, error) {
	query := `SELECT id, recruiter_id, title, ai_filters, short_link, is_archived 
              FROM vacancies WHERE is_archived = false`

	rows, err := r.DB.Query(c, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vacancies []models.Vacancy
	for rows.Next() {
		var v models.Vacancy
		err := rows.Scan(&v.ID, &v.RecruiterID, &v.Title, &v.AIFilters, &v.ShortLink, &v.IsArchived)
		if err != nil {
			return nil, err
		}
		vacancies = append(vacancies, v)
	}
	return vacancies, nil
}

