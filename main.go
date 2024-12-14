package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"matrix-guardian/check"
	"matrix-guardian/db"
	"matrix-guardian/filter"
	"matrix-guardian/util"
	"matrix-guardian/validation"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"os"
	"regexp"
	"time"
)

var config Config
var client *mautrix.Client
var database *sql.DB

func main() {
	fmt.Println("Hello human, Guardian is starting up...")
	config = readConfig()
	database = db.InitDB()
	client, withBatchToken := createClient()
	//_, err := readline.New("[no room]> ")
	//if err != nil {
	//	panic(err)
	//}
	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, func(ctx context.Context, evt *event.Event) {
		onMessage(client, ctx, evt)
	})
	syncer.OnEventType(event.StateMember, onRoomInvite)

	syncCtx, cancelSync := context.WithCancel(context.Background())
	_, err := client.Login(syncCtx, &mautrix.ReqLogin{
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
	if err == nil {
		util.Printf("Joined rooms: %s", list.JoinedRooms)
	} else {
		util.Printf("No joined rooms found")
	}
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
	// ignore own messages
	if evt.Sender == client.UserID {
		return
	}
	if evt.RoomID == config.mngtRoomId {
		// message in management room
		if !config.testMode {
			onManagementMessage(client, ctx, evt)
		} else {
			onProtectedRoomMessage(client, ctx, evt)
		}
	} else {
		// message in protected room
		if !config.testMode {
			onProtectedRoomMessage(client, ctx, evt)
		}
	}
}

func onManagementMessage(client *mautrix.Client, ctx context.Context, evt *event.Event) {
	message := evt.Content.AsMessage().Body
	if util.IsGuardianCommand(message) {
		command, subcommands := util.ParseCommands(message)
		util.Printf("Received management command: %s %s", command, subcommands)
		switch command {
		case "url":
			if len(subcommands) > 0 {
				switch subcommands[0] {
				case "block":
					if len(subcommands) == 2 {
						db.BlockDomain(database, subcommands[1])
						return
					}
				case "unblock":
					if len(subcommands) == 2 {
						db.UnblockDomain(database, subcommands[1])
						return
					}
				}
			}
			ShowUrlHelp(client, ctx, config.mngtRoomId)
		default:
			ShowHelp(client, ctx, config.mngtRoomId)
		}
	}
}

func onProtectedRoomMessage(client *mautrix.Client, ctx context.Context, evt *event.Event) {
	const keyMessageType = "msgtype"

	contentJson, err := json.Marshal(evt.Content.Parsed)
	var contentParsed map[string]interface{}
	err = json.Unmarshal(contentJson, &contentParsed)
	if err != nil {
		return
	}
	messageType := contentParsed[keyMessageType]
	if messageType == "m.text" || messageType == "m.notice" || messageType == "m.emote" {
		if !config.useUrlFilter && !config.useUrlCheckVt && !config.useUrlCheckFf {
			return
		}
		reg := regexp.MustCompile(filter.RegexUrl)
		urls := reg.FindAllString(evt.Content.AsMessage().Body, -1)
		if config.useUrlFilter && filter.IsUrlFiltered(database, urls) {
			redactMessage(client, ctx, evt, "found blocklisted URL")
			return
		}
		if config.useUrlCheckVt && check.HasVirusTotalWarning(config.virusTotalKey, urls) {
			redactMessage(client, ctx, evt, "found suspicious URL (VirusTotal)")
			return
		}
	}
}

func redactMessage(client *mautrix.Client, ctx context.Context, evt *event.Event, reason string) {
	_, err := client.RedactEvent(ctx, evt.RoomID, evt.ID)
	if err != nil {
		return
	}
	if !config.mngtRoomReports {
		return
	}
	message := fmt.Sprintf("Message redacted - %s:<br/><blockquote>%s</blockquote>", reason, evt.Content.AsMessage().Body)
	rawMessage := fmt.Sprintf("Message redacted - %s:%s", reason, evt.Content.AsMessage().Body)
	contentJson := &event.MessageEventContent{
		MsgType:       event.MsgNotice,
		Format:        event.FormatHTML,
		Body:          rawMessage,
		FormattedBody: message,
	}
	_, err = client.SendMessageEvent(ctx, config.mngtRoomId, event.EventMessage, contentJson)
	if err != nil {
		return
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
	homeserver := util.GetEnv("GUARDIAN_HOMESERVER", true, true)
	username := util.GetEnv("GUARDIAN_USERNAME", true, true)
	password := util.GetEnv("GUARDIAN_PASSWORD", false, false)
	mngtRoomId := util.GetEnv("GUARDIAN_MANAGEMENT_ROOM_ID", true, false)
	mngtRoomReports := util.GetEnv("GUARDIAN_MANAGEMENT_ROOM_REPORTS", true, true)
	testMode := util.GetEnv("GUARDIAN_TEST_MODE", true, true)
	virusTotalKey := util.GetEnv("GUARDIAN_VIRUS_TOTAL_KEY", true, false)
	useUrlFilter := util.GetEnv("GUARDIAN_URL_FILTER", true, true)
	useUrlCheckVt := util.GetEnv("GUARDIAN_URL_CHECK_VIRUS_TOTAL", true, true)
	useUrlCheckFf := util.GetEnv("GUARDIAN_URL_CHECK_FISHFISH", true, true)
	mngtRoomReportsBool := true
	testModeBool := false
	useUrlFilterBool := true
	useUrlCheckVtBool := false
	useUrlCheckFfBool := false

	// REQUIRED //
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

	// OPTIONAL //
	if mngtRoomReports == "false" {
		mngtRoomReportsBool = false
	}
	if testMode == "true" {
		testModeBool = true
		fmt.Println("!!! Running in test mode !!!")
	}
	if useUrlFilter == "false" {
		useUrlFilterBool = false
	}
	if useUrlCheckVt == "true" {
		if virusTotalKey == "" {
			fmt.Println("No VirusTotal API key provided!")
			os.Exit(1)
		}
		useUrlCheckVtBool = true
	}
	if useUrlCheckFf == "true" {
		useUrlCheckFfBool = true
	}

	config = Config{
		// REQUIRED //
		homeserver: homeserver,
		username:   username,
		password:   password,
		mngtRoomId: id.RoomID(mngtRoomId),
		// OPTIONAL //
		mngtRoomReports: mngtRoomReportsBool,
		testMode:        testModeBool,
		virusTotalKey:   virusTotalKey,
		useUrlFilter:    useUrlFilterBool,
		useUrlCheckVt:   useUrlCheckVtBool,
		useUrlCheckFf:   useUrlCheckFfBool,
	}
	return config
}
