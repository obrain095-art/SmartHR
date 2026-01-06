package repository

import (
	"AI_recruit/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VacancyRepository struct {
	DB *pgxpool.Pool
}

func NewVacancyRepository(conn *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{conn}
}

// GetApplications получает список откликов
func (r *VacancyRepository) GetApplications(c context.Context, vacancyID string) ([]models.Application, error) {
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
		// Сканируем данные
		err := rows.Scan(&a.ID, &a.VacancyID, &a.CandidateID, &a.CandidateName, &a.Status, &a.AIScore, &a.AppliedAt)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

// Archive архивирует вакансию
func (r *VacancyRepository) Archive(c context.Context, id string) error {
	query := `UPDATE vacancies SET is_archived = true WHERE id = $1`
	result, err := r.DB.Exec(c, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

// DeArchive разархивирует вакансию
func (r *VacancyRepository) DeArchive(c context.Context, id string) error {
	query := `UPDATE vacancies SET is_archived = false WHERE id = $1`
	result, err := r.DB.Exec(c, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

// Create создает вакансию
func (r *VacancyRepository) Create(c context.Context, v models.VacancyCreateRequest) error {
	query := `INSERT INTO vacancies (recruiter_id, title, ai_filters, short_link, is_archived) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(c, query, v.RecruiterID, v.Title, v.AIFilters, v.ShortLink, false)
	return err
}

// GetActive возвращает только активные вакансии
func (r *VacancyRepository) GetActive(c context.Context) ([]models.Vacancy, error) {
	query := `SELECT id, recruiter_id, title, ai_filters, short_link, is_archived 
              FROM vacancies WHERE is_archived = false`
	return r.fetchVacancies(c, query)
}

// GetInactive возвращает только заархивированные вакансии
func (r *VacancyRepository) GetInactive(c context.Context) ([]models.Vacancy, error) {
	query := `SELECT id, recruiter_id, title, ai_filters, short_link, is_archived 
              FROM vacancies WHERE is_archived = true`
	return r.fetchVacancies(c, query)
}

// GetAll возвращает абсолютно все вакансии
func (r *VacancyRepository) GetAll(c context.Context, id string) ([]models.Vacancy, error) {
	query := `SELECT id, recruiter_id, title, ai_filters, short_link, is_archived 
              FROM vacancies Where recruiter_id = $1`
	return r.fetchVacancies(c, query, id)
}

// GetByID возвращает одну вакансию по ID
func (r *VacancyRepository) GetByID(c context.Context, id string) (*models.Vacancy, error) {
	query := `SELECT id, recruiter_id, title, ai_filters, short_link, is_archived 
              FROM vacancies WHERE id = $1`
	
	var v models.Vacancy
	err := r.DB.QueryRow(c, query, id).Scan(&v.ID, &v.RecruiterID, &v.Title, &v.AIFilters, &v.ShortLink, &v.IsArchived)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &v, nil
}

// Вспомогательная функция для выборки списка вакансий
func (r *VacancyRepository) fetchVacancies(c context.Context, query string, args ...interface{}) ([]models.Vacancy, error) {
	rows, err := r.DB.Query(c, query, args...)
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
	// Если пусто, возвращаем пустой слайс, а не nil (для удобства JSON)
	if vacancies == nil {
		vacancies = []models.Vacancy{}
	}
	return vacancies, nil
}