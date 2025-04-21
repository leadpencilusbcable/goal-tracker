package main

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestConstructGoalInsertQuery(t *testing.T) {
	goals := []Goal{
		{
			title: "title",
			end_date: time.Now(),
			notes: "",
		},
		{
			title: "title",
		},
	}

	expected_query := `INSERT INTO Goal (title, start_datetime, end_date, notes, username) VALUES

	`

	query, params, err := constructGoalInsertQuery("username", &goals)

	if(
}

func TestHashPassword(t *testing.T) {
	format_str := hashPassword("password")

	//this is the expected length of the base64 string based on input key len
	b64_hash_len := 4 * params.key_len / 3
	//round up to multiple of 4
	b64_hash_len += b64_hash_len % 4

	b64_salt_len := 4 * params.salt_len / 3
	b64_salt_len += b64_salt_len % 4

	rgx := fmt.Sprintf("\\$argon2id\\$v=[0-9]+\\$m=%d,t=%d,p=%d\\$[A-Za-z0-9\\+/]{%d}\\$[A-Za-z0-9\\+/]{%d}",
		params.memory,
		params.time,
		params.threads,
		b64_salt_len,
		b64_hash_len,
	)

	match, err := regexp.Match(rgx, []byte(format_str))

	if err != nil {
		t.Errorf("regex error: %s", err.Error())
	} else if !match {
		t.Error("regex did not match")
	}
}

