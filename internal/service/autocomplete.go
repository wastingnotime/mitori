package service

import (
	"context"
	"slices"
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
)

func StructuredValues(field string) []string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "initiative":
		return toStringSlice(domain.InitiativeValues)
	case "type":
		return toStringSlice(domain.TaskTypeValues)
	case "loop":
		return toStringSlice(domain.LoopValues)
	case "energy", "energy_type":
		return toStringSlice(domain.EnergyTypeValues)
	case "nature":
		return toStringSlice(domain.NatureValues)
	default:
		return nil
	}
}

func PrefixSuggestions(values []string, prefix string) []string {
	p := strings.ToLower(strings.TrimSpace(prefix))
	out := make([]string, 0)
	for _, value := range values {
		v := strings.TrimSpace(value)
		if v == "" {
			continue
		}
		if p == "" || strings.HasPrefix(strings.ToLower(v), p) {
			out = append(out, v)
		}
	}
	slices.SortFunc(out, func(a, b string) int {
		la := strings.ToLower(a)
		lb := strings.ToLower(b)
		if la != lb {
			return strings.Compare(la, lb)
		}
		return strings.Compare(a, b)
	})
	return out
}

func (s *TaskService) StructuredSuggestions(ctx context.Context, field, prefix string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "project":
		snap, err := s.store.Load(ctx)
		if err != nil {
			return nil, err
		}
		seen := make(map[string]struct{})
		projects := make([]string, 0, len(snap.Projects))
		for _, p := range snap.Projects {
			name := strings.TrimSpace(p.Name)
			if name == "" {
				continue
			}
			k := strings.ToLower(name)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			projects = append(projects, name)
		}
		return PrefixSuggestions(projects, prefix), nil
	default:
		return PrefixSuggestions(StructuredValues(field), prefix), nil
	}
}

func toStringSlice[T ~string](in []T) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, string(v))
	}
	return out
}
