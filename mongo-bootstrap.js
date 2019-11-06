// This file is designed to create the xp-bot db two collections required.
// After creating them, it creates indexes on those collections to increase performance.

// Globally set the db used
db = new Mongo().getDB("xp-bot")

// Create our collections
db.createCollection("days")
db.createCollection("people")

db.days.createIndex({ date: -1 }, { name: "date" })
db.people.createIndex({ date: -1 }, { name: "date" })
db.people.createIndex({ name: 1 }, { name: "name" })