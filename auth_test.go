package main

import (
	"testing"
	"time"
)

func TestGenerateSessionId(t *testing.T) {
	session_id, err := generateSessionId(64)

	if err != nil {
		t.Errorf("error generating session_id, err: %s", err.Error())
	}

	if len(session_id) != 64 * 2 {
		t.Errorf("session_id len is not as expected. expected: %d, got: %d", 64 * 2, len(session_id))
	}
}

func TestSessionIdFlow(t *testing.T) {
	session_id, err := CreateUserSessionId("username", time.Second)

	if err != nil {
		t.Errorf("error generating session_id, err: %s", err.Error())
	}

	authorised := VerifyUser(session_id)

	if !authorised {
		t.Errorf("user should be authorised after generating session_id")
	}

	//simulate token expiring
	time.Sleep(time.Second * 2)

	authorised = VerifyUser(session_id)

	if authorised {
		t.Errorf("user should NOT be authorised after waiting the expiry time")
	}
}

