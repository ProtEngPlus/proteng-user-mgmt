package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/k3a/html2text"
	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/models"
	"gopkg.in/gomail.v2"
)

type EmailData struct {
	URL           string
	FirstName     string
	Subject       string
	ExpiryDays    int
	JobName       string
	JobState      string
	StageID       string
	StageName     string
	ExpiryMinutes int
}

func SendEmail(user *models.User, data *EmailData, temp *template.Template, templateName string) error {
	html, err := RenderEmail(temp, templateName, data)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", configs.Config.EmailFrom)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/plain", PlainTextFromHTML(html))
	m.AddAlternative("text/html", html)

	d := gomail.NewDialer(configs.Config.SMTPHost, configs.Config.SMTPPort, configs.Config.SMTPUser, configs.Config.SMTPPass)
	return d.DialAndSend(m)
}

func RenderEmail(temp *template.Template, templateName string, data *EmailData) (string, error) {
	var body bytes.Buffer
	if err := temp.ExecuteTemplate(&body, templateName, data); err != nil {
		return "", fmt.Errorf("render email template %q: %w", templateName, err)
	}
	return body.String(), nil
}

var (
	selfClosingBr = regexp.MustCompile(`(?i)<br\s*/>`)
	closingAnchor = regexp.MustCompile(`(?i)</a\s*>`)
)

func PlainTextFromHTML(html string) string {
	html = selfClosingBr.ReplaceAllString(html, "<br>")
	html = closingAnchor.ReplaceAllString(html, "</a><br>")
	lines := strings.Split(html2text.HTML2Text(html), "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if !blank && len(out) > 0 {
				out = append(out, "")
			}
			blank = true
			continue
		}
		out = append(out, line)
		blank = false
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
