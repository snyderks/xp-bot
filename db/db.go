package db

import (
	"context"
	"time"

	"github.com/snyderks/xp-bot/bot"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var database = "xp-bot"
var daysCollection = "days"
var peopleCollection = "people"

// DB is a client used to interact with the database.
type DB struct {
	*mongo.Client

	People *mongo.Collection
	Days   *mongo.Collection
}

// day is the structure of the documents in the Days collection.
type day struct {
	Date   time.Time
	People []bot.Person
}

type result struct {
	date time.Time
	xp   int
}

func transactionCTX() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// CreateDB returns a new instance of a DB client, given a URI to connect to.
func CreateDB(URI string) (DB, error) {
	ctx, cancel := transactionCTX()
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URI))
	if err != nil {
		return DB{}, err
	}
	db := client.Database(database)
	people := db.Collection(peopleCollection)
	days := db.Collection(daysCollection)
	return DB{client, people, days}, nil
}

// // AddDay inserts a new record to track the top XP members for a day.
// func (db *DB) AddDay() {
// 	ctx, cancel := transactionCTX()
// 	defer cancel()

// 	db.Days.
// }

// ReadNewestDay returns an array of the results for the given day and the date.
func (db *DB) ReadNewestDay() (time.Time, []bot.Person, error) {
	ctx, cancel := transactionCTX()
	defer cancel()

	opts := options.FindOne().SetSort(bson.M{"date": -1})

	var result day

	b := db.Days.FindOne(ctx, bson.D{}, opts)
	err := b.Decode(&result)

	if err != nil {
		return time.Now(), nil, err
	}

	return result.Date, result.People, nil
}

// ReadPeople returns a list of XP ranges for all users requested.
// If a username cannot be found, it will be returned in the notFound
// variable. Errors from this function are not user-friendly.
// This function is designed to succeed even with partial failure.
func (db *DB) ReadPeople(unames []string, maxDays int) (notFound []string,
	people []bot.HistoryRange, err error) {
	ctx, cancel := transactionCTX()
	defer cancel()

	notFound = make([]string, 0)

	people = make([]bot.HistoryRange, 0)
	opts := options.Find().SetSort(bson.M{"date": -1})

	for _, uname := range unames {
		cursor, err := db.People.Find(ctx,
			bson.M{"person": uname,
				"date": bson.M{
					"$gt": time.Now().AddDate(0, 0, -maxDays)}},
			opts)
		if err != nil {
			return nil, nil, err
		}
		person := make([]bot.History, 0)
		err = cursor.All(ctx, &person)
		if len(person) == 0 {
			notFound = append(notFound, uname)
		} else {
			people = append(people, bot.HistoryRange{History: person, UName: uname})
		}
	}
	return notFound, people, nil
}
