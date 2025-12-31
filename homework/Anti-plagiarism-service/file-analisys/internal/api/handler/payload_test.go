package handler

import "testing"

func TestPayloadString(t *testing.T) {
	if val, ok := payloadString(nil, "key"); ok || val != "" {
		t.Fatalf("expected empty result for nil payload, got %q, %t", val, ok)
	}

	payload := map[string]interface{}{
		"str": "value",
		"num": 10,
	}
	if val, ok := payloadString(payload, "missing"); ok || val != "" {
		t.Fatalf("expected empty result for missing key, got %q, %t", val, ok)
	}
	if val, ok := payloadString(payload, "num"); ok || val != "" {
		t.Fatalf("expected empty result for non-string value, got %q, %t", val, ok)
	}
	if val, ok := payloadString(payload, "str"); !ok || val != "value" {
		t.Fatalf("unexpected result: %q, %t", val, ok)
	}
}
