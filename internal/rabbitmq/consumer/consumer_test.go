package consumer

import (
	"testing"

	"github.com/protengplus/proteng-user-mgmt/configs"
)

func TestJobDetailURL(t *testing.T) {
	configs.Config.Origin = "http://localhost:5173"

	if got, want := jobDetailURL("abc123"), "http://localhost:5173/dashboard/job-detail/abc123"; got != want {
		t.Errorf("jobDetailURL(id) = %q, want %q", got, want)
	}
	if got, want := jobDetailURL(""), "http://localhost:5173/dashboard"; got != want {
		t.Errorf("jobDetailURL(\"\") = %q, want %q", got, want)
	}
}

func TestJobNotificationSubject(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"COMPLETED", `Your job "my job" has completed`},
		{"FAILED", `Your job "my job" has failed`},
		{"ONGOING", `Update on your job "my job"`},
	}
	for _, tt := range tests {
		if got := jobNotificationSubject("my job", tt.state); got != tt.want {
			t.Errorf("jobNotificationSubject(%q) = %q, want %q", tt.state, got, tt.want)
		}
	}
}
