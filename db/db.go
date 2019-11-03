package db

import (
	"context"
	"strings"
	"time"

	"github.com/snyderks/xp-bot/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var database = "xp-bot"
var daysCollection = "days"
var peopleCollection = "people"
var minutesBeforeNewRecord = 60

// DB is a client used to interact with the database.
type DB struct {
	*mongo.Client

	People *mongo.Collection
	Days   *mongo.Collection
}

// History is a record of a user's XP at a given moment.
type History struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	UName string             `bson:"name"`
	XP    int                `bson:"xp"`
	Date  time.Time          `bson:"date"`
}

// Day is the structure of the documents in the Days collection.
type Day struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Date    time.Time          `bson:"date"`
	People  []Person           `bson:"people"`
	MaxRank int                `bson:"maxRank"`
	MinRank int                `bson:"minRank"`
}

// Person is the XP for a given uname.
type Person struct {
	UName string `bson:"name"`
	XP    int    `bson:"xp"`
	Rank  int    `bson:"rank"`
}

// HistoryRange is a list of moments of a user's XP.
type HistoryRange struct {
	History []History
	UName   string
}

type result struct {
	date time.Time
	xp   int
}

// farEnoughInPast returns whether a time t is far enough in the past
// to create a new record in the DB instead of simply updating.
func farEnoughInPast(t time.Time) bool {
	return int(time.Now().Sub(t).Minutes()) < minutesBeforeNewRecord
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

// AddDay inserts a new record to track the top XP members for a day.
func (db *DB) AddDay(people map[string]Person) error {
	result, err := db.ReadNewestDay()

	if err != nil &&
		err.Error() != mongo.ErrNoDocuments.Error() &&
		!strings.Contains(err.Error(), mongo.ErrNoDocuments.Error()) {
		return err
	}

	// We'll use this later. Declaring it here to avoid duplication.
	ctx, cancel := transactionCTX()
	defer cancel()

	// We will overwrite what we retrieved and put it back in the DB,
	// adding new records if they aren't in there and updating ones
	// that are.
	if err == nil && !farEnoughInPast(result.Date) {
		// Update everything in the map
		for i, el := range result.People {
			if val, ok := people[el.UName]; ok {
				result.People[i].Rank = val.Rank
				result.People[i].XP = val.XP
				// We used this one, so remove it from the map.
				delete(people, el.UName)

				// Determine if the min/max ranks need to be updated.
				if val.Rank < result.MinRank {
					result.MinRank = val.Rank
				}
				if val.Rank > result.MaxRank {
					result.MaxRank = val.Rank
				}
			}
		}

		// All that is left in the map is new records for this document.
		// Append them.
		for _, v := range people {
			result.People = append(result.People, v)
		}

		// Now we add the new record to the database.
		db.Days.UpdateOne(ctx, bson.M{"_id": result.ID}, bson.M{"$set": result})
	} else {
		peopleList := make([]Person, 0)
		ranks := make([]int, 0)
		// Hoist the people out of a map and into a list.
		for _, v := range people {
			ranks = append(ranks, v.Rank)
			peopleList = append(peopleList, v)
		}
		// Want to make sure we have the max and min available.
		max, min := util.MaxMin(ranks)
		newRecord := Day{
			Date:    time.Now(),
			People:  peopleList,
			MaxRank: max,
			MinRank: min,
		}
		db.Days.InsertOne(ctx, newRecord)
	}

	return nil
}

// ReadNewestDay returns an array of the results for the given day and the date.
func (db *DB) ReadNewestDay() (Day, error) {
	ctx, cancel := transactionCTX()
	defer cancel()

	opts := options.FindOne().SetSort(bson.M{"date": -1})

	var result Day

	b := db.Days.FindOne(ctx, bson.D{}, opts)
	err := b.Decode(&result)

	if err != nil {
		return Day{}, err
	}

	return result, nil
}

// ReadPeople returns a list of XP ranges for all users requested.
// If a username cannot be found, it will be returned in the notFound
// variable. Errors from this function are not user-friendly.
// This function is designed to succeed even with partial failure.
func (db *DB) ReadPeople(unames []string, maxDays int) (notFound []string,
	people []HistoryRange, err error) {
	ctx, cancel := transactionCTX()
	defer cancel()

	notFound = make([]string, 0)

	people = make([]HistoryRange, 0)
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
		person := make([]History, 0)
		err = cursor.All(ctx, &person)
		if len(person) == 0 {
			notFound = append(notFound, uname)
		} else {
			people = append(people, HistoryRange{History: person, UName: uname})
		}
	}
	return notFound, people, nil
}

func (db *DB) readMostRecentPerson(uname string) (History, error) {
	ctx, cancel := transactionCTX()
	defer cancel()

	opts := options.FindOne().SetSort(bson.M{"date": -1})
	result := History{}

	err := db.People.FindOne(ctx, bson.M{"name": uname}, opts).Decode(&result)
	if err != nil {
		return History{}, err
	}

	return result, nil
}

// AddPeople inserts new records into the People collection.
// Populate the Person map with the string being the username.
// Not guaranteed to be (and not currently!) represented the
// same in the database.
func (db *DB) AddPeople(people map[string]Person) error {
	ctx, cancel := transactionCTX()
	defer cancel()
	for k, v := range people {
		result, err := db.readMostRecentPerson(k)
		// We either couldn't get a result (which is fine)
		// or the result was more than minutesBeforeNewRecord
		// old, so we make a new one.
		if err != nil || farEnoughInPast(result.Date) {
			result = History{UName: k, XP: v.XP, Date: time.Now()}
			_, err = db.People.InsertOne(ctx, result)
			if err != nil {
				return err
			}
		} else {
			result = History{ID: result.ID, UName: result.UName, XP: v.XP, Date: result.Date}
			_, err = db.People.UpdateOne(ctx, bson.M{"_id": result.ID}, result)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
