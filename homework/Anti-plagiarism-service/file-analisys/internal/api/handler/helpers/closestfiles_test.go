package helpers

import (
	"regexp"
	"strings"
	"testing"
)

func sanitizeForTest(workID string) string {
	trimmed := strings.TrimSpace(workID)
	sanitized := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		case r >= '0' && r <= '9':
			return r
		case r == '_' || r == '-':
			return r
		default:
			return '_'
		}
	}, trimmed)
	if len(sanitized) > maxCollectionName {
		sanitized = sanitized[:maxCollectionName]
	}
	return sanitized
}

func TestCollectionNameForWorkID_Empty(t *testing.T) {
	if _, err := collectionNameForWorkID(" "); err == nil {
		t.Fatal("expected error for empty workID")
	}
}

func TestCollectionNameForWorkID_SanitizedPrefix(t *testing.T) {
	name, err := collectionNameForWorkID("My Work#1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(name, "work_my_work_1_") {
		t.Fatalf("unexpected name prefix: %s", name)
	}
	matched, err := regexp.MatchString(`^work_[a-z0-9_-]+_[0-9a-f]{8}$`, name)
	if err != nil {
		t.Fatalf("regex error: %v", err)
	}
	if !matched {
		t.Fatalf("unexpected name format: %s", name)
	}
}

func TestCollectionNameForWorkID_Truncates(t *testing.T) {
	longID := strings.Repeat("a", maxCollectionName+10)
	name, err := collectionNameForWorkID(longID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sanitized := sanitizeForTest(longID)
	expectedPrefix := "work_" + sanitized + "_"
	if !strings.HasPrefix(name, expectedPrefix) {
		t.Fatalf("unexpected prefix: %s", name)
	}
}
