package main

import (
	"AI_recruit/config"
	"AI_recruit/handlers"
	"AI_recruit/logger"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// @title SmartHR API
// @version 1.0
// @description API сервис для автоматизации рекрутинга.
// @host SmartHR.onrender.com
// @BasePath /
func main() {
	// 1. Инициализация логгера
	l := logger.GetLogger()

	err := loadConfig()
	if err != nil {
		l.Fatal("Failed to load config", zap.Error(err))
	}

	l.Info("Проверка ключа:", zap.String("ai_key", config.Config.AI_Key))
	l.Info("Проверка ключа:", zap.String("DB_string", config.Config.DbConnectionString))
	l.Info("Проверка ключа:", zap.String("host", config.Config.AppHost))

	// 2. Подключение к БД
	conn, err := connectToDb()
	if err != nil {
		l.Fatal("Не удалось подключиться к базе данных", zap.Error(err))
	}
	defer conn.Close()
	l.Info("Успешное подключение к PostgreSQL")

	// 4. Настройка Gin
	r := gin.Default()

	// Настройка CORS
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Разрешаем только нужные фронты
		if origin == "https://smarthr-frontend-t41p.onrender.com" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin") // чтобы браузеры кэш правильно делали
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	handlers.SetupRoutes(r, conn, config.Config.AI_Key)

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
