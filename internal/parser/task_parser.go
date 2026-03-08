package parser

import (
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
)

type ParsedTaskInput struct {
	Initiative domain.Initiative
	Project    string
	Title      string
}

func ParseTaskInput(input string) ParsedTaskInput {
	raw := strings.TrimSpace(input)
	out := ParsedTaskInput{Title: raw}

	parts := strings.SplitN(raw, "->", 2)
	if len(parts) != 2 {
		return out
	}

	left := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(parts[1])
	if title == "" {
		return out
	}

	leftParts := strings.SplitN(left, "-", 2)
	if len(leftParts) != 2 {
		return out
	}

	initStr := strings.ToLower(strings.TrimSpace(leftParts[0]))
	project := strings.TrimSpace(leftParts[1])
	if !domain.ValidInitiative(domain.Initiative(initStr)) || project == "" {
		return out
	}

	out.Initiative = domain.Initiative(initStr)
	out.Project = project
	out.Title = title
	return out
}
