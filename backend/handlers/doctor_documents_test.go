package handlers

import (
	"strings"
	"testing"
)

func TestStubFailFilename_Demo(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"report-fail.pdf", true},
		{"FAIL.txt", true},
		{"healthy.pdf", false},
		{"notes.docx", false},
	}
	for _, tc := range cases {
		got := strings.Contains(strings.ToLower(tc.name), "fail")
		if got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.name, got, tc.want)
		}
	}
}
