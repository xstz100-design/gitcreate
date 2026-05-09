package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 全局应用配置（从环境变量 / .env 文件读取）
type Config struct {
	DatabaseURL string

	// JWT
	SecretKey                string
	AccessTokenExpireMinutes int

	// CORS
	AllowedOrigins []string

	// Telegram
	TGBotToken    string
	TGBotUsername string

	// Site
	SiteURL string

	// 货币汇率
	USDToKHRRate float64
}

var C *Config

func Load(envFile string) {
	// 尝试加载 .env 文件，找不到时不报错
	_ = godotenv.Load(envFile)

	C = &Config{
		DatabaseURL:              getEnv("DATABASE_URL", "./cambodia_wholesale.db"),
		SecretKey:                getEnv("SECRET_KEY", "your-secret-key-please-change-in-production-123456789"),
		AccessTokenExpireMinutes: getInt("ACCESS_TOKEN_EXPIRE_MINUTES", 10080),
		AllowedOrigins:           getStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		TGBotToken:               getEnv("TG_BOT_TOKEN", "8584315237:AAEuLvseKLQ7ZLa1SZKRQo2MaFg_1s3eRgc"),
		TGBotUsername:            getEnv("TG_BOT_USERNAME", "testshopmouy_bot"),
		SiteURL:                  getEnv("SITE_URL", "https://khmerai.cn"),
		USDToKHRRate:             getFloat("USD_TO_KHR_RATE", 4000.0),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func getStringSlice(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return def
	}
	return result
}
