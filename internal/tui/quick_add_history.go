package tui

import "strings"

type quickAddHistory struct {
	items    []string
	idx      int
	draft    string
	browsing bool
}

func (h *quickAddHistory) Add(raw string) {
	entry := strings.TrimSpace(raw)
	if entry == "" {
		return
	}
	if len(h.items) > 0 && h.items[len(h.items)-1] == entry {
		h.idx = len(h.items)
		h.browsing = false
		h.draft = ""
		return
	}
	h.items = append(h.items, entry)
	h.idx = len(h.items)
	h.browsing = false
	h.draft = ""
}

func (h *quickAddHistory) Prev(current string) string {
	if len(h.items) == 0 {
		return current
	}
	if !h.browsing {
		h.draft = current
		h.idx = len(h.items)
		h.browsing = true
	}
	if h.idx > 0 {
		h.idx--
	}
	return h.items[h.idx]
}

func (h *quickAddHistory) Next(current string) string {
	if len(h.items) == 0 {
		return current
	}
	if !h.browsing {
		return current
	}
	if h.idx < len(h.items)-1 {
		h.idx++
		return h.items[h.idx]
	}
	h.browsing = false
	h.idx = len(h.items)
	return h.draft
}
