package main

import (
	"testing"
	"time"
)

func TestGoalStatus(t *testing.T) {
	loc, err := time.LoadLocation("UTC")

	if err != nil {
		t.Errorf("error parsing UTC location. %s", err.Error())
	}

	now := time.Now().In(loc)
	now_date_str := now.Format(time.DateOnly)

	goal := Goal{ end_date: now_date_str }
	status, err := getGoalStatus(goal, &now, loc)

	if err != nil {
		t.Errorf("error determining goal status. %s", err.Error())
	} else if status != "In progress" {
		t.Errorf("expected status In progress. Got:  %s", status)
	}

	yesterday := now.Add(-time.Hour * 24)
	goal.end_date = yesterday.Format(time.DateOnly)
	status, err = getGoalStatus(goal, &now, loc)

	if err != nil {
		t.Errorf("error determining goal status. %s", err.Error())
	} else if status != "Failed" {
		t.Errorf("expected status Failed. Got:  %s", status)
	}

	goal.completed_datetime = &now
	status, err = getGoalStatus(goal, &yesterday, loc)

	if status != "Complete" {
		t.Errorf("expected status Complete. Got:  %s", status)
	}
}

