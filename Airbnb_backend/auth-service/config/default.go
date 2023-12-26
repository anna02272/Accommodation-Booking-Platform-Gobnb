package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	SecretKey     string
	EmailFrom     string
	SMTPHost      string
	SMTPPass      string
	SMTPPort      int
	SMTPUser      string
	JaegerAddress string
	ServiceName   string
}

func LoadConfig() *Config {
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Printf("Error converting SMTP_PORT to int: %v", err)
	}
	return &Config{
		SecretKey:     os.Getenv("SECRET_KEY"),
		EmailFrom:     os.Getenv("EMAIL_FROM"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPass:      os.Getenv("SMTP_PASS"),
		SMTPPort:      smtpPort,
		SMTPUser:      os.Getenv("SMTP_USER"),
		JaegerAddress: os.Getenv("JAEGER_ADDRESS"),
		ServiceName:   "auth-service",
	}

}
