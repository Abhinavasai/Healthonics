package handlers

import (
	"math"
	"testing"
)

func TestCosineSimilarity_identical(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}
	if math.Abs(CosineSimilarity(a, b)-1) > 1e-9 {
		t.Fatalf("expected 1, got %v", CosineSimilarity(a, b))
	}
}

func TestCosineSimilarity_orthogonal(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{0, 1}
	if math.Abs(CosineSimilarity(a, b)) > 1e-9 {
		t.Fatalf("expected 0, got %v", CosineSimilarity(a, b))
	}
}
