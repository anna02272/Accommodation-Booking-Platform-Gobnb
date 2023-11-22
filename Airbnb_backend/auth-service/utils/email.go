package utils

import (
	"auth-service/config"
	"auth-service/domain"
	"bytes"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
)

type EmailData struct {
	URL      string
	Username string
	Subject  string
}

func SendEmail(user *domain.Credentials, data *EmailData, config *config.Config) error {
	var from = "gobnb@gobnb.com"
	var to = user.Email

	var smtpPass = config.SMTPPass
	var smtpUser = config.SMTPUser
	var smtpHost = config.SMTPHost
	var smtpPort = config.SMTPPort

	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("Hi %s,\n", data.Username))
	body.WriteString("your code is: \n")
	body.WriteString(fmt.Sprintf("%s\n", data.URL))

	m := gomail.NewMessage()

	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/plain", body.String())
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	err := d.DialAndSend(m)
	if err != nil {
		log.Printf("Could not send email: %v", err)
		return err
	}
	return nil
}
