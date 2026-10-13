package utils

import (
	"html/template"
	"strings"
	"testing"
)

func TestRenderEmail(t *testing.T) {
	temp := template.Must(template.ParseGlob("../templates/*.html"))
	data := &EmailData{
		URL:           "https://example.com/link?token=abc",
		FirstName:     "Mary Ann",
		Subject:       "Subject",
		ExpiryDays:    7,
		ExpiryMinutes: 15,
		JobName:       "my job",
		JobState:      "COMPLETED",
		StageID:       "3",
		StageName:     "mutation",
	}

	for _, name := range []string{"verificationEmail", "resetPassword.html", "passwordChanged", "jobStatusNotification"} {
		t.Run(name, func(t *testing.T) {
			html, err := RenderEmail(temp, name, data)
			if err != nil {
				t.Fatalf("RenderEmail(%q) error = %v", name, err)
			}
			if !strings.Contains(strings.ToLower(html), "<html") {
				t.Errorf("RenderEmail(%q) is not a full HTML document", name)
			}
			if !strings.Contains(html, "Hi Mary Ann,") {
				t.Errorf("RenderEmail(%q) does not greet the user by name", name)
			}
		})
	}

	if _, err := RenderEmail(temp, "noSuchTemplate", data); err == nil {
		t.Error("RenderEmail with an unknown template should return an error")
	}
}

func TestPlainTextFromHTML(t *testing.T) {
	html := `<p>
	    Hi Mary,</p>
	  <table><tr><td><a href="https://example.com/reset"
	    >Reset password</a
	  ></td></tr></table>
	  <p>This link will expire.</p>


	  <p>Best regards,<br />Team</p>`

	want := "Hi Mary,\n\nhttps://example.com/reset\nThis link will expire.\n\nBest regards,\nTeam"
	if got := PlainTextFromHTML(html); got != want {
		t.Errorf("PlainTextFromHTML() =\n%q\nwant\n%q", got, want)
	}
}
