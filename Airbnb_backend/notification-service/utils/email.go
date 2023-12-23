package utils

import (
	"bytes"
	"gopkg.in/gomail.v2"
	"log"
	"notification-service/config"
)

type EmailData struct {
	Subject string
	Text    string
	Email   string
}

func SendEmail(data *EmailData, config *config.Config) error {
	var from = "gobnb@gobnb.com"
	var to = data.Email
	var smtpPass = config.SMTPPass
	var smtpUser = config.SMTPUser
	var smtpHost = config.SMTPHost
	var smtpPort = config.SMTPPort

	var body bytes.Buffer
	body.WriteString(data.Text)

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
