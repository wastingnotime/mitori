package service

import (
	"fmt"

	"github.com/wastingnotime/mitori/internal/domain"
)

func FormatEvent(evt domain.Event) string {
	switch evt.Type {
	case domain.EventLaneChanged:
		from := evt.Payload["from"]
		to := evt.Payload["to"]
		if class, ok := evt.Payload["class"]; ok && fmt.Sprint(class) == string(domain.TransitionRecovery) {
			return fmt.Sprintf("lane_changed (%v <- %v)", to, from)
		}
		return fmt.Sprintf("lane_changed (%v -> %v)", from, to)
	case domain.EventTaskParked:
		return fmt.Sprintf("task_parked (from %v)", evt.Payload["from"])
	case domain.EventTaskUnparked:
		return fmt.Sprintf("task_unparked (to %v)", evt.Payload["to"])
	case domain.EventTaskArchived:
		if from, ok := evt.Payload["from_lane"]; ok {
			return fmt.Sprintf("task_archived (from %v)", from)
		}
		return "task_archived"
	case domain.EventTaskUpdated:
		if action, ok := evt.Payload["action"]; ok {
			return fmt.Sprintf("task_updated (%v)", action)
		}
		return "task_updated"
	default:
		return string(evt.Type)
	}
}
