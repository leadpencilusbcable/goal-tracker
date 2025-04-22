package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("healthy"))
}

func parseFormIntoGoals(form url.Values) (*[]Goal, error) {
	if len(form["title"]) == 0 {
		return nil, errors.New("no goals")
	}

	goals := make([]Goal, len(form["title"]))

	for i := range form["title"] {
		title := ""
		var end_date *time.Time = nil
		notes := ""

		if len(form["title"]) > i {
			title = form["title"][i]
		}

		if title == "" {
			return nil, errors.New("goal does not have a title")
		}

		if len(form["notes"]) > i {
			notes = form["notes"][i]
		}

		end_date_str := ""

		if len(form["due"]) > i {
			end_date_str = form["due"][i]
		}

		if end_date_str != "" {
			end, err := time.Parse(time.DateOnly, end_date_str)

			if err != nil {
				return nil, err
			}

			end_date = &end
		}

		goals[i] = Goal{
			title,
			end_date,
			notes,
		}
	}

	return &goals, nil
}

func handleGoals(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()

		if err != nil {
			err_msg := "malformed form request"

			slog.Error(
				err_msg,
				"err", err.Error(),
				"response_code", http.StatusBadRequest,
			)

			http.Error(w, err_msg, http.StatusBadRequest)
		}

		goals, err := parseFormIntoGoals(r.PostForm)

		if err != nil {
			slog.Error(
				"error parsing form into goals",
				"err", err.Error(),
				"response_code", http.StatusBadRequest,
			)

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//TODO Replace constant with username retrieved from auth
		err = CreateGoals(db, "username", goals)

		if err != nil {
			http.Error(w, "error posting goals", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("OK"))
	}
}

func handleLoginGet(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("public/login.html")

	if err != nil {
		slog.Error(
			"error reading public/login.html file",
			"err", err.Error(),
			"response_code", http.StatusInternalServerError,
		)

		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	w.Write(content)
	return
}

func parseFormIntoUser(form url.Values) (*User, error) {
	//there should only be a single username or password
	//just catering for the nature of forms
	usernames := form["username"]
	passwords := form["password"]

	if len(usernames) == 0 {
		return nil, errors.New("no username on form")
	}

	if len(passwords) == 0 {
		return nil, errors.New("no password on form")
	}

	username := usernames[0]

	if username == "" {
		return nil, errors.New("empty username on form")
	}

	password := passwords[0]

	if password == "" {
		return nil, errors.New("empty password on form")
	}

	user := User{
		username: username,
		password: password,
	}

	return &user, nil
}

func handleLoginPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()

		if err != nil {
			err_msg := "malformed form request"

			slog.Error(
				err_msg,
				"err", err.Error(),
				"response_code", http.StatusBadRequest,
			)

			http.Error(w, err_msg, http.StatusBadRequest)
			return
		}

		user, err := parseFormIntoUser(r.PostForm)

		if err != nil {
			slog.Error(
				"error parsing form into user",
				"err", err.Error(),
				"response_code", http.StatusInternalServerError,
			)

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		db_user, err := GetUser(db, user.username)

		if errors.Is(err, sql.ErrNoRows) {
			slog.Debug(
				"username mismatch",
				"username", user.username,
				"err", err.Error(),
				"response_code", http.StatusUnauthorized,
			)

			http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, "error validating user", http.StatusInternalServerError)
			return
		}

		match, err := comparePasswordWithHash(user.password, db_user.password)

		if err != nil {
			slog.Error(
				"error comparing passwords",
				"err", err.Error(),
				"response_code", http.StatusInternalServerError,
			)

			http.Error(w, "error validating user", http.StatusInternalServerError)
			return
		}

		if !match {
			slog.Debug(
				"password mismatch",
				"username", user.username,
				"response_code", http.StatusUnauthorized,
			)

			http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}

		w.Write([]byte("OK"))
	}
}

func validateUserConstraints(user *User) string {
	err_str := ""

	//TODO figure out how this handles utf8 longer chars
	if len(user.password) < 8 {
		err_str = "Password must be 8 characters or longer"
	}

	if err_str != "" {
		slog.Debug(
			"user violated constraints",
			"username", user.username,
			"err", err_str,
		)
	}

	return err_str
}

func handleRegisterPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()

		if err != nil {
			err_msg := "malformed form request"

			slog.Error(
				err_msg,
				"err", err.Error(),
				"response_code", http.StatusBadRequest,
			)

			http.Error(w, err_msg, http.StatusBadRequest)
			return
		}

		user, err := parseFormIntoUser(r.PostForm)

		if err != nil {
			slog.Error(
				"error parsing form into user",
				"err", err.Error(),
				"response_code", http.StatusInternalServerError,
			)

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		valid_user := validateUserConstraints(user)

		if valid_user != "" {
			http.Error(w, valid_user, http.StatusUnprocessableEntity)
			return
		}

		pg_err := CreateUser(db, user)

		if pg_err != nil {
			if pg_err.pg_err.Code == "23505" {
				http.Error(w, "Username already exists", http.StatusConflict)
			} else {
				http.Error(w, "error creating user", http.StatusInternalServerError)
			}

			return
		}

		slog.Info(
			"successfully created user",
			"user", user.username,
			"response_code", http.StatusCreated,
		)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("OK"))
	}
}

func handleRegisterGet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile("public/register.html")

		if err != nil {
			slog.Error(
				"error reading public/register.html file",
				"err", err.Error(),
				"response_code", http.StatusInternalServerError,
			)

			http.Error(w, "unknown error", http.StatusInternalServerError)
			return
		}

		w.Write(content)
	}
}

func initialiseHTTPServer(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /", http.FileServer(http.Dir("public")))
	mux.HandleFunc("GET /ping", handlePing)
	mux.HandleFunc("GET /goals", handleGoals(db))
	mux.HandleFunc("GET /login", handleLoginGet)
	mux.HandleFunc("POST /login", handleLoginPost(db))
	mux.HandleFunc("GET /register", handleRegisterGet(db))
	mux.HandleFunc("POST /register", handleRegisterPost(db))

	return mux
}

