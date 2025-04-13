package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("healthy"))
}

func parseFormIntoGoals(form url.Values) (*[]Goal, error) {
	goals := make([]Goal, len(form["title"]))

	for i := range form["title"] {
		title := ""
		end_date := time.Time{}
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

		var err error = nil

		end_date_str := ""

		if len(form["due"]) > i {
			end_date_str = form["due"][i]
		}

		if end_date_str != "" {
			end_date, err = time.Parse(time.DateOnly, end_date_str)

			if err != nil {
				return nil, err
			}
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
	return (func(w http.ResponseWriter, r *http.Request) {
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
				"response_code", http.StatusInternalServerError,
			)

			http.Error(w, "error posting goals", http.StatusInternalServerError)
			return
		}

		//TODO Replace constant with username retrieved from auth
		err = CreateGoals(db, "username", goals)

		if err != nil {
			http.Error(w, "error posting goals", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("OK"))
	})
}

//handler for a route that specifies allowed methods. implements http.Handler interface
type RouteHandler struct {
	allowed_methods []string
	handler_func http.HandlerFunc
}

func (rh RouteHandler) ServeHTTP (w http.ResponseWriter, r *http.Request) {
	slog.Info(
		"received request",
		"route", r.URL,
		"method", r.Method,
	)

	if rh.allowed_methods != nil {
		if !slices.Contains(rh.allowed_methods, r.Method) {
			err_msg := "only " + strings.Join(rh.allowed_methods, ", ") + " allowed"

			slog.Error(
				err_msg,
				"response_code", http.StatusMethodNotAllowed,
			)

			http.Error(w, err_msg, http.StatusMethodNotAllowed)
			return
		}
	}

	rh.handler_func(w, r)
}

func initialiseHTTPServer(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	ping_handler := RouteHandler{
		allowed_methods: []string{"GET"},
		handler_func: handlePing,
	}

	goals_handler := RouteHandler{
		allowed_methods: []string{"POST"},
		handler_func: handleGoals(db),
	}

	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.Handle("/ping", ping_handler)
	mux.Handle("/goals", goals_handler)

	return mux
}

