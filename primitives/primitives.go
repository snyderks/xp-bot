package primitives

import (
	"time"

	"go.mongodb.org/mongo-driver/primitive"
)

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
	People  People             `bson:"people"`
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
