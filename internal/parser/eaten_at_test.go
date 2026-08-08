package parser

import (
	"testing"
	"time"
)

func TestResolveEatenAt_YesterdayDate(t *testing.T) {
	now := time.Date(2026, 8, 6, 18, 30, 0, 0, MealLocation())
	got, err := ResolveEatenAt("2026-08-05", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	local := got.In(MealLocation())
	if local.Format("2006-01-02") != "2026-08-05" {
		t.Fatalf("expected 2026-08-05, got %s", local.Format("2006-01-02"))
	}
	if local.Hour() != 18 || local.Minute() != 30 {
		t.Fatalf("expected current clock time kept, got %02d:%02d", local.Hour(), local.Minute())
	}
}

func TestResolveEatenAt_Empty(t *testing.T) {
	got, err := ResolveEatenAt("", time.Now().UTC())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsZero() {
		t.Fatalf("expected zero time")
	}
}

func TestResolveEatenAt_Invalid(t *testing.T) {
	_, err := ResolveEatenAt("not-a-date", time.Now().UTC())
	if err == nil {
		t.Fatal("expected error")
	}
}
