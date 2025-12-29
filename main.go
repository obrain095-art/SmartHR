package main

import (
	"AI_recruit/config"
	"AI_recruit/docs"
	"AI_recruit/handlers"
	"AI_recruit/logger"
	"AI_recruit/repository"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title AI Recruiting API
// @version 1.0
// @description API сервис для автоматизации рекрутинга.
// @host ai-recruiting-api.onrender.com
// @BasePath /
func main() {
	// 1. Инициализация логгера
	l := logger.GetLogger()

	err := loadConfig()
	if err != nil {
		l.Fatal("Failed to load config", zap.Error(err))
	}

	// 2. Подключение к БД
	pool, err := connectToDb()
	if err != nil {
		l.Fatal("Не удалось подключиться к базе данных", zap.Error(err))
	}
	defer pool.Close()
	l.Info("Успешное подключение к PostgreSQL")

	// 3. Инициализация слоев (Repo -> Handler)
	// Убедитесь, что в структурах репозиториев поле называется Pool (или DB) и имеет тип *pgxpool.Pool
	vacRepo := &repository.VacancyRepository{DB: pool}
	candRepo := &repository.CandidateRepository{DB: pool}
	authRepo := &repository.AuthRepository{DB: pool}
	portalRepo := &repository.CandidatePortalRepository{DB: pool}

	vacHandler := &handlers.RecruiterHandler{Repo: vacRepo}
	candHandler := &handlers.CandidateHandler{Repo: candRepo}
	authHandler := &handlers.AuthHandler{Repo: authRepo}
	portalHandler := &handlers.CandidatePortalHandler{Repo: portalRepo}

	// 4. Настройка Gin
	r := gin.Default()

	// Настройка CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 5. Определение маршрутов

	// Зона 1: Авторизация
	auth := r.Group("/auth")
	{
		auth.POST("/recruiter/signup", authHandler.RecruiterSignup)
		auth.POST("/candidate/signup", authHandler.CandidateSignup)
		auth.POST("/login", authHandler.Login)
	}

	// Зона 2: Вакансии
	// Чтобы избежать конфликта с :short_link, мы объединяем все в одну группу
	v := r.Group("/vacancies")
	{
		v.GET("", vacHandler.ListActiveVacancies) // GET /vacancies
		v.POST("", vacHandler.CreateVacancy)      // POST /vacancies
		v.PATCH("/:id/archive", vacHandler.ArchiveVacancy)
		v.GET("/:id/applications", vacHandler.GetApplications)

		// ИСПРАВЛЕНИЕ: Добавляем префикс /v/ для коротких ссылок, чтобы Gin не путал их с другими GET запросами
		v.GET("/link/:short_link", portalHandler.ViewVacancy)
	}

	// Зона 3: Подача и Трекинг
	r.POST("/applications", portalHandler.Apply)
	r.GET("/my-applications", portalHandler.MyApplications)

	// Зона 4: AI Анализ и Статусы
	apps := r.Group("/applications")
	{
		apps.GET("/:id/ai-data", candHandler.GetAIData)
		apps.PATCH("/:id/status", candHandler.UpdateStatus)
	}

	r.POST("/templates/:id/generate", candHandler.GenerateTelegramText)

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

	// 6. Запуск сервера
	l.Info("Сервер запущен на порту :8080")
	if err := r.Run(config.Config.AppHost); err != nil {
		l.Fatal("Ошибка при запуске сервера", zap.Error(err))
	}
}

func connectToDb() (*pgxpool.Pool, error) {
	// ИСПРАВЛЕНИЕ: Порт изменен на 5432 (стандарт для Postgres)
	// Также убедитесь, что имя БД 'AI_recruiting' написано верно (регистр важен)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(config.Config.DbConnectionString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func loadConfig() error {
	// Указываем путь к .env файлу
	viper.SetConfigFile(".env")

	// Загружаем переменные из .env, если он есть (необязательно)
	_ = viper.ReadInConfig() // не падаем, если файла нет

	// Читаем переменные окружения (например, из Railway)
	viper.AutomaticEnv()

	// Мапим переменные в структуру
	var mapConfig config.MapConfig
	err := viper.Unmarshal(&mapConfig)
	if err != nil {
		return err
	}

	config.Config = &mapConfig
	return nil

}
