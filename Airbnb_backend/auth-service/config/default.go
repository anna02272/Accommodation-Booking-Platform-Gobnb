package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	SecretKey string
	EmailFrom string
	SMTPHost  string
	SMTPPass  string
	SMTPPort  int
	SMTPUser  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		fmt.Errorf("couldn't load config")
	}
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		fmt.Errorf("couldn't convert SMTP_PORT to int: %v", err)
	}
	return &Config{
		SecretKey: os.Getenv("SECRET_KEY"),
		EmailFrom: os.Getenv("EMAIL_FROM"),
		SMTPHost:  os.Getenv("SMTP_HOST"),
		SMTPPass:  os.Getenv("SMTP_PASS"),
		SMTPPort:  smtpPort,
		SMTPUser:  os.Getenv("SMTP_USER"),
	}

}
