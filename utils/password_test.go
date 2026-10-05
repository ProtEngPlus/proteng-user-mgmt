package utils

import "testing"

func TestValidatePasswordStrength(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Passw0rd", false},
		{"minimum valid length", "Aa123456", false},
		{"too short", "Pas0rd", true},
		{"no uppercase", "password1", true},
		{"no lowercase", "PASSWORD1", true},
		{"no digit", "Password", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePasswordStrength(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
		})
	}
}
