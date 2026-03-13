package service

import (
	"testing"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
)

func TestFormatEvent_LaneChangedDirection(t *testing.T) {
	t.Parallel()

	normal := FormatEvent(domain.Event{
		Type:      domain.EventLaneChanged,
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"from":  domain.LaneBacklog,
			"to":    domain.LaneTodo,
			"class": domain.TransitionNormal,
		},
	})
	if normal != "lane_changed (backlog -> todo)" {
		t.Fatalf("unexpected normal formatting: %q", normal)
	}

	recovery := FormatEvent(domain.Event{
		Type:      domain.EventLaneChanged,
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"from":  domain.LaneTodo,
			"to":    domain.LaneBacklog,
			"class": domain.TransitionRecovery,
		},
	})
	if recovery != "lane_changed (backlog <- todo)" {
		t.Fatalf("unexpected recovery formatting: %q", recovery)
	}
}
