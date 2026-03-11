package export

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRenderJSON_IncludesTypeAndTimestamp(t *testing.T) {
	t.Parallel()

	content, err := RenderJSON(BoardExport{
		Type:        "board_export",
		GeneratedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		t.Fatalf("invalid json output: %v\ncontent: %s", err, content)
	}
	if decoded["type"] != "board_export" {
		t.Fatalf("unexpected type field: %v", decoded["type"])
	}
	if _, ok := decoded["generated_at"]; !ok {
		t.Fatalf("missing generated_at field: %v", decoded)
	}
}
