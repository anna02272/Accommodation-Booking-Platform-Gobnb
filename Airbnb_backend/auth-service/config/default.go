package config

//
//import (
//	"github.com/spf13/viper"
//)
//
//type Config struct {
//	SecretKey string `mapstructure:"SECRET_KEY"`
//	EmailFrom string `mapstructure:"EMAIL_FROM"`
//	SMTPHost  string `mapstructure:"SMTP_HOST"`
//	SMTPPass  string `mapstructure:"SMTP_PASS"`
//	SMTPPort  int    `mapstructure:"SMTP_PORT"`
//	SMTPUser  string `mapstructure:"SMTP_USER"`
//}
//
//func LoadConfig(path string) (config Config, err error) {
//	viper.AddConfigPath(path)
//	viper.SetConfigType("env")
//	//viper.SetConfigName("app")
//
//	viper.AutomaticEnv()
//
//	err = viper.ReadInConfig()
//	if err != nil {
//		return
//	}
//
//	err = viper.Unmarshal(&config)
//	return
//}

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
