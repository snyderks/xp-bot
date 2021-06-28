package db

import (
	"context"
	"time"
)

var testURI = "mongodb://localhost:27017"

func testContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// func TestCreateDB(t *testing.T) {
// 	client, err := CreateDB(testURI)
// 	if err != nil {
// 		t.Error(err.Error())
// 	}

// 	ctx, cancel := testContext()
// 	defer cancel()
// 	err = client.Ping(ctx, readpref.Primary())
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// }

// func TestGetPeople(t *testing.T) {
// 	people := []string{"98183996185255936"}

// 	client, _ := CreateDB(testURI)

// 	_, _, err := client.ReadPeople(people, 10)

// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// }

// func TestAddDay(t *testing.T) {
// 	people := map[string]primitives.Person{"98183996185255936": primitives.Person{UserID: "98183996185255936", XP: 5, Rank: 1}}

// 	// Add a record that should just add and work.
// 	err := AddDay(people)
// 	if err != nil {
// 		t.Error("Failed to add the record.", err.Error())
// 	}

// 	// Should read this one back.
// 	result, err := ReadNewestDay()
// 	if err != nil {
// 		t.Error("Failed to retrieve a result.", err.Error())
// 	}
// 	if len(result.People) == 0 || result.People[0].UserID != people["ode"].UserID ||
// 		result.People[0].XP != people["ode"].XP ||
// 		result.People[0].Rank != people["ode"].Rank {
// 		t.Error("Retrieved record mismatch. Expected", people, "got", result)
// 	}

// 	people2 := map[string]primitives.Person{"ode": primitives.Person{UserID: "ode", XP: 5, Rank: 1},
// 		"croot": primitives.Person{UserID: "croot", XP: 2, Rank: 2}}

// 	client.AddDay(people2)

// 	// Reassigning here because AddDay scribbles all over people2.
// 	people2 = map[string]primitives.Person{"ode": primitives.Person{UserID: "ode", XP: 5, Rank: 1},
// 		"croot": primitives.Person{UserID: "croot", XP: 2, Rank: 2}}
// 	if err != nil {
// 		t.Error("Failed to add the record.", err.Error())
// 	}

// 	result, err = ReadNewestDay()
// 	if err != nil {
// 		t.Error("Failed to retrieve a result.", err.Error())
// 	}
// 	// A long equality check, but good to be verbose here.
// 	if len(result.People) != 2 || result.People[0].UserID != people2["ode"].UserID ||
// 		result.People[0].XP != people2["ode"].XP ||
// 		result.People[0].Rank != people2["ode"].Rank ||
// 		result.People[1].UserID != people2["croot"].UserID ||
// 		result.People[1].XP != people2["croot"].XP ||
// 		result.People[1].Rank != people2["croot"].Rank {
// 		t.Error("Retrieved record mismatch. Expected", people2, "got", result)
// 	}

// 	ctx, cancel := testContext()
// 	defer cancel()
// 	client.Days.DeleteOne(ctx, bson.M{"_id": result.ID})
// }

// func TestAddPerson(t *testing.T) {
// 	person := map[string]primitives.Person{"ode": primitives.Person{UserID: "ode", XP: 5, Rank: 1}}

// 	client, _ := CreateDB(testURI)

// 	err := AddPeople(person)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	result, _ := client.readMostRecentPerson("ode")

// 	ctx, cancel := testContext()
// 	defer cancel()
// 	client.People.DeleteOne(ctx, bson.M{"_id": result.ID})
// }
