package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

var templates *template.Template

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("healthy"))
}

func parseFormIntoGoals(form url.Values) (*[]GoalInsert, error) {
	if len(form["title"]) == 0 {
		return nil, errors.New("no goals")
	}

	goals := make([]GoalInsert, len(form["title"]))

	for i := range form["title"] {
		title := ""
		var end_date *time.Time = nil
		var start_date *time.Time = nil
		notes := ""

		if len(form["title"]) > i {
			title = form["title"][i]
		}

		if title == "" {
			return nil, errors.New("goal does not have a title")
		}

		start_date_str := ""

		if len(form["start"]) > i {
			start_date_str = form["start"][i]
		}

		if start_date_str != "" {
			start, err := time.Parse(time.DateOnly, start_date_str)

			if err != nil {
				return nil, err
			}

			start_date = &start
		} else {
			return nil, errors.New("goal does not have a start")
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

		goals[i] = GoalInsert{
			title: title,
			start_date: start_date,
			end_date: end_date,
			notes: notes,
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

		username := r.Context().Value("username").(string)

		err = InsertGoals(db, username, goals)

		if err != nil {
			http.Error(w, "error posting goals", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("OK"))
	}
}

func generateLoginUserTemplate(username string) (*bytes.Buffer, error) {
	user := struct{ Username string }{ Username: username }

	buf := bytes.Buffer{}
	err := templates.ExecuteTemplate(&buf, "login.html", user)

	if err != nil {
		slog.Error(
			"error executing login template",
			"err", err.Error(),
			"response_code", http.StatusInternalServerError,
		)

		return nil, err
	}

	return &buf, nil
}

func handleLoginGet(w http.ResponseWriter, r *http.Request) {
	if username, ok := r.URL.Query()["username"]; ok {
		buf, err := generateLoginUserTemplate(username[0])

		if err != nil {
			http.Error(w, "unknown error", http.StatusInternalServerError)
			return
		}

		buf.WriteTo(w)
		return
	}

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

func validateUserAgainstDB(db *sql.DB, user *User) (err_msg string, status_code int) {
  db_user, err := GetUser(db, user.username)

  if errors.Is(err, sql.ErrNoRows) {
    slog.Debug(
      "username mismatch",
      "username", user.username,
      "err", err.Error(),
      "response_code", http.StatusUnauthorized,
    )

    return "Incorrect username or password", http.StatusUnauthorized
  } else if err != nil {
    return "Error validating user", http.StatusInternalServerError
  }

  match, err := comparePasswordWithHash(user.password, db_user.password)

  if err != nil {
    slog.Error(
      "error comparing passwords",
      "err", err.Error(),
      "response_code", http.StatusInternalServerError,
    )

    return "Error validating user", http.StatusInternalServerError
  }

  if !match {
    slog.Debug(
      "password mismatch",
      "username", user.username,
      "response_code", http.StatusUnauthorized,
    )

    return "Incorrect username or password", http.StatusUnauthorized
  }

  return "", 0
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

      err_str, status_code := validateUserAgainstDB(db, user)

      if status_code != 0 {
			http.Error(w, err_str, status_code)
			return
      }

		session_id, err := CreateUserSessionId(db, user.username)

		if err != nil {
			slog.Error(
				"error generating session id",
				"username", user.username,
				"response_code", http.StatusInternalServerError,
			)

			http.Error(w, "Error validating user", http.StatusInternalServerError)
			return
		}

		slog.Info(
			"successfully logged in user",
			"username", user.username,
			"response_code", http.StatusOK,
		)

		cookie_str := fmt.Sprintf("session_id=%s", session_id)

		w.Header().Add("Set-Cookie", cookie_str)
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

		pg_err := InsertUser(db, user)

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

func handleRegisterGet(w http.ResponseWriter, r *http.Request) {
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

type GoalDisplay struct {
	Title string
	Status string
	Notes string
	StartDate string
	DueDate string
}

func goalsToDisplayGoals(goals []Goal) []GoalDisplay {
	goal_display := make([]GoalDisplay, len(goals))

	for i, goal := range goals {
		goal_display[i] = GoalDisplay{
			Title: goal.title,
			Status: "",
			Notes: goal.notes,
			StartDate: goal.start_date.Format("2006-01-02"),
			DueDate: goal.end_date.Format("2006-01-02"),
		}
	}

	return goal_display
}

type GoalDisplayTemplate struct {
	StartDate string
	EndDate string
	GoalDisplay []GoalDisplay
}

func handleHomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

func authorisationMiddleware(next http.Handler, db *sql.DB) http.Handler {
	handler_func := func(w http.ResponseWriter, r *http.Request) {
		session_id, err := r.Cookie("session_id")

		if err != nil {
			slog.Info("session_id cookie not provided or in incorrect format", "response_code", http.StatusSeeOther)
			w.Header().Add("Location", "/login")
			http.Error(w, "session_id cookie not provided", http.StatusSeeOther)
			return
		}

		username, err := VerifyUser(db, session_id.Value)

		if err != nil {
			slog.Error(
				"error verifying user session id",
				"err", err.Error(),
				"response_code", http.StatusInternalServerError,
			)
		}

		if username == "" {
			slog.Info("session_id cookie doesn't exist or has expired", "response_code", http.StatusSeeOther)
			w.Header().Add("Location", "/login")
			http.Error(w, "session_id cookie doesn't exist or has expired", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "username", username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler_func)
}

func initialiseTemplates() *template.Template {
	templates, err := template.ParseFiles("templates/login.html")

	if err != nil {
		slog.Error(
			"error parsing templates/login.html",
			"err", err.Error(),
		)

		return nil
	}

	_, err = templates.ParseFiles("templates/components/goal-table.html")

	if err != nil {
		slog.Error(
			"error parsing templates/components/goal-table.html",
			"err", err.Error(),
		)

		return nil
	}

	return templates
}

func handleGoalsGet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)

		start := time.Now()
		end := time.Now().Add(time.Hour * 24 * 7)

		goals, err := GetGoals(
			db,
			username,
			&start,
			&end,
			nil,
		)

		if err != nil {
			slog.Error(
				"error retrieving goals",
				"err", err.Error(),
				"response_code", http.StatusInternalServerError,
			)

			http.Error(w, "error retrieving goals", http.StatusInternalServerError)
			return
		}

		slog.Debug(
			"goals for user",
			"goals", goals,
			"username", username,
		)

		if len(goals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		goal_display := goalsToDisplayGoals(goals)

		template_data := GoalDisplayTemplate{
			StartDate: start.Format("2006-01-02"),
			EndDate: end.Format("2006-01-02"),
			GoalDisplay: goal_display,
		}

		buf := bytes.Buffer{}
		err = templates.ExecuteTemplate(&buf, "goal-table.html", template_data)

		if err != nil {
			http.Error(w, "unknown error", http.StatusInternalServerError)
			return
		}

		buf.WriteTo(w)
	}
}

func initialiseHTTPServer(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	templates = initialiseTemplates()

	if templates == nil {
		return nil
	}

	home_handler := authorisationMiddleware(http.HandlerFunc(handleHomePage), db)
	goals_post_handler := authorisationMiddleware(handleGoals(db), db)
	goals_get_handler := authorisationMiddleware(handleGoalsGet(db), db)

	mux.Handle("GET /", http.FileServer(http.Dir("./public")))
	mux.Handle("GET /{$}", home_handler)
	mux.Handle("GET /goals", goals_get_handler)
	mux.Handle("POST /goals", goals_post_handler)
	mux.HandleFunc("GET /ping", handlePing)
	mux.HandleFunc("GET /login", handleLoginGet)
	mux.HandleFunc("POST /login", handleLoginPost(db))
	mux.HandleFunc("GET /register", handleRegisterGet)
	mux.HandleFunc("POST /register", handleRegisterPost(db))

	return mux
}

