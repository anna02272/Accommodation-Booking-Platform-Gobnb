package utils

import (
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

func SendEmail(user *domain.Credentials, data *EmailData) error {
	var from = "gobnb@gobnb.com"
	var to = user.Email

	var smtpPass = "8aa59d58ea56fd"
	var smtpUser = "86d650c613b23b"
	var smtpHost = "sandbox.smtp.mailtrap.io"
	var smtpPort = 587

	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("Hi %s,\n", data.Username))
	body.WriteString("Your code is: \n")
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
