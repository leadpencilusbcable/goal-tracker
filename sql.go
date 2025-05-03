package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

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

type User struct {
	username string
	password string
}

type pgErr struct {
	err error
	pg_err *pq.Error
}

func InsertUser(db *sql.DB, user *User) *pgErr {
	query := `
	INSERT INTO User_ (username, password_params)
	VALUES ($1, $2)
	`

	password_params := hashPassword(user.password)

	_, err := db.Exec(query, user.username, password_params)

	if err != nil {
		slog.Error(
			"error inserting user into db",
			"username", user.username,
			"err", err.Error(),
		)

		pg_err := err.(*pq.Error)

		return &pgErr{
			err: err,
			pg_err: pg_err,
		}
	}

	return nil
}

func GetUser(db *sql.DB, username string) (*User, error) {
	var user User

	row := db.QueryRow(
		"SELECT username, password_params FROM User_ WHERE username = $1",
		username,
	)

	err := row.Scan(&user.username, &user.password)

	if err != nil {
		slog.Error(
			"error retrieving user from db",
			"username", username,
			"err", err.Error(),
		)

		return nil, err
	}

	return &user, nil
}

type Goal struct {
	title string
	end_date *time.Time
	notes string
}

func InsertGoals(db *sql.DB, username string, goals *[]Goal) error {
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

func UpsertSessionId(db *sql.DB, username string, session_id_sha256 [32]byte) error {
	if username == "" {
		return errors.New("empty username when attempting to insert auth token")
	} else if len(session_id_sha256) != 32 {
		return errors.New("invalid token_sha256 []byte when attempting to insert auth token")
	}

	query := `
	INSERT INTO SessionId (username, session_id_sha256)
	VALUES ($1, $2)
	ON CONFLICT (username)
	DO UPDATE SET session_id_sha256 = $2
	`

	slog.Debug(
		"executing db query",
		"query", query,
	)

	_, err := db.Exec(query, username, session_id_sha256[:])

	if err != nil {
		slog.Error(
			"error inserting auth_token into db",
			"username", username,
			"err", err.Error(),
		)
	}

	return err
}

func GetSessionId(db *sql.DB, session_id_sha256 [32]byte) (string, error) {
	row := db.QueryRow(
		"SELECT username FROM SessionId WHERE session_id_sha256 = $1",
		session_id_sha256[:],
	)

	var username string
	err := row.Scan(&username)

	if err != nil {
		slog.Error(
			"error retrieving session id from db",
			"err", err.Error(),
		)

		return "", err
	}

	return username, nil
}

