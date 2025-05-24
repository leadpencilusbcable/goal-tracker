package main

import (
	"testing"
	"time"
)

func TestGoalStatus(t *testing.T) {
	now := time.Now()
	now_date_str := now.Format(time.DateOnly)

	//set now to start of day so it is not ahead of end_date when parsed
	now, _ = time.Parse(time.DateOnly, now_date_str)

	goal := Goal{ end_date: now_date_str }
	status, err := getGoalStatus(goal, &now)

	if err != nil {
		t.Errorf("error determining goal status. %s", err.Error())
	} else if status != "In progress" {
		t.Errorf("expected status In progress. Got: %s", status)
	}

	yesterday := now.Add(-time.Hour * 24)
	goal.end_date = yesterday.Format(time.DateOnly)
	status, err = getGoalStatus(goal, &now)

	if err != nil {
		t.Errorf("error determining goal status. %s", err.Error())
	} else if status != "Failed" {
		t.Errorf("expected status Failed. Got: %s", status)
	}

	goal.completed_datetime = &now
	status, err = getGoalStatus(goal, &yesterday)

	if status != "Complete" {
		t.Errorf("expected status Complete. Got: %s", status)
	}
}

