package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"

	"github.com/k3a/html2text"
	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/models"
	"gopkg.in/gomail.v2"
)

type EmailData struct {
	URL        string
	FirstName  string
	Subject    string
	ExpiryDays int
	JobName    string
	JobState   string
	StageID    string
	StageName  string
}

// 👇 Email template parser
func SendEmail(user *models.User, data *EmailData, temp *template.Template, templateName string) error {

	// Sender data.
	from := configs.Config.EmailFrom
	smtpPass := configs.Config.SMTPPass
	smtpUser := configs.Config.SMTPUser
	to := user.Email
	smtpHost := configs.Config.SMTPHost
	smtpPort := configs.Config.SMTPPort

	var body bytes.Buffer

	if err := temp.ExecuteTemplate(&body, templateName, &data); err != nil {
		log.Fatal("Could not execute template", err)
	}

	m := gomail.NewMessage()

	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", body.String())
	m.AddAlternative("text/plain", html2text.HTML2Text(body.String()))

	fmt.Println(m)
	fmt.Println(smtpHost)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	fmt.Println(d)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send Email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
