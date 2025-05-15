package main

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestConstructGoalInsertQuery(t *testing.T) {
	now := time.Now()

	goals := []GoalInsert{
		{
			title: "title",
			start_date: &now,
			end_date: &now,
			notes: "",
		},
		{
			title: "title",
			start_date: &now,
			end_date: nil,
			notes: "notes",
		},
	}

	expected_query := `INSERT INTO Goal (title, start_date, end_date, notes, username) VALUES ($1, $2, $3, $4, $5), ($6, $7, $8, $9, $10)`

	var nil_time *time.Time = nil

	expected_params := []any{
		"title",
		&now,
		&now,
		"",
		"username",
		"title",
		&now,
		nil_time,
		"notes",
		"username",
	}

	query, params, err := constructGoalInsertQuery("username", &goals)

	if(err != nil) {
		t.Errorf("query error: %s", err.Error())
	}

	if query != expected_query {
		t.Errorf("query was not as expected. expected: %s, got %s", expected_query, query)
	}

	if len(expected_params) != len(*params) {
		t.Errorf("more params than expected. expected: %d, got: %d", len(expected_params), len(*params))
	}

	for i, param := range *params {
		if param != expected_params[i] {
			t.Errorf("difference in params at index %d. expected: %v, got %v", i, param, expected_params[i])
		}
	}
}

func TestHashPassword(t *testing.T) {
	format_str := hashPassword("password")

	//this is the expected length of the base64 string based on input key len
	b64_hash_len := 4 * default_params.key_len / 3
	//round up to multiple of 4
	b64_hash_len += b64_hash_len % 4

	b64_salt_len := 4 * default_params.salt_len / 3
	b64_salt_len += b64_salt_len % 4

	rgx := fmt.Sprintf("\\$argon2id\\$v=[0-9]+\\$m=%d,t=%d,p=%d\\$[A-Za-z0-9\\+/]{%d}\\$[A-Za-z0-9\\+/]{%d}",
		default_params.memory,
		default_params.time,
		default_params.threads,
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

