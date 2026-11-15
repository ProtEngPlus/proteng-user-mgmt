package models

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestRegisterInputToUser(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"normal sign up", `{"email":"a@b.com","password":"Pass-word-1","name":"A","surname":"B","user_role":"student"}`},
		{"client sends admin role", `{"email":"a@b.com","password":"Pass-word-1","role":["admin"]}`},
		{"client sends is_verified", `{"email":"a@b.com","password":"Pass-word-1","is_verified":true}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var in RegisterInput
			if err := json.Unmarshal([]byte(tc.body), &in); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := in.ToUser()

			if !slices.Equal(got.Role, []string{"user"}) {
				t.Errorf("Role = %v, want [user]", got.Role)
			}
			if got.IsVerified {
				t.Error("IsVerified = true, want false")
			}
			if got.Email != in.Email || got.Name != in.Name || got.Surname != in.Surname || got.UserRole != in.UserRole {
				t.Errorf("profile fields not copied: got %+v from %+v", got, in)
			}
		})
	}
}
