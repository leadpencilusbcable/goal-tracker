package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const SESSION_ID_LEN_BYTE = 64
const SESSION_ID_EXPIRE_HOURS = 24

//returns the session id as a hex string.
//therefore returned string's len will be 2 times that of byte_len
func generateSessionId(byte_len int) (string, error) {
	session_bytes := make([]byte, byte_len)
	bytes_filled, err := rand.Read(session_bytes)

	if err != nil {
		return "", err
	} else if bytes_filled != byte_len {
		return "", errors.New("rand.Read did not fill the desired amount of bytes")
	}

	hex_session_id := hex.EncodeToString(session_bytes)
	return hex_session_id, nil
}

//keyed by user
var session_ids map[string]string = map[string]string{}
var session_ids_mu sync.RWMutex

func VerifyUser(session_id string) bool {
	session_ids_mu.RLock()
	_, has_session_id := session_ids[session_id]
	session_ids_mu.RUnlock()

	return has_session_id
}

func removeUserSessionId(session_id string) {
	session_ids_mu.Lock()
	delete(session_ids, session_id)
	session_ids_mu.Unlock()
}

func CreateUserSessionId(username string, expires time.Duration) (string, error) {
	session_id, err := generateSessionId(SESSION_ID_LEN_BYTE)

	if err != nil {
		return "", err
	}

	session_ids_mu.Lock()
	session_ids[session_id] = username
	session_ids_mu.Unlock()

	remove_func := func() { removeUserSessionId(session_id) }
	time.AfterFunc(expires, remove_func)

	return session_id, nil
}

