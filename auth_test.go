package main

import "testing"

func TestGenerateSessionId(t *testing.T) {
	session_id, err := generateSessionId(64)

	if err != nil {
		t.Errorf("error generating session_id, err: %s", err.Error())
	}

	if len(session_id) != 64 * 2 {
		t.Errorf("session_id len is not as expected. expected: %d, got: %d", 64 * 2, len(session_id))
	}
}

