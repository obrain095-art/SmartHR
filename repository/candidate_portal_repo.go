package repository

import (
	"AI_recruit/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CandidatePortalRepository struct {
	DB *pgxpool.Pool
}

func NewCandidatePortalRepository(conn *pgxpool.Pool) *CandidatePortalRepository {
	return &CandidatePortalRepository{conn}
}

// UC-C2: Создание или получение профиля кандидата
func (r *CandidatePortalRepository) UpsertCandidate(ctx context.Context, c models.Candidate) (string, error) {
	var id string
	query := `
		INSERT INTO candidates (email, password_hash, telegram_username) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (email) DO UPDATE SET telegram_username = EXCLUDED.telegram_username
		RETURNING id`
	err := r.DB.QueryRow(ctx, query, c.Email, c.PasswordHash, c.TelegramUsername).Scan(&id)
	return id, err
}

// CreateApplication сохраняет данные в транзакции и возвращает ID
func (r *CandidatePortalRepository) CreateApplication(ctx context.Context, app models.Application, ai models.ResumeData) (string, error) {
	// Начинаем транзакцию
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// 1. Вставляем в applications (теперь с ai_score)
	queryApp := `INSERT INTO applications (id, vacancy_id, candidate_id, status, ai_score, applied_at) 
                 VALUES ($1, $2, $3, $4, $5, NOW())`
	
	_, err = tx.Exec(ctx, queryApp, 
		app.ID, 
		app.VacancyID, 
		app.CandidateID, 
		app.Status, 
		ai.AIScore, // Передаем процент совместимости сюда
	)
	if err != nil {
		return "", err
	}

	// 2. Вставляем остальные детали анализа в resume_data
	// Убедись, что в resume_data у тебя нет дублирующей колонки ai_score, если она не нужна
	queryResume := `INSERT INTO resume_data (
                        application_id, 
                        ai_verdict, 
                        parsed_text, 
                        skills_detected
                    ) VALUES ($1, $2, $3, $4)`
	
	_, err = tx.Exec(ctx, queryResume, 
		app.ID, 
		ai.AIVerdict, 
		ai.ParsedText, 
		ai.SkillsDetected,
	)
	if err != nil {
		return "", err
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return app.ID, nil
}
// UC-C3: Сохранение результатов парсинга
func (r *CandidatePortalRepository) SaveResumeData(c context.Context, data models.ResumeData) error {
	query := `INSERT INTO resume_data (application_id, ai_verdict, parsed_text, skills_detected) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.Exec(c, query, data.ApplicationID, data.AIVerdict, data.ParsedText, data.SkillsDetected)
	return err
}

// UC-C3: Проверка активности вакансии перед откликом
func (r *CandidatePortalRepository) IsVacancyActive(c context.Context, vacancyID string) bool {
	var isArchived bool
	query := `SELECT is_archived FROM vacancies WHERE id = $1`
	err := r.DB.QueryRow(c, query, vacancyID).Scan(&isArchived)
	if err != nil || isArchived {
		return false
	}
	return true
}

// UC-C4: Получение списка всех откликов кандидата с актуальными статусами
func (r *CandidatePortalRepository) GetCandidateApplications(c context.Context, candidateID string) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			a.id, 
			v.title as vacancy_title, 
			a.status, 
			a.ai_score,
			a.applied_at 
		FROM applications a
		JOIN vacancies v ON a.vacancy_id = v.id
		WHERE a.candidate_id = $1
		ORDER BY a.applied_at DESC`
	
	rows, err := r.DB.Query(c, query, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, title, status string
		var appliedAt time.Time
		var score int
		if err := rows.Scan(&id, &title, &status, &score, &appliedAt); err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"application_id": id,
			"vacancy":        title,
			"status":         status, // "На рассмотрении ИИ", "Интервью", "Оффер", "Отказ"
			"ai_score":       score,
			"date":           appliedAt,
		})
	}
	return results, nil
}

func (r *CandidatePortalRepository) GetByShortLink(link string) (models.Vacancy, error) {
    var v models.Vacancy
    // Ищем только активные вакансии
    query := `SELECT id, title, ai_filters, short_link FROM vacancies 
              WHERE (short_link = $1 OR title = $1) AND is_archived = false`
    
    err := r.DB.QueryRow(context.Background(), query, link).Scan(
        &v.ID, &v.Title, &v.AIFilters, &v.ShortLink,
    )
    return v, err
}

func (r *CandidatePortalRepository) GetVacancyByID(ctx context.Context, id string) (*models.Vacancy, error) {
    var v models.Vacancy
    query := `SELECT id, ai_filters FROM vacancies WHERE id = $1`
    err := r.DB.QueryRow(ctx, query, id).Scan(&v.ID, &v.AIFilters)
    return &v, err
}