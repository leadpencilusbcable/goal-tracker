package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	_ "github.com/lib/pq"
)

type argon2_params struct {
	time     uint32
	memory   uint32
	threads  uint8
	key_len  uint32
	salt_len uint16
}

var params = argon2_params{
	time: 1,
	memory: 64 * 1024,
	threads: 4,
	key_len: 16,
	salt_len: 16,
}

func OpenDB(host string, port uint16, db_name string, username string, password string) (*sql.DB, error) {
	conn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		host,
		port,
		db_name,
		username,
		password,
	)

	db, err := sql.Open("postgres", conn)

	return db, err
}

func CreateUser(db *sql.DB, username string, password string) error {
	query := `
	INSERT INTO User_ (username, password_params)
	VALUES ($1, $2)
	`

	password_params := hashPassword(password)

	_, err := db.Exec(query, username, password_params)

	if err != nil {
		slog.Error(
			"error inserting user into db",
			"err", err.Error(),
		)
	}

	return err
}

type Goal struct {
	title string
	end_date time.Time
	notes string
}

func CreateGoals(db *sql.DB, username string, goals *[]Goal) error {
	query, params, err := constructGoalInsertQuery(username, goals)

	if err != nil {
		slog.Error(
			"error constructing goal insert query",
			"err", err.Error(),
		)

		return err
	}

	slog.Debug(
		"executing db query",
		"query", query,
	)

	_, err = db.Exec(query, *params...)

	if err != nil {
		slog.Error(
			"error posting goals to db",
			"err", err.Error(),
		)
	}

	return err
}

//returns the query string and the params
func constructGoalInsertQuery(username string, goals *[]Goal) (string, *[]any, error) {
	if len(*goals) == 0 {
		return "", nil, errors.New("no goals provided to construct query")
	}

	var query strings.Builder
	params := []any{}

	query.WriteString("INSERT INTO Goal (title, start_datetime, end_date, notes, username) VALUES ")

	for i, goal := range *goals {
		value_str := fmt.Sprintf("($%d, NOW(), $%d, $%d, $%d)",
			i * 4 + 1,
			i * 4 + 2,
			i * 4 + 3,
			i * 4 + 4,
		)

		query.WriteString(value_str)

		params = append(params, goal.title, goal.end_date, goal.notes, username)

		if i != len(*goals) - 1 {
			query.WriteString(", ")
		}
	}

	return query.String(), &params, nil
}

//Hashes a password with argon2 and returns a string
//containing argon2 params for hashing the password
func hashPassword(password string) string {
	salt := generateSalt(params.salt_len)
	hash := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, params.key_len)

	b64_salt := base64.RawStdEncoding.EncodeToString(salt)
	b64_hash := base64.RawStdEncoding.EncodeToString(hash)

	format_str := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.memory,
		params.time,
		params.threads,
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

