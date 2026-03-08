package parser

import (
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
)

type ParsedTaskInput struct {
	Initiative domain.Initiative
	Project    string
	Title      string
	TaskType   domain.TaskType
}

func ParseTaskInput(input string) ParsedTaskInput {
	raw := strings.TrimSpace(input)
	out := ParsedTaskInput{
		Title:    raw,
		TaskType: InferTaskType(raw),
	}

	if parsed, ok := parseArrowSyntax(raw); ok {
		return parsed
	}
	if parsed, ok := parseSimpleSyntax(raw); ok {
		return parsed
	}
	return out
}

func parseArrowSyntax(raw string) (ParsedTaskInput, bool) {
	parts := strings.SplitN(raw, "->", 2)
	if len(parts) != 2 {
		return ParsedTaskInput{}, false
	}

	left := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(parts[1])
	if title == "" {
		return ParsedTaskInput{}, false
	}

	leftParts := strings.SplitN(left, "-", 2)
	if len(leftParts) != 2 {
		return ParsedTaskInput{}, false
	}

	initStr := strings.ToLower(strings.TrimSpace(leftParts[0]))
	project := strings.TrimSpace(leftParts[1])
	if !domain.ValidInitiative(domain.Initiative(initStr)) || project == "" {
		return ParsedTaskInput{}, false
	}

	return ParsedTaskInput{
		Initiative: domain.Initiative(initStr),
		Project:    project,
		Title:      title,
		TaskType:   InferTaskType(title),
	}, true
}

func parseSimpleSyntax(raw string) (ParsedTaskInput, bool) {
	parts := strings.SplitN(raw, " - ", 3)
	if len(parts) != 3 {
		return ParsedTaskInput{}, false
	}
	initStr := strings.ToLower(strings.TrimSpace(parts[0]))
	project := strings.TrimSpace(parts[1])
	title := strings.TrimSpace(parts[2])
	if !domain.ValidInitiative(domain.Initiative(initStr)) || project == "" || title == "" {
		return ParsedTaskInput{}, false
	}
	return ParsedTaskInput{
		Initiative: domain.Initiative(initStr),
		Project:    project,
		Title:      title,
		TaskType:   InferTaskType(title),
	}, true
}

func InferTaskType(title string) domain.TaskType {
	t := strings.ToLower(strings.TrimSpace(title))
	if t == "" {
		return domain.TaskTypeChore
	}

	if hasAnyToken(t, []string{
		"bug", "fix", "hotfix", "broken", "break", "error", "crash", "issue", "regress", "repair", "patch",
	}) {
		return domain.TaskTypeFix
	}
	if hasAnyToken(t, []string{
		"refactor", "refact", "cleanup", "clean up", "restructure", "rename", "extract", "simplify",
	}) {
		return domain.TaskTypeRefact
	}
	if hasAnyToken(t, []string{
		"research", "investigate", "explore", "spike", "learn", "understand", "discover", "diagnose",
	}) {
		return domain.TaskTypeDiscovery
	}
	if hasAnyToken(t, []string{
		"add", "implement", "enable", "support", "allow", "create", "build", "introduce",
	}) {
		return domain.TaskTypeFeat
	}
	return domain.TaskTypeChore
}

func hasAnyToken(s string, tokens []string) bool {
	for _, token := range tokens {
		if strings.Contains(s, token) {
			return true
		}
	}
	return false
}
