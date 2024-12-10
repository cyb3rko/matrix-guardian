package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type table struct {
	name, values string
}

var tables = []table{
	//{""},
	{"attributes", "key TEXT PRIMARY KEY, value TEXT"},
}

const driver = "sqlite3"
const dbPath = "data.sqlite"

func InitDB() *sql.DB {
	database, err := sql.Open(driver, dbPath)
	if err != nil {
		panic(err)
	}
	createAllTables(database)
	return database
}

func createAllTables(db *sql.DB) {
	for _, tab := range tables {
		createTable(db, tab.name, tab.values)
	}
}

func createTable(db *sql.DB, name, values string) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS " + name + " (" + values + ")")
	if err != nil {
		panic(err)
	}
}
