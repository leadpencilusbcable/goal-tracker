package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
)

const SESSION_ID_LEN_BYTE = 64

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

//returns empty string in case of bad auth token
func VerifyUser(db *sql.DB, session_id string) (username string, err error) {
	hash := sha256.Sum256([]byte(session_id))
	username, err = GetSessionId(db, hash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		} else {
			return "", err
		}
	}

	return username, nil
}

func CreateUserSessionId(db *sql.DB, username string) (string, error) {
	session_id, err := generateSessionId(SESSION_ID_LEN_BYTE)

	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(session_id))
	UpsertSessionId(db, username, hash)

	return session_id, nil
}

