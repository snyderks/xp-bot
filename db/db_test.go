package db

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var testURI = "mongodb://localhost:27017"

func testContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func TestCreateDB(t *testing.T) {
	client, err := CreateDB(testURI)
	if err != nil {
		t.Error(err.Error())
	}

	ctx, cancel := testContext()
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNewestDay(t *testing.T) {
	client, _ := CreateDB(testURI)

	time, people, err := client.ReadNewestDay()

	if err != nil {
		t.Error(err.Error())
	}
	if time.Day() != 2 {
		t.Error("Date was incorrect.")
	}
	if len(people) != 1 {
		t.Error("People were not retrieved correctly.")
	}
}

func TestGetPeople(t *testing.T) {
	people := []string{"ode", "test", "croot"}

	client, _ := CreateDB(testURI)

	notFound, results, err := client.ReadPeople(people, 10)

	if err != nil {
		t.Error(err.Error())
	}

	populatedNotFound := false
	for _, x := range notFound {
		if x == "test" {
			populatedNotFound = true
		}
	}
	if !populatedNotFound {
		t.Error("Failed to populate notFound with the correct usernames")
		return
	}
	if len(results) != 1 {
		t.Error("Incorrect number of results", results)
		return
	}
	if len(results[0].History) != 2 {
		t.Error("Not all the records were returned. Hmm")
	}
	if results[0].History[0].XP != 1 {
		t.Error("XP was wrong.")
	}
}
