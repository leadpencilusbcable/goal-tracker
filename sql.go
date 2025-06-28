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
	start_date string
	end_date string
	completed_datetime *time.Time
	notes string
	status string
}

func GetGoals(
	db *sql.DB,
	username string,
	start_date *time.Time,
	end_date *time.Time,
) ([]Goal, error) {
	if start_date == nil {
		return nil, errors.New("start_date cannot be nil")
	}
	if end_date == nil {
		return nil, errors.New("end_date cannot be nil")
	}

	query := `SELECT title, start_date, end_date, completed_datetime, notes
	FROM Goal WHERE username = $1 AND (end_date BETWEEN $2 AND $3)`

	slog.Info(
		"executing db query",
		"query", query,
	)

	rows, err := db.Query(
		query,
		username,
		start_date,
		end_date,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var goals []Goal

	for rows.Next() {
		var goal Goal

		err = rows.Scan(
			&goal.title,
			&goal.start_date,
			&goal.end_date,
			&goal.completed_datetime,
			&goal.notes,
		)

		//because postgres returns DATE types as a datetime string,
		//manually getting just the date string so that it can be
		//parsed into user's timezone as just date
		goal.start_date = goal.end_date[:10]
		goal.end_date = goal.end_date[:10]

		if err != nil {
			return nil, err
		}

		goals = append(goals, goal)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return goals, nil
}

type GoalInsert struct {
	title string
	start_date *time.Time
	end_date *time.Time
	notes string
}

func InsertGoals(db *sql.DB, username string, goals *[]GoalInsert) error {
	query, params, err := constructGoalInsertQuery(username, goals)

	if err != nil {
		slog.Error(
			"error constructing goal insert query",
			"err", err.Error(),
		)

		return err
	}

	slog.Info(
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
func constructGoalInsertQuery(username string, goals *[]GoalInsert) (string, *[]any, error) {
	if goals == nil || len(*goals) == 0 {
		return "", nil, errors.New("no goals provided to construct query")
	}

	var query strings.Builder
	params := []any{}

	query.WriteString("INSERT INTO Goal (title, start_date, end_date, notes, username) VALUES ")

	for i, goal := range *goals {
		value_str := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
			i * 5 + 1,
			i * 5 + 2,
			i * 5 + 3,
			i * 5 + 4,
			i * 5 + 5,
		)

		query.WriteString(value_str)

		params = append(params, goal.title, goal.start_date, goal.end_date, goal.notes, username)

		if i != len(*goals) - 1 {
			query.WriteString(", ")
		}
	}

	return query.String(), &params, nil
}

func DeleteSessionId(db *sql.DB, session_id_sha256 [32]byte) error {
	query := "DELETE FROM SessionId WHERE session_id_sha256=$1"

	slog.Info(
		"executing db query",
		"query", query,
	)

	_, err := db.Exec(query, session_id_sha256[:])

	if err != nil {
		slog.Error(
			"error deleting session id from db",
			"err", err.Error(),
		)
	}

	return err
}

func UpsertSessionId(db *sql.DB, username string, session_id_sha256 [32]byte) error {
	if username == "" {
		return errors.New("empty username when attempting to insert auth token")
	}

	query := `
	INSERT INTO SessionId (username, session_id_sha256)
	VALUES ($1, $2)
	ON CONFLICT (username)
	DO UPDATE SET session_id_sha256 = $2
	`

	slog.Info(
		"executing db query",
		"query", query,
	)

	_, err := db.Exec(query, username, session_id_sha256[:])

	if err != nil {
		slog.Error(
			"error inserting session id into db",
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

