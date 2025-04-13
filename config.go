package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
)

type Config struct {
	Db struct {
		Host          string `json:"host"`
		Port          uint16 `json:"port"`
		Database_name string `json:"database_name"`
		Username      string `json:"username"`
		Password      string `json:"password"`
	} `json:"db"`
}

func parseConfig(path string) (*Config, error) {
	slog.Info("attempting to open config file at " + path)
	json_str, err := os.ReadFile(path)

	if err != nil {
		slog.Error("error opening config file: " + err.Error())
		return nil, err
	}

	slog.Debug("file successfully read into str\n" + string(json_str))

	var conf Config
	err = json.Unmarshal(json_str, &conf)

	if err != nil {
		slog.Error("error unmarshalling json file: " + err.Error())
		return nil, err
	}

	slog.Info("config successfully parsed")

	return &conf, nil
}

func validateConfig(conf *Config) error {
	if conf.Db.Host == "" {
		err := errors.New("config is missing db.host value")
		slog.Error(err.Error())
		return err
	}
	if conf.Db.Port == 0 {
		err := errors.New("config is missing db.port value")
		slog.Error(err.Error())
		return err
	}
	if conf.Db.Database_name == "" {
		err := errors.New("config is missing db.database_name value")
		slog.Error(err.Error())
		return err
	}
	if conf.Db.Username == "" {
		err := errors.New("config is missing db.username value")
		slog.Error(err.Error())
		return err
	}
	if conf.Db.Password == "" {
		err := errors.New("config is missing db.password value")
		slog.Error(err.Error())
		return err
	}

	slog.Info("config validated succesfully")

	return nil
}

