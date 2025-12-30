package models

import "time"

// Рекрутер (UC-0)
type Recruiter struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // Скрываем при сериализации в JSON
	CompanyName  string `json:"company_name"`
}

// Соискатель (UC-C2)
type Candidate struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	PasswordHash     string `json:"-"`
	TelegramUsername string `json:"telegram_username"`
}

// Вакансия (UC-1, UC-6)
type Vacancy struct {
	ID          string `json:"id"`
	RecruiterID string `json:"recruiter_id"`
	Title       string `json:"title"`
	AIFilters   string `json:"ai_filters"`
	ShortLink   string `json:"short_link"`
	IsArchived  bool   `json:"is_archived"`
}

type VacancyCreateRequest struct {
	RecruiterID string `json:"recruiter_id"`
	Title       string `json:"title"`
	AIFilters   string `json:"ai_filters"`
	ShortLink   string `json:"short_link"`
	IsArchived  bool   `json:"is_archived"`
}

// Отклик (UC-C3, UC-C4)
// Отклик (UC-C3, UC-C4)
type Application struct {
	ID            string    `json:"id"`
	VacancyID     string    `json:"vacancy_id"`
	CandidateID   string    `json:"candidate_id"`
	CandidateName string    `json:"candidate_name,omitempty"` // Для отображения в списке рекрутера
	Status        string    `json:"status"`
	AIScore       int       `json:"ai_score"`
	AppliedAt     time.Time `json:"applied_at"`
}

type ResumeData struct {
	ID             string `json:"id"`
	ApplicationID  string `json:"application_id"`
	AIVerdict      string `json:"ai_verdict"`
	ParsedText     string `json:"parsed_text"`
	SkillsDetected string `json:"skills_detected"`
	AIScore        int    `json:"ai_score"` // Добавлено поле
}

// Шаблоны сообщений (UC-5)
type MessageTemplate struct {
	ID          string `json:"id"`
	RecruiterID string `json:"recruiter_id"`
	Title       string `json:"title"`
	BodyText    string `json:"body_text"`
}
type UserRole string

const (
	RoleRecruiter UserRole = "recruiter"
	RoleCandidate UserRole = "candidate"
)

type UserAuthResponse struct {
	Token  string   `json:"token"`
	Role   UserRole `json:"role"`
	UserID string   `json:"user_id"`
}
