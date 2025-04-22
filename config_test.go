package main

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestNoConfig(t *testing.T) {
	_, err := os.Stat("./config.json")

	if !os.IsNotExist(err) {
		//rename file temporarily to test behaviour without file
		os.Rename("./config.json", "./configzdxzdzddd.jsq")
		defer os.Rename("./configzdxzdzddd.jsq", "./config.json")
	}

	_, err = parseConfig("./config.json")

	if !errors.Is(err, os.ErrNotExist) {
		t.Error("parseConfig did not return non existent error")
	}
}

func TestMalformedConfig(t *testing.T) {
	_, err := os.Stat("./config.json")

	if !os.IsNotExist(err) {
		//rename file temporarily to test behaviour without file
		os.Rename("./config.json", "./configzdxzdzddd.jsq")
		defer os.Rename("./configzdxzdzddd.jsq", "./config.json")
	}

	file, err := os.Create("./config.json")

	if err != nil {
		t.Error("error creating config file for test")
	}

	defer file.Close()

	_, err = file.Write([]byte("abcdef"))

	if err != nil {
		t.Error("error writing to config file for test")
	}

	_, err = parseConfig("./config.json")

	if _, ok := err.(*json.UnmarshalTypeError); ok {
		t.Error("parseConfig did not return json unmarshal error")
	}
}

func TestValidateIncompleteConfig(t *testing.T) {
	_, err := os.Stat("./config.json")

	if !os.IsNotExist(err) {
		//rename file temporarily to test behaviour without file
		os.Rename("./config.json", "./configzdxzdzddd.jsq")
		defer os.Rename("./configzdxzdzddd.jsq", "./config.json")
	}

	file, err := os.Create("./config.json")

	if err != nil {
		t.Error("error creating config file for test")
	}

	defer file.Close()

	_, err = file.Write([]byte("{\"db\": { \"host\": null }}"))

	if err != nil {
		t.Error("error writing to config file for test")
	}

	conf, err := parseConfig("./config.json")

	if err != nil {
		t.Error("parseConfig should not fail")
	}

	err = validateConfig(conf)

	if err == nil {
		t.Error("validate config should fail with missing properties")
	}
}

func TestValidateCompleteConfig(t *testing.T) {
	_, err := os.Stat("./config.json")

	if !os.IsNotExist(err) {
		//rename file temporarily to test behaviour without file
		os.Rename("./config.json", "./configzdxzdzddd.jsq")
		defer os.Rename("./configzdxzdzddd.jsq", "./config.json")
	}

	file, err := os.Create("./config.json")

	if err != nil {
		t.Error("error creating config file for test")
	}

	defer file.Close()

	_, err = file.Write([]byte("{ \"port\": 1800, \"db\": { \"host\": \"localhost\", \"port\": 5432, \"database_name\": \"goal\", \"username\": \"username\", \"password\": \"username\" } }"))

	if err != nil {
		t.Error("error writing to config file for test")
	}

	conf, err := parseConfig("./config.json")

	if err != nil {
		t.Error("parseConfig should not fail")
	}

	err = validateConfig(conf)

	if err != nil {
		t.Error("validate config should not fail with valid properties")
	}
}

