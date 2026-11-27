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

func TestUpdateUserInputApplyTo(t *testing.T) {
	original := func() User {
		return User{
			Email:      "old@b.com",
			Password:   "$2a$10$hashedvalue",
			Name:       "Old",
			Surname:    "Name",
			Role:       []string{"user"},
			UserRole:   "student",
			IsVerified: false,
		}
	}

	cases := []struct {
		name string
		body string
		want User
	}{
		{
			name: "only name sent",
			body: `{"name":"New"}`,
			want: User{Email: "old@b.com", Password: "$2a$10$hashedvalue", Name: "New", Surname: "Name", Role: []string{"user"}, UserRole: "student"},
		},
		{
			name: "protected fields are ignored",
			body: `{"name":"New","role":["admin"],"is_verified":true,"password":"plain"}`,
			want: User{Email: "old@b.com", Password: "$2a$10$hashedvalue", Name: "New", Surname: "Name", Role: []string{"user"}, UserRole: "student"},
		},
		{
			name: "all profile fields sent",
			body: `{"name":"N","surname":"S","user_role":"staff"}`,
			want: User{Email: "old@b.com", Password: "$2a$10$hashedvalue", Name: "N", Surname: "S", Role: []string{"user"}, UserRole: "staff"},
		},
		{
			name: "email is ignored",
			body: `{"email":"new@b.com","name":"N"}`,
			want: User{Email: "old@b.com", Password: "$2a$10$hashedvalue", Name: "N", Surname: "Name", Role: []string{"user"}, UserRole: "student"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var in UpdateUserInput
			if err := json.Unmarshal([]byte(tc.body), &in); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := original()
			in.ApplyTo(&got)

			if got.Email != tc.want.Email || got.Password != tc.want.Password ||
				got.Name != tc.want.Name || got.Surname != tc.want.Surname ||
				got.UserRole != tc.want.UserRole || got.IsVerified != tc.want.IsVerified ||
				!slices.Equal(got.Role, tc.want.Role) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
