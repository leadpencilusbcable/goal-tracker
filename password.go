package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type argon2_params struct {
	version  uint
	time     uint32
	memory   uint32
	threads  uint8
	key_len  uint32
	salt_len uint16
}

var default_params = argon2_params{
	time: 1,
	memory: 64 * 1024,
	threads: 4,
	key_len: 16,
	salt_len: 16,
}

//Hashes a password with argon2 and returns a string
//containing argon2 params for hashing the password
func hashPassword(password string) string {
	salt := generateSalt(default_params.salt_len)
	hash := argon2.IDKey([]byte(password), salt, default_params.time, default_params.memory, default_params.threads, default_params.key_len)

	b64_salt := base64.RawStdEncoding.EncodeToString(salt)
	b64_hash := base64.RawStdEncoding.EncodeToString(hash)

	format_str := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		default_params.memory,
		default_params.time,
		default_params.threads,
		b64_salt,
		b64_hash,
	)

	return format_str
}

func generateSalt(length uint16) []byte {
	salt := make([]byte, length)
	rand.Read(salt)

	return salt
}

func comparePasswordWithHash(password string, password_params string) (bool, error) {
	params, salt, existing_hash, err := extractPasswordParams(password_params)
	in_hash := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, params.key_len)

	if err != nil {
		return false, err
	}

	if subtle.ConstantTimeCompare(in_hash, existing_hash) == 1 {
		return true, nil
	}

	return false, nil
}

func extractPasswordParams(password_params string) (
	params *argon2_params,
	salt []byte,
	hash []byte,
	err error,
) {
	params = &argon2_params{}
	split_params := strings.Split(password_params, "$")

	if len(split_params) != 6 {
		return nil, nil, nil, errors.New("invalid param string")
	}

	if split_params[1] != "argon2id" {
		return nil, nil, nil, errors.New("algorithm is not argon2id")
	}

	_, err = fmt.Sscanf(
		split_params[2],
		"v=%d",
		&params.version,
	)

	if err != nil {
		return nil, nil, nil, errors.New("invalid param string")
	}

	_, err = fmt.Sscanf(
		split_params[3],
		"m=%d,t=%d,p=%d",
		&params.memory,
		&params.time,
		&params.threads,
	)

	if err != nil {
		return nil, nil, nil, errors.New("invalid param string")
	}

	salt, err = base64.RawStdEncoding.DecodeString(split_params[4])

	if err != nil {
		return nil, nil, nil, errors.New("error decoding salt")
	}

	params.salt_len = uint16(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(split_params[5])

	if err != nil {
		return nil, nil, nil, errors.New("error decoding hash")
	}

	params.key_len = uint32(len(hash))

	return params, salt, hash, nil
}

