package handlers

import (
	"AI_recruit/docs"
	"AI_recruit/repository"
	"AI_recruit/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, conn *pgxpool.Pool, AI_KEY string) {

	recRepo := repository.NewVacancyRepository(conn)
	authRepo := repository.NewAuthRepository(conn)
	portalRepo := repository.NewCandidatePortalRepository(conn)
	candRepo := repository.NewCandidateRepository(conn)

	aiClient := openai.NewClient(AI_KEY)
	Ai_service := &services.AIService{Client: aiClient}

	recHandler := NewVacancyHandler(recRepo)
	candHandler := NewCandidateHandler(candRepo)
	authHandler := NewAuthHandler(authRepo)
	portalHandler := NewCandidatePortalHandler(portalRepo, Ai_service)

	templateRepo := repository.NewTemplateRepository(conn)
	templateHandler := NewTemplateHandler(templateRepo)

	r.POST("/applications", portalHandler.Apply)            //соискатели
	r.GET("/my-applications", portalHandler.MyApplications) //соискатели

	// Зона 4: AI Анализ и Статусы
	apps := r.Group("/applications")
	{
		apps.GET("/:id/vacancy", candHandler.GetApplicationDetails) //рекуретры
		apps.GET("/:id", candHandler.GetListByVacancy)              //рекрутеры
		apps.GET("/:id/ai-data", candHandler.GetAIData)             //рекрутеры
		apps.PATCH("/:id/status", candHandler.UpdateStatus)         //рекрутеры
	}

	r.POST("/templates/:id/generate", candHandler.GenerateTelegramText) //рекрутеры

	auth := r.Group("/auth")
	{
		auth.POST("/recruiter/signup", authHandler.RecruiterSignup) //все
		auth.POST("/candidate/signup", authHandler.CandidateSignup) //все
		auth.POST("/login", authHandler.Login)                      //все
	}

	tGroup := r.Group("/templates")
	{
		tGroup.POST("", templateHandler.CreateTemplate)       // рекрутеры
		tGroup.GET("", templateHandler.ListTemplates)         // рекрутеры
		tGroup.PUT("/:id", templateHandler.UpdateTemplate)    // рекрутеры
		tGroup.DELETE("/:id", templateHandler.DeleteTemplate) // рекрутеры
	}
	v := r.Group("/vacancies")
	{
		v.GET("/all", recHandler.ListAllVacancies)             //рекрутеры
		v.GET("/active", recHandler.ListActiveVacancies)       //все
		v.GET("/inactive", recHandler.ListInactiveVacancies)   //рекрутеры
		v.POST("", recHandler.CreateVacancy)                   //рекрутеры
		v.GET("/:id", recHandler.ListOneVacancy)               //все
		v.PATCH("/:id/archive", recHandler.ArchiveVacancy)     //рекрутеры
		v.PATCH("/:id/dearchive", recHandler.DeArchiveVacancy) //рекрутеры
		v.GET("/:id/applications", recHandler.GetApplications) //рекрутеры

		// ИСПРАВЛЕНИЕ: Добавляем префикс /v/ для коротких ссылок, чтобы Gin не путал их с другими GET запросами
		v.GET("/link/:short_link", portalHandler.ViewVacancy) //все
	}
	// Swagger
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))
}
