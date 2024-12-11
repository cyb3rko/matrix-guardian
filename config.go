package main

import "maunium.net/go/mautrix/id"

type Config struct {
	homeserver string
	username   string
	password   string
	mngtRoomId id.RoomID
	testMode   bool
}
