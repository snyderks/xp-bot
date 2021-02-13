// This file is designed to create the xp-bot db two collections required.
// After creating them, it creates indexes on those collections to increase performance.

// Globally set the db used
db = new Mongo().getDB("xp-bot")

// Create our collections
db.createCollection("days")
db.createCollection("people")
db.createCollection("nicknames")

db.days.createIndex({ date: -1 }, { name: "date" })
db.people.createIndex({ date: -1 }, { name: "date" })
db.people.createIndex({ userid: 1 }, { name: "userid" })
db.nicknames.createIndex({ userid: 1 }, { name: "userid" })