package db

import (
	"database/sql"
	"fmt"
)

const keyNextBatchToken = "next_batch"

func GetNextBatchToken(db *sql.DB) string {
	row := db.QueryRow("SELECT value FROM attributes WHERE key = ?", keyNextBatchToken)
	var value string
	err := row.Scan(&value)
	if err != nil {
		fmt.Println("No 'next_batch' token found")
		return ""
	}
	return value
}

func SaveNextBatchToken(db *sql.DB, token string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO attributes(key, value) VALUES (?, ?)", keyNextBatchToken, token)
	if err != nil {
		return err
	}
	return nil
}
