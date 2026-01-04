package handlers

import (
	"AI_recruit/config"
	"AI_recruit/docs"
	"AI_recruit/repository"
	"AI_recruit/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, conn *pgxpool.Pool) {

	recRepo := repository.NewRecruiterRepository(conn)
	authRepo := repository.NewAuthRepository(conn)
	portalRepo := repository.NewCandidatePortalRepository(conn)
	candRepo := repository.NewCandidateRepository(conn)

	aiClient := openai.NewClient(config.Config.AI_Key)
	Ai_service := &services.AIService{Client: aiClient}


	recHandler := NewRecruiterHandler(recRepo)
	candHandler := NewCandidateHandler(candRepo)
	authHandler := NewAuthHandler(authRepo)
	portalHandler := NewCandidatePortalHandler(portalRepo, Ai_service)

	templateRepo := repository.NewTemplateRepository(conn)
	templateHandler := NewTemplateHandler(templateRepo)

	r.POST("/applications", portalHandler.Apply)
	r.GET("/my-applications", portalHandler.MyApplications)

	// Зона 4: AI Анализ и Статусы
	apps := r.Group("/applications")
	{
		apps.GET("/:id/ai-data", candHandler.GetAIData)
		apps.PATCH("/:id/status", candHandler.UpdateStatus)
	}

	r.POST("/templates/:id/generate", candHandler.GenerateTelegramText)

	auth := r.Group("/auth")
	{
		auth.POST("/recruiter/signup", authHandler.RecruiterSignup)
		auth.POST("/candidate/signup", authHandler.CandidateSignup)
		auth.POST("/login", authHandler.Login)
	}

	tGroup := r.Group("/templates")
	{
		tGroup.POST("", templateHandler.CreateTemplate)
		tGroup.GET("", templateHandler.ListTemplates)
		tGroup.PUT("/:id", templateHandler.UpdateTemplate)
		tGroup.DELETE("/:id", templateHandler.DeleteTemplate)
	}
	v := r.Group("/vacancies")
	{
		v.GET("", recHandler.ListActiveVacancies) // GET /vacancies
		v.POST("", recHandler.CreateVacancy)      // POST /vacancies
		v.PATCH("/:id/archive", recHandler.ArchiveVacancy)
		v.GET("/:id/applications", recHandler.GetApplications)

		// ИСПРАВЛЕНИЕ: Добавляем префикс /v/ для коротких ссылок, чтобы Gin не путал их с другими GET запросами
		v.GET("/link/:short_link", portalHandler.ViewVacancy)
	}
	// Swagger
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))
}
