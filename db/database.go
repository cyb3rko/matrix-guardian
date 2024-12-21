package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type table struct {
	name, values string
}

type mimetype struct {
	name  string
	count int
}

var tables = []table{
	{"domains", "name TEXT PRIMARY KEY, count INT"},
	{"mimetypes", "name TEXT PRIMARY KEY, count INT"},
	{"attributes", "key TEXT PRIMARY KEY, value TEXT"},
}

const driver = "sqlite3"
const dbPath = "data/data.sqlite"

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

func IsMimeBlocked(db *sql.DB, mime string) bool {
	query := db.QueryRow("SELECT count FROM mimetypes WHERE name = ?", mime)
	var count int
	err := query.Scan(&count)
	if err != nil {
		// not found in database, implicitly allowed
		return false
	}
	// update usage counter
	_, _ = db.Exec("UPDATE mimetypes SET count = ? WHERE name = ?", count+1, mime)
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

func BlockMime(db *sql.DB, mime string) bool {
	_, err := db.Exec("INSERT INTO mimetypes (name, count) values (?, 0)", mime)
	return err == nil
}

func UnblockMime(db *sql.DB, mime string) bool {
	_, err := db.Exec("DELETE FROM mimetypes WHERE name = ?", mime)
	return err == nil
}

func ListMimes(db *sql.DB) ([]string, error) {
	query, err := db.Query("SELECT name, count FROM mimetypes ORDER BY count DESC")
	if err != nil {
		return nil, err
	}
	var rows []string
	for query.Next() {
		var row mimetype
		_ = query.Scan(&row.name, &row.count)
		rows = append(rows, fmt.Sprintf("- %s (%d)", row.name, row.count))
	}
	return rows, nil
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
