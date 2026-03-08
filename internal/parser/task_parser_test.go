package parser

import (
	"testing"

	"github.com/wastingnotime/mitori/internal/domain"
)

func TestParseTaskInput_SupportedForms(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		init    domain.Initiative
		project string
		title   string
	}{
		{
			name:    "simple_dash_form",
			input:   "cz - collision playground - allow fullscreen mode",
			init:    domain.InitiativeCZ,
			project: "collision playground",
			title:   "allow fullscreen mode",
		},
		{
			name:    "arrow_form",
			input:   "cz - collision playground -> allow fullscreen mode",
			init:    domain.InitiativeCZ,
			project: "collision playground",
			title:   "allow fullscreen mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTaskInput(tt.input)
			if got.Initiative != tt.init {
				t.Fatalf("initiative mismatch: got=%q want=%q", got.Initiative, tt.init)
			}
			if got.Project != tt.project {
				t.Fatalf("project mismatch: got=%q want=%q", got.Project, tt.project)
			}
			if got.Title != tt.title {
				t.Fatalf("title mismatch: got=%q want=%q", got.Title, tt.title)
			}
		})
	}
}

func TestParseTaskInput_MalformedFallback(t *testing.T) {
	tests := []string{
		"",
		"plain title only",
		"xx - project - title",
		"cz - -> title",
		"cz - collision playground ->",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got := ParseTaskInput(input)
			if got.Initiative != "" {
				t.Fatalf("expected empty initiative for malformed input, got=%q", got.Initiative)
			}
			if got.Project != "" {
				t.Fatalf("expected empty project for malformed input, got=%q", got.Project)
			}
		})
	}
}
