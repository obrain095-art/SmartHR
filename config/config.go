package config

import "time"

var Config *MapConfig

type MapConfig struct {
	AppHost                  string        `mapstructure:"APP_HOST"`
	DbConnectionString       string        `mapstructure:"DB_CONNECTION_STRING"`
	Admin_Name               string        `mapstructure:"ADMIN_NAME"`
	JwtSecretKey             string        `mapstructure:"JWT_SECRET_KEY"`
	Token_Expiration_Date    time.Duration `mapstructure:"TOKEN_EXPIRATION_DATE"`
	Password_salt            string        `mapstructure:"PASSWORD_SALT"`
	Seed_PIN_Code            int           `mapstructure:"SEED_PIN"`
	Admin_email              string        `mapstructure:"ADMIN_EMAIL"`
	Admin_password           string        `mapstructure:"ADMIN_PASSWORD"`
	WHATSAPP_PHONE_NUMBER_ID string        `mapstructure:"WHATSAPP_PHONE_NUMBER_ID"`
	WHATSAPP_ACCESS_TOKEN    string        `mapstructure:"WHATSAPP_ACCESS_TOKEN"`
	REDIS_ACCESS_PASSWORD    string        `mapstructure:"REDIS_ACCESS_PASSWORD"`
	REDIS_PORT               string        `mapstructure:"REDIS_PORT"`
	REDIS_HOST               string        `mapstructure:"REDIS_HOST"`
}
