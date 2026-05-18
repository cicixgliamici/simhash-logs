package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvalIntValues_UsesFallbackWhenListIsEmpty(t *testing.T) {
	got, err := evalIntValues("", 6, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{6}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected values: got=%v want=%v", got, want)
	}
}

func TestEvalIntValues_TrimsAndDeduplicatesInOrder(t *testing.T) {
	got, err := evalIntValues(" 3, 6,3, 9 ", 0, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{3, 6, 9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected values: got=%v want=%v", got, want)
	}
}

func TestEvalIntValues_RejectsEmptyListEntry(t *testing.T) {
	_, err := evalIntValues("3,,6", 0, "k")
	if err == nil {
		t.Fatal("expected error for empty list entry")
	}
	if !strings.Contains(err.Error(), "empty value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvalIntValues_RejectsOutOfRangeValues(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		fallback int
		valueSet string
	}{
		{name: "negative k", raw: "-1", valueSet: "k"},
		{name: "k over 64", raw: "65", valueSet: "k"},
		{name: "bands over 64", raw: "65", valueSet: "bands"},
		{name: "negative fallback", fallback: -1, valueSet: "bands"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evalIntValues(tt.raw, tt.fallback, tt.valueSet)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestEvaluateLSH_ComputesGroundTruthAndRecall(t *testing.T) {
	sigs := []uint64{
		0,
		1, // dist(0,1)=1
		3, // dist(0,3)=2; dist(1,3)=1
		7, // dist(3,7)=1
	}

	result := evaluateLSH(sigs, 1, 2)

	if result.K != 1 || result.Bands != 2 || result.Records != len(sigs) {
		t.Fatalf("unexpected identity fields: %+v", result)
	}
	if result.BruteComps != 6 {
		t.Fatalf("expected 6 brute comparisons, got %d", result.BruteComps)
	}
	if result.TotalActual != 3 {
		t.Fatalf("expected 3 true pairs, got %d", result.TotalActual)
	}
	if result.LSHMatches != result.TruePositive {
		t.Fatalf("expected all LSH matches to be true positives, got %+v", result)
	}
	if result.RecallPct < 0 || result.RecallPct > 100 {
		t.Fatalf("recall should be a percentage, got %.2f", result.RecallPct)
	}
	if result.ReductionPct < 0 || result.ReductionPct > 100 {
		t.Fatalf("comparison reduction should be a percentage, got %.2f", result.ReductionPct)
	}
}
