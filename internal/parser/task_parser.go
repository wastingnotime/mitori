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
	out.TaskType = InferTaskType(title)
	return out
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
