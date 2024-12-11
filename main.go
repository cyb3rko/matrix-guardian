package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/chzyer/readline"
	"matrix-guardian/db"
	"matrix-guardian/util"
	"matrix-guardian/validation"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"os"
	"strings"
	"time"
)

var config Config
var client *mautrix.Client
var database *sql.DB

func main() {
	fmt.Println("Hello human, Guardian is starting up!")
	config = readConfig()
	database = db.InitDB()
	client, withBatchToken := createClient()
	_, err := readline.New("[no room]> ")
	if err != nil {
		panic(err)
	}
	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, func(ctx context.Context, evt *event.Event) {
		onMessage(client, ctx, evt)
	})
	syncer.OnEventType(event.StateMember, onRoomInvite)

	syncCtx, cancelSync := context.WithCancel(context.Background())
	_, err = client.Login(syncCtx, &mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: config.username},
		Password:         config.password,
		StoreCredentials: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Login successful")
	list, err := client.JoinedRooms(syncCtx)
	util.Printf("Joined rooms: %s", list.JoinedRooms)
	if !withBatchToken {
		resp, err := client.FullSyncRequest(syncCtx, mautrix.ReqSync{
			Since: fmt.Sprintf("s%d", time.Now().UnixMilli()),
		})
		if resp == nil {
			fmt.Println("No response for initial sync")
			os.Exit(1)
		}
		if err != nil {
			fmt.Println("Initial syncing error")
			panic(err)
		}
		//_ = db.SaveNextBatchToken(database, resp.NextBatch)
		err = client.Store.SaveNextBatch(syncCtx, client.UserID, resp.NextBatch)
		if err != nil {
			fmt.Println("Saving 'next_batch' error")
			panic(err)
		}
	}
	fmt.Println("Guarding is running...")
	err = client.Sync()
	go func() {

	}()
	if err != nil {
		fmt.Println("Syncing error")
		panic(err)
	}
	cancelSync()
	//err = cryptoHelper.Close()
	err = database.Close()
	if err != nil {
		fmt.Println("Error closing database")
	}
}

func onMessage(client *mautrix.Client, ctx context.Context, evt *event.Event) {
	if evt.RoomID == config.mngtRoomId {
		if !config.testMode {
			onManagementMessage(evt)
		} else {
			onProtectedRoomMessage(client, ctx, evt)
		}
	} else {
		if !config.testMode {
			onProtectedRoomMessage(client, ctx, evt)
		}
	}
}

func onManagementMessage(evt *event.Event) {
	message := evt.Content.AsMessage().Body
	if util.IsGuardianCommand(message) {
		command, subcommands := util.ParseCommands(message)
		util.Printf("Received management command: %s %s", command, subcommands)
	}
}

func onProtectedRoomMessage(client *mautrix.Client, ctx context.Context, evt *event.Event) {
	if evt.Sender.Localpart() == "cyb3rko" && strings.Contains(evt.Content.AsMessage().Body, "https://t.me") {
		_, err := client.RedactEvent(ctx, evt.RoomID, evt.ID)
		if err != nil {
			return
		}
	}
}

func onRoomInvite(ctx context.Context, evt *event.Event) {
	if evt.GetStateKey() == client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite {
		_, err := client.JoinRoomByID(ctx, evt.RoomID)
		if err == nil {
			util.Printf("Joined room after invite: %s", evt.RoomID.String())
			//rl.SetPrompt(fmt.Sprintf("%s> ", lastRoomID))
			//log.Info().
			//	Str("room_id", evt.RoomID.String()).
			//	Str("inviter", evt.Sender.String()).
			//	Msg("Joined room after invite")
		} else {
			util.Printf("Failed to join room after invite: %s", evt.RoomID.String())
			//log.Error().Err(err).
			//	Str("room_id", evt.RoomID.String()).
			//	Str("inviter", evt.Sender.String()).
			//	Msg("Failed to join room after invite")
		}
	}
}

func createClient() (*mautrix.Client, bool) {
	client, err := mautrix.NewClient(config.homeserver, "", "")
	if err != nil {
		panic(err)
	}
	nextBatchToken := db.GetNextBatchToken(database)
	withBatchToken := false
	if nextBatchToken != "" {
		err := client.Store.SaveNextBatch(context.Background(), client.UserID, nextBatchToken)
		if err != nil {
			panic(err)
		}
		withBatchToken = true
	}
	return client, withBatchToken
}

func readConfig() Config {
	homeserver := util.GetEnv("GUARDIAN_HOME", true, true)
	username := util.GetEnv("GUARDIAN_USER", true, true)
	password := util.GetEnv("GUARDIAN_PASSWORD", false, false)
	mngtRoomId := util.GetEnv("GUARDIAN_MANAGEMENT_ROOM_ID", true, false)
	testMode := util.GetEnv("GUARDIAN_TEST_MODE", true, true)
	testModeBool := false
	if !validation.IsValidUrl(homeserver) {
		util.Printf("Invalid homeserver URL: %s", homeserver)
		os.Exit(1)
	}
	if !validation.IsValidUsername(username) {
		fmt.Println("Invalid username format provided!")
		os.Exit(1)
	}
	if username == "" {
		fmt.Println("No username provided!")
		os.Exit(1)
	}
	if password == "" {
		fmt.Println("No password provided!")
		os.Exit(1)
	}
	if mngtRoomId == "" {
		fmt.Println("No management room ID provided!")
		os.Exit(1)
	}
	if testMode == "true" {
		testModeBool = true
	}
	config = Config{
		homeserver: homeserver,
		username:   username,
		password:   password,
		mngtRoomId: id.RoomID(mngtRoomId),
		testMode:   testModeBool,
	}
	return config
}
