package db

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/snyderks/xp-bot/logger"
	"github.com/snyderks/xp-bot/primitives"
	"github.com/snyderks/xp-bot/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// DBURI is the location of the database.
	DBURI                  = os.Getenv("XP_BOT_DB_URI")
	database               = "xp-bot"
	daysCollection         = "days"
	peopleCollection       = "people"
	nicknamesCollection    = "nicknames"
	minutesBeforeNewRecord = 5
)

// DB is a client used to interact with the database.
type DB struct {
	*mongo.Client

	People    *mongo.Collection
	Days      *mongo.Collection
	Nicknames *mongo.Collection
}

type result struct {
	date time.Time
	xp   int
}

// farEnoughInPast returns whether a time t is far enough in the past
// to create a new record in the DB instead of simply updating.
func farEnoughInPast(t time.Time) bool {
	return int(time.Now().Sub(t).Minutes()) > minutesBeforeNewRecord
}

func transactionCTX() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func (db *DB) closeDB(ctx context.Context) {
	if err := db.Client.Disconnect(ctx); err != nil {
		logger.Log.Error(err)
	}
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
	nicknames := db.Collection(nicknamesCollection)
	return DB{client, people, days, nicknames}, nil
}

// AddDay inserts a new record to track the top XP members for a day.
func AddDay(people map[string]primitives.Person) error {
	d, err := CreateDB(DBURI)
	if err != nil {
		return fmt.Errorf("Failed to initialize DB: %s", err.Error())
	}
	ctx, cancel := transactionCTX()
	defer cancel()
	defer d.closeDB(ctx)
	result, err := ReadNewestDay()

	if err != nil &&
		err.Error() != mongo.ErrNoDocuments.Error() &&
		!strings.Contains(err.Error(), mongo.ErrNoDocuments.Error()) {
		return err
	}

	// We will overwrite what we retrieved and put it back in the DB,
	// adding new records if they aren't in there and updating ones
	// that are.
	if err == nil && !farEnoughInPast(result.Date) {
		// Update everything in the map
		for i, el := range result.People {
			if val, ok := people[el.UserID]; ok {
				result.People[i].Rank = val.Rank
				result.People[i].XP = val.XP
				// We used this one, so remove it from the map.
				delete(people, el.UserID)

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
		ranks := make([]int, 0)
		for _, v := range people {
			ranks = append(ranks, v.Rank)
			result.People = append(result.People, v)
		}
		// Want to make sure we have the max and min available.
		max, min := util.MaxMinInt(ranks)

		if max != -1 && max > result.MaxRank {
			result.MaxRank = max
		}
		if min != -1 && min < result.MinRank {
			result.MinRank = min
		}

		// Sort the list of people ascending
		sort.Sort(&result.People)

		// Now we add the new record to the database.
		logger.Log.Info("Updating day with ID", result.ID)
		d.Days.UpdateOne(ctx, bson.M{"_id": result.ID}, bson.M{"$set": result})
	} else {
		peopleList := make([]primitives.Person, 0)
		ranks := make([]int, 0)
		// Hoist the people out of a map and into a list.
		for _, v := range people {
			ranks = append(ranks, v.Rank)
			peopleList = append(peopleList, v)
		}
		// Want to make sure we have the max and min available.
		max, min := util.MaxMinInt(ranks)
		newRecord := primitives.Day{
			Date:    time.Now(),
			People:  peopleList,
			MaxRank: max,
			MinRank: min,
		}

		// Sort the list of people ascending
		sort.Sort(&newRecord.People)

		logger.Log.Info("Inserting day:", newRecord)
		d.Days.InsertOne(ctx, newRecord)
	}

	return nil
}

// ReadNewestDay returns an array of the results for the given day and the date.
func ReadNewestDay() (primitives.Day, error) {
	d, err := CreateDB(DBURI)
	if err != nil {
		return primitives.Day{}, fmt.Errorf("Failed to initialize DB: %s", err.Error())
	}
	ctx, cancel := transactionCTX()
	defer cancel()
	defer d.closeDB(ctx)

	opts := options.FindOne().SetSort(bson.M{"date": -1})

	var result primitives.Day

	b := d.Days.FindOne(ctx, bson.D{}, opts)
	err = b.Decode(&result)

	if err != nil {
		return primitives.Day{}, err
	}

	return result, nil
}

// ReadPeople returns a list of XP ranges for all users requested.
// If a username cannot be found, it will be returned in the notFound
// variable. Errors from this function are not user-friendly.
// This function is designed to succeed even with partial failure.
func ReadPeople(UserIDs []string, maxDays int) (notFound []string,
	people []primitives.HistoryRange, err error) {
	d, err := CreateDB(DBURI)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to initialize DB: %s", err.Error())
	}
	ctx, cancel := transactionCTX()
	defer cancel()
	defer d.closeDB(ctx)

	notFound = make([]string, 0)

	people = make([]primitives.HistoryRange, 0)
	opts := options.Find().SetSort(bson.M{"date": 1})

	for _, UserID := range UserIDs {
		cursor, err := d.People.Find(ctx,
			bson.M{"userid": UserID,
				"date": bson.M{
					"$gt": time.Now().AddDate(0, 0, -maxDays)}},
			opts)
		if err != nil {
			return nil, nil, err
		}
		person := make([]primitives.History, 0)
		err = cursor.All(ctx, &person)
		if len(person) == 0 {
			notFound = append(notFound, UserID)
		} else {
			people = append(people, primitives.HistoryRange{History: person, UserID: UserID})
		}
	}
	return notFound, people, nil
}

func readMostRecentPerson(UserID string) (primitives.History, error) {
	d, err := CreateDB(DBURI)
	if err != nil {
		return primitives.History{}, fmt.Errorf("Failed to initialize DB: %s", err.Error())
	}
	ctx, cancel := transactionCTX()
	defer cancel()
	defer d.closeDB(ctx)

	opts := options.FindOne().SetSort(bson.M{"date": -1})
	result := primitives.History{}

	err = d.People.FindOne(ctx, bson.M{"userid": UserID}, opts).Decode(&result)
	if err != nil {
		return primitives.History{}, err
	}

	return result, nil
}

// AddPeople inserts new records into the People collection.
// Populate the Person map with the string being the username.
// Not guaranteed to be (and not currently!) represented the
// same in the database.
func AddPeople(people map[string]primitives.Person) error {
	d, err := CreateDB(DBURI)
	if err != nil {
		return fmt.Errorf("Failed to initialize DB: %s", err.Error())
	}
	ctx, cancel := transactionCTX()
	defer cancel()
	defer d.closeDB(ctx)

	t := time.Now()
	for k, v := range people {
		result, err := readMostRecentPerson(k)
		// We either couldn't get a result (which is fine)
		// or the result was more than minutesBeforeNewRecord
		// old, so we make a new one.
		if err != nil || farEnoughInPast(result.Date) {
			result = primitives.History{UserID: k, XP: v.XP, Date: t}
			logger.Log.Info("Adding person:", result)
			_, err = d.People.InsertOne(ctx, result)
			if err != nil {
				return err
			}
		} else {
			result = primitives.History{ID: result.ID, UserID: result.UserID, XP: v.XP, Date: result.Date}
			logger.Log.Info("Updating person with ID", result.ID)
			_, err = d.People.UpdateOne(ctx, bson.M{"_id": result.ID}, bson.M{"$set": result})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetNicknames retrieves a list of nicknames based on user IDs from the DB.
func GetNicknames(userIDs []string) map[string]primitives.Nickname {
	d, err := CreateDB(DBURI)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}
	ctx, cancel := transactionCTX()
	defer cancel()
	defer d.closeDB(ctx)

	m := make(map[string]primitives.Nickname)

	for _, id := range userIDs {
		nick := primitives.Nickname{}
		err := d.Nicknames.FindOne(ctx, bson.M{"userid": id}).Decode(&nick)
		if err != nil {
			m[id] = primitives.Nickname{}
		}
		m[id] = nick
	}

	return m
}

// SetNicknames takes a list of nicknames and either updates them or creates new ones.
func SetNicknames(nicknames []primitives.Nickname) {
	d, err := CreateDB(DBURI)
	if err != nil {
		logger.Log.Errorf("Failed to initialize DB: %s", err.Error())
	}
	ctx, cancel := transactionCTX()
	defer cancel()
	defer d.closeDB(ctx)

	fmt.Println(nicknames)

	for _, nick := range nicknames {
		d.Nicknames.FindOneAndUpdate(ctx, bson.M{"userid": nick.UserID},
			bson.M{"$set": nick}, options.FindOneAndUpdate().SetUpsert(true))
	}
}
