package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type table struct {
	name, values string
}

var tables = []table{
	{"domains", "name TEXT PRIMARY KEY, count INT"},
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

func IsDomainBlocked(db *sql.DB, domain string) bool {
	query := db.QueryRow("SELECT count FROM domains WHERE name = ?", domain)
	var count int
	err := query.Scan(&count)
	if err != nil {
		// not found in database, implicitly allowed
		return false
	}
	// update usage counter
	_, _ = db.Exec("UPDATE domains SET count = ? WHERE name = ?", count+1, domain)
	// found in database, explicitly blocked
	return true
}

func BlockDomain(db *sql.DB, domain string) bool {
	_, err := db.Exec("INSERT INTO domains (name, count) values (?, 0)", domain)
	return err == nil
}

func UnblockDomain(db *sql.DB, domain string) bool {
	_, err := db.Exec("DELETE FROM domains WHERE name = ?", domain)
	return err == nil
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
