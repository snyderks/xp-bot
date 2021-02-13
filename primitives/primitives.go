package primitives

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// People is an array of Person
type People []Person

func (x People) Len() int           { return len(x) }
func (x People) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x People) Less(i, j int) bool { return x[i].Rank < x[j].Rank }

// History is a record of a user's XP at a given moment.
type History struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID string             `bson:"userid"`
	XP     int                `bson:"xp"`
	Date   time.Time          `bson:"date"`
}

// Day is the structure of the documents in the Days collection.
type Day struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Date    time.Time          `bson:"date"`
	People  People             `bson:"people"`
	MaxRank int                `bson:"maxRank"`
	MinRank int                `bson:"minRank"`
}

// Person is the XP for a given Discord user ID.
type Person struct {
	UserID string `bson:"userid"`
	XP     int    `bson:"xp"`
	Rank   int    `bson:"rank"`
}

// Nickname is a cache of the user's nickname.
type Nickname struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      string             `bson:"userid"`
	Nickname    string             `bson:"nickname"`
	LastUpdated time.Time          `bson:"date"`
}

// HistoryRange is a list of moments of a user's XP.
type HistoryRange struct {
	History []History
	UserID  string
}
