package rime

import "testing"

func TestCandidateEntryFromItemMapsDictionaryLookupComment(t *testing.T) {
	raw := "\u000c\u000d1,好,hou2,,,,,,,adj,yue,,,,,good; well"
	entry := candidateEntryFromItem(candidateItem{Text: "好", Comment: raw}, "hou")
	if entry.RawLookupComment != raw {
		t.Fatalf("expected raw lookup comment preserved, got %q", entry.RawLookupComment)
	}
	if entry.Comment != "" {
		t.Fatalf("expected display comment empty for dictionary rows, got %q", entry.Comment)
	}
	if entry.Jyutping != "hou2" {
		t.Fatalf("expected jyutping hou2, got %q", entry.Jyutping)
	}
	if entry.InputCode != "hou" {
		t.Fatalf("expected input code hou, got %q", entry.InputCode)
	}
}

func TestCandidateEntryFromItemKeepsPlainComment(t *testing.T) {
	entry := candidateEntryFromItem(candidateItem{Text: "你", Comment: "nei5"}, "nei")
	if entry.Comment != "nei5" {
		t.Fatalf("expected plain comment, got %q", entry.Comment)
	}
	if entry.RawLookupComment != "" {
		t.Fatalf("expected empty raw lookup comment, got %q", entry.RawLookupComment)
	}
}
