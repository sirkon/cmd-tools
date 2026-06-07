package main

import (
	"testing"
)

func Test_isValidUserToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "valid-a",
			token: "abcde",
			want:  true,
		},
		{
			name:  "valid-a-a",
			token: "ab-cd",
			want:  true,
		},
		{
			name:  "valid-a-z",
			token: "ab-cd2",
			want:  true,
		},
		{
			name:  "valid-a-a-z",
			token: "ab-cd-ef2",
			want:  true,
		},
		{
			name:  "invalid-EMPTY",
			token: "",
			want:  false,
		},
		{
			name:  "invalid-SPLITS",
			token: "-",
			want:  false,
		},
		{
			name:  "invalid-TOOSHORT(a)",
			token: "a",
			want:  false,
		},
		{
			name:  "invalid-TOOSHORT(a)-a",
			token: "a-bc",
			want:  false,
		},
		{
			name:  "invalid-TOOSHORT(a)-z",
			token: "a-bc2",
			want:  false,
		},
		{
			name:  "invalid-TOOSHORT_LEADING(z)",
			token: "a2",
			want:  false,
		},
		{
			name:  "invalid-DIGITS",
			token: "23",
			want:  false,
		},
		{
			name:  "invalid-DIGITS-a",
			token: "12-ab",
			want:  false,
		},
		{
			name:  "invalid-a-TOOSHORT(a)",
			token: "ab-c",
			want:  false,
		},
		{
			name:  "invalid-a-TOOSHORT(a)-a",
			token: "ab-c-de",
			want:  false,
		},
		{
			name:  "invalid-a-INVALID(z)",
			token: "ab-cd0",
			want:  false,
		},
		{
			name:  "invalid-INVALIDCHARS(a)-a",
			token: "AB-cd",
			want:  false,
		},
		{
			name:  "invalid-INVALIDCHARS(a)",
			token: "abCD",
			want:  false,
		},
		{
			name:  "invalid-a-UNCLASSIFIED",
			token: "ab-cd23e",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidUserToken(tt.token); got != tt.want {
				t.Errorf("isValidUserToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
