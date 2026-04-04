package pime

import "testing"

func TestParseRequestAcceptsNumericKeyStates(t *testing.T) {
	req, err := ParseRequest([]byte(`{
		"method": "onKeyDown",
		"seqNum": 1,
		"keyStates": [0, 1, 0, 2]
	}`))
	if err != nil {
		t.Fatalf("ParseRequest returned error: %v", err)
	}

	want := []bool{false, true, false, true}
	if len(req.KeyStates) != len(want) {
		t.Fatalf("expected %d key states, got %d", len(want), len(req.KeyStates))
	}
	for i, expected := range want {
		if bool(req.KeyStates[i]) != expected {
			t.Fatalf("expected keyStates[%d]=%t, got %t", i, expected, bool(req.KeyStates[i]))
		}
	}
}

func TestParseRequestAcceptsBooleanKeyStates(t *testing.T) {
	req, err := ParseRequest([]byte(`{
		"method": "onKeyDown",
		"seqNum": 1,
		"keyStates": [true, false, true]
	}`))
	if err != nil {
		t.Fatalf("ParseRequest returned error: %v", err)
	}

	want := []bool{true, false, true}
	if len(req.KeyStates) != len(want) {
		t.Fatalf("expected %d key states, got %d", len(want), len(req.KeyStates))
	}
	for i, expected := range want {
		if bool(req.KeyStates[i]) != expected {
			t.Fatalf("expected keyStates[%d]=%t, got %t", i, expected, bool(req.KeyStates[i]))
		}
	}
}
