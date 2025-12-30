package repository

import (
	"AI_recruit/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateRepository struct {
	DB *pgxpool.Pool
}

func (r *TemplateRepository) Create(ctx context.Context, t models.MessageTemplate) error {
	query := `INSERT INTO message_templates (recruiter_id, title, body_text) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(ctx, query, t.RecruiterID, t.Title, t.BodyText)
	return err
}

func (r *TemplateRepository) GetAllByRecruiter(ctx context.Context, recruiterID string) ([]models.MessageTemplate, error) {
	query := `SELECT id, recruiter_id, title, body_text FROM message_templates WHERE recruiter_id = $1`
	rows, err := r.DB.Query(ctx, query, recruiterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.MessageTemplate
	for rows.Next() {
		var t models.MessageTemplate
		if err := rows.Scan(&t.ID, &t.RecruiterID, &t.Title, &t.BodyText); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *TemplateRepository) Update(ctx context.Context, id string, t models.MessageTemplate) error {
	query := `UPDATE message_templates SET title = $1, body_text = $2 WHERE id = $3`
	_, err := r.DB.Exec(ctx, query, t.Title, t.BodyText, id)
	return err
}

func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM message_templates WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	return err
}