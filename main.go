package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
)

func initialiseDBConn(
	host string,
	port uint16,
	database_name string,
	username string,
	password string,
) (*sql.DB, error) {
	db, err := OpenDB(host, port, database_name, username, password)

	if err != nil {
		slog.Error("could not initialise db conn: " + err.Error())
		return nil, err
	}

	slog.Info("db init success")

	err = db.Ping()

	if err != nil {
		slog.Error("error pinging db: " + err.Error())
		return nil, err
	}

	return db, nil
}

func initialiseConfig() (*Config, error) {
	conf, err := parseConfig("config.json")

	if err != nil {
		return nil, err
	}

	err = validateConfig(conf)

	if err != nil {
		return nil, err
	}

	return conf, nil
}

func main(){
	slog.SetLogLoggerLevel(slog.LevelDebug)

	conf, err := initialiseConfig()

	if err != nil {
		return
	}

	db, err := initialiseDBConn(
		conf.Db.Host,
		conf.Db.Port,
		conf.Db.Database_name,
		conf.Db.Username,
		conf.Db.Password,
	)

	if err != nil {
		return
	}

	defer db.Close()

	mux := initialiseHTTPServer(db)

	http_str := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	err = http.ListenAndServe(http_str, mux)

	if err != nil {
		slog.Error(err.Error())
	}

	return
}

