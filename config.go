package main

import (
	"fmt"
	"maunium.net/go/mautrix/id"
	"os"
)

const ConfigDefaultUsername = "yourcoolusername"
const ConfigDefaultPassword = "yourVerySecurePassword"

type Config struct {
	// REQUIRED //
	homeserver string    // "GUARDIAN_HOMESERVER"
	username   string    // "GUARDIAN_USERNAME"
	password   string    // "GUARDIAN_PASSWORD"
	mngtRoomId id.RoomID // "GUARDIAN_MANAGEMENT_ROOM_ID"
	// OPTIONAL //
	mngtRoomReports bool   // "GUARDIAN_MANAGEMENT_ROOM_REPORTS", default: true
	testMode        bool   // "GUARDIAN_TEST_MODE",               default: false
	hiddenMode      bool   // "GUARDIAN_HIDDEN_MODE",             default: false
	virusTotalKey   string // "GUARDIAN_VIRUS_TOTAL_KEY",         default:
	useUrlFilter    bool   // "GUARDIAN_URL_FILTER",              default: true
	useUrlCheckVt   bool   // "GUARDIAN_URL_CHECK_VIRUS_TOTAL",   default: false
	useUrlCheckFf   bool   // "GUARDIAN_URL_CHECK_FISHFISH",      default: false
	useMimeFilter   bool   // "GUARDIAN_MIME_FILTER",             default: true
	useVirusCheckVt bool   // "GUARDIAN_VIRUS_CHECK_VIRUS_TOTAL", default: false
}

func CheckForDefaultConfig(username string, password string) {
	if username == ConfigDefaultUsername && password == ConfigDefaultPassword {
		fmt.Println("Default values found, please change them!")
		os.Exit(1)
	}
}
