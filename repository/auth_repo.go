package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository struct {
	DB *pgxpool.Pool
}

func NewAuthRepository(conn *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{conn}
}

// Регистрация рекрутера
func (r *AuthRepository) CreateRecruiter(c context.Context, email, password, company string) (string, error) {
	var id string
	query := `INSERT INTO recruiters (email, password_hash, company_name) 
              VALUES ($1, $2, $3) RETURNING id`
	err := r.DB.QueryRow(c, query, email, password, company).Scan(&id)
	return id, err
}

// Регистрация соискателя
func (r *AuthRepository) CreateCandidate(c context.Context, email, password, tg string) (string, error) {
	var id string
	query := `INSERT INTO candidates (email, password_hash, telegram_username) 
              VALUES ($1, $2, $3) RETURNING id`
	err := r.DB.QueryRow(c, query, email, password, tg).Scan(&id)
	return id, err
}

// Универсальный вход: ищет сначала в рекрутерах, затем в кандидатах
func (r *AuthRepository) Authenticate(ctx context.Context, email, password string) (interface{}, error) {
	var id, hashedPassword string
	
	// 1. Сначала ищем хеш в таблице рекрутеров
	query := `SELECT id, password_hash FROM recruiters WHERE email = $1`
	err := r.DB.QueryRow(ctx, query, email).Scan(&id, &hashedPassword)
	
	// 2. Если не нашли, ищем в кандидатах
	if err != nil {
		query = `SELECT id, password_hash FROM candidates WHERE email = $1`
		err = r.DB.QueryRow(ctx, query, email).Scan(&id, &hashedPassword)
	}

	if err != nil {
		return nil, errors.New("user not found")
	}

	// 3. Сравниваем введенный пароль с хешем из базы
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password") // Пароли не совпали
	}

	return map[string]string{"id": id, "email": email}, nil
}


// AuthenticateRecruiter проверяет данные в таблице recruiters
func (r *AuthRepository) AuthenticateRecruiter(ctx context.Context, email, password string) (map[string]string, error) {
	var id, hashedPassword, company string
	query := `SELECT id, password_hash, company_name FROM recruiters WHERE email = $1`
	err := r.DB.QueryRow(ctx, query, email).Scan(&id, &hashedPassword, &company)
	if err != nil {
		return nil, errors.New("recruiter not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return map[string]string{
		"id":           id,
		"email":        email,
		"role":         "recruiter",
		"company_name": company,
	}, nil
}

// AuthenticateCandidate проверяет данные в таблице candidates
func (r *AuthRepository) AuthenticateCandidate(ctx context.Context, email, password string) (map[string]string, error) {
	var id, hashedPassword, tg string
	query := `SELECT id, password_hash, telegram_username FROM candidates WHERE email = $1`
	err := r.DB.QueryRow(ctx, query, email).Scan(&id, &hashedPassword, &tg)
	if err != nil {
		return nil, errors.New("candidate not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return map[string]string{
		"id":                id,
		"email":             email,
		"role":              "candidate",
		"telegram_username": tg,
	}, nil
}


