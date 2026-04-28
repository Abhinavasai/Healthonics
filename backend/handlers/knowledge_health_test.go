package handlers

import "testing"

func TestKnowledgeHealthScore_LinearDrop(t *testing.T) {
	if KnowledgeHealthScore(0, 180) != 100 {
		t.Fatal("fresh review")
	}
	if KnowledgeHealthScore(90, 180) != 50 {
		t.Fatalf("half interval: got %d", KnowledgeHealthScore(90, 180))
	}
	if KnowledgeHealthScore(180, 180) != 0 {
		t.Fatalf("at interval: got %d", KnowledgeHealthScore(180, 180))
	}
	if KnowledgeHealthScore(400, 180) != 0 {
		t.Fatal("beyond interval clamps to 0")
	}
}

func TestKnowledgeIsStale(t *testing.T) {
	if KnowledgeIsStale(180, 180) {
		t.Fatal("equal boundary not stale")
	}
	if !KnowledgeIsStale(181, 180) {
		t.Fatal("beyond boundary stale")
	}
}
