package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cyb3rko/matrix-botc/botc"
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
	"strings"
	"time"
)

var config Config
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
	syncer.OnEventType(event.StateMember, func(ctx context.Context, evt *event.Event) {
		onRoomInvite(client, ctx, evt)
	})
	syncCtx, cancelSync := context.WithCancel(context.Background())

	commandMapping := map[string]botc.Command{
		"url": {
			false,
			1,
			nil,
			func() { ShowUrlHelp(client, syncCtx, config.mngtRoomId) },
			map[string]botc.Command{
				"list": {
					true,
					0,
					func(evt *event.Event, args []string) { onUrlList(client, syncCtx) },
					nil,
					nil,
				},
				"block": {
					true,
					1,
					func(evt *event.Event, args []string) { onUrlBlock(client, syncCtx, database, evt, args) },
					nil,
					nil,
				},
				"unblock": {
					true,
					1,
					func(evt *event.Event, args []string) { onUrlUnblock(client, syncCtx, evt, args) },
					nil,
					nil,
				},
			},
		},
		"mime": {
			false,
			1,
			nil,
			func() { ShowMimeHelp(client, syncCtx, config.mngtRoomId) },
			map[string]botc.Command{
				"list": {
					true,
					0,
					func(evt *event.Event, args []string) { onMimeList(client, syncCtx) },
					nil,
					nil,
				},
				"block": {
					true,
					1,
					func(evt *event.Event, args []string) { onMimeBlock(client, syncCtx, evt, args) },
					nil,
					nil,
				},
				"unblock": {
					true,
					1,
					func(evt *event.Event, args []string) { onMimeUnblock(client, syncCtx, evt, args) },
					nil,
					nil,
				},
			},
		},
	}
	botc.RegisterCommands(&botc.Config{
		Prefix:       "!gd",
		Mapping:      commandMapping,
		HelpFunction: func() { ShowHelp(client, syncCtx, config.mngtRoomId) },
	})
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
			onManagementMessage(evt)
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

func onManagementMessage(evt *event.Event) {
	botc.ProcessCommandChain(evt.Content.AsMessage().Body, evt)
}

func onUrlList(client *mautrix.Client, ctx context.Context) {
	list, err := db.ListDomains(database)
	if err != nil {
		return
	}
	message := fmt.Sprintf(
		"Configured domains to block:\n%s",
		strings.Join(list, "\n"),
	)
	_, _ = client.SendNotice(ctx, config.mngtRoomId, message)
}

func onUrlBlock(client *mautrix.Client, ctx context.Context, database *sql.DB, evt *event.Event, args []string) {
	success, response := db.BlockDomain(database, args[0])
	if success {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "✅")
	} else {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "❌")
		_, _ = client.SendNotice(ctx, evt.RoomID, response)
	}
}

func onUrlUnblock(client *mautrix.Client, ctx context.Context, evt *event.Event, args []string) {
	success, response := db.UnblockDomain(database, args[0])
	if success {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "✅")
	} else {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "❌")
		_, _ = client.SendNotice(ctx, evt.RoomID, response)
	}
	return
}

func onMimeList(client *mautrix.Client, ctx context.Context) {
	list, err := db.ListMimes(database)
	if err != nil {
		return
	}
	message := fmt.Sprintf(
		"Configured MIME types to block:\n%s",
		strings.Join(list, "\n"),
	)
	_, err = client.SendNotice(ctx, config.mngtRoomId, message)
	if err != nil {
		return
	}
	return
}

func onMimeBlock(client *mautrix.Client, ctx context.Context, evt *event.Event, args []string) {
	success, response := db.BlockMime(database, args[0])
	if success {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "✅")
	} else {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "❌")
		_, _ = client.SendNotice(ctx, evt.RoomID, response)
	}
	return
}

func onMimeUnblock(client *mautrix.Client, ctx context.Context, evt *event.Event, args []string) {
	success, response := db.UnblockMime(database, args[0])
	if success {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "✅")
	} else {
		_, _ = client.SendReaction(ctx, evt.RoomID, evt.ID, "❌")
		_, _ = client.SendNotice(ctx, evt.RoomID, response)
	}
	return
}

func onProtectedRoomMessage(client *mautrix.Client, ctx context.Context, evt *event.Event) {
	const keyMessageType = "msgtype"
	const keyInfo = "info"
	const keyMimetype = "mimetype"
	const keyAttachmentUrl = "url"

	contentJson, err := json.Marshal(evt.Content.Parsed)
	var contentParsed map[string]interface{}
	err = json.Unmarshal(contentJson, &contentParsed)
	if err != nil {
		return
	}
	messageType := contentParsed[keyMessageType]
	if messageType == "m.text" || messageType == "m.notice" || messageType == "m.emote" {
		// text message
		if !config.useUrlFilter && !config.useUrlCheckVt && !config.useUrlCheckFf {
			return
		}
		body := evt.Content.AsMessage().Body
		body = filter.DropMentionedUsers(body, evt.Content.AsMessage().Mentions)
		reg := regexp.MustCompile(filter.RegexUrl)
		urlStrings := reg.FindAllString(body, -1)
		urls := filter.ParseValidUrls(urlStrings)
		if len(urls) == 0 {
			if !config.hiddenMode {
				err := client.SendReceipt(ctx, evt.RoomID, evt.ID, event.ReceiptTypeRead, nil)
				if err != nil {
					return
				}
			}
			return
		}
		if config.useUrlFilter && filter.IsUrlFiltered(database, urls) {
			redactMessage(client, ctx, evt, "found blocklisted URL")
			return
		}
		if config.useUrlCheckVt && check.HasVirusTotalWarning(config.virusTotalKey, urlStrings) {
			redactMessage(client, ctx, evt, "found suspicious URL (VirusTotal)")
			return
		}
		if config.useUrlCheckFf && check.HasFishFishWarning(urlStrings, client.UserID.String()) {
			redactMessage(client, ctx, evt, "found suspicious URL (FishFish)")
			return
		}
		if config.hiddenMode {
			return
		}
		_, err := client.SendReaction(ctx, evt.RoomID, evt.ID, "🛡️")
		if err != nil {
			return
		}
	} else {
		// file message
		if !config.useMimeFilter && !config.useVirusCheckVt {
			// skip file check when no related option is activated
			return
		}
		var mimetype string
		if messageInfo, exists := contentParsed[keyInfo]; exists {
			// try to extract file MIME type
			if _, exists = messageInfo.(map[string]interface{})[keyMimetype]; exists {
				mimetype = messageInfo.(map[string]interface{})[keyMimetype].(string)
				if config.useMimeFilter && db.IsMimeBlocked(database, mimetype) {
					redactMessage(client, ctx, evt, "found blocklisted MIME type")
					return
				}
			}
		}
		if strings.HasPrefix(mimetype, "image/") {
			// ignore images and don't show reaction emoji, only mark it as read
			if !config.hiddenMode {
				_ = client.SendReceipt(ctx, evt.RoomID, evt.ID, event.ReceiptTypeRead, nil)
			}
			return
		}
		if config.useVirusCheckVt {
			messageAttachmentUrl := contentParsed[keyAttachmentUrl]
			if messageAttachmentUrl != nil {
				contentUri, err := id.ParseContentURI(messageAttachmentUrl.(string))
				if err == nil {
					download, err := client.Download(ctx, contentUri)
					if err == nil && download != nil && download.Body != nil {
						if check.HasVirusTotalFinding(config.virusTotalKey, download.Body) {
							redactMessage(client, ctx, evt, "detected malicious file (VirusTotal)")
							return
						}
					}
				}
			}
		}
		if config.hiddenMode {
			return
		}
		_, err := client.SendReaction(ctx, evt.RoomID, evt.ID, "🛡️")
		if err != nil {
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
	message := getRedactNotice(reason, evt)
	rawMessage := getRawRedactNotice(reason, evt)
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

func getRedactNotice(reason string, evt *event.Event) string {
	roomId := evt.RoomID
	userId := evt.Sender
	template := "Message redacted - %s;<br/>" +
		"User %s in room %s :<br/>" +
		"<blockquote>%s</blockquote>"
	return fmt.Sprintf(template, reason, util.GetUserHtmlUrl(userId), util.GetRoomHtmlUrl(roomId), evt.Content.AsMessage().Body)
}

func getRawRedactNotice(reason string, evt *event.Event) string {
	roomId := evt.RoomID.String()
	userId := evt.Sender.String()
	template := "Message redacted - %s; User '%s' in room '%s': %s"
	return fmt.Sprintf(template, reason, userId, roomId, evt.Content.AsMessage().Body)
}

func onRoomInvite(client *mautrix.Client, ctx context.Context, evt *event.Event) {
	if evt.GetStateKey() == client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite {
		_, err := client.JoinRoomByID(ctx, evt.RoomID)
		if err == nil {
			rawMessage := fmt.Sprintf("Joined room after invite: %s", evt.RoomID.String())
			util.Print(rawMessage)
			message := fmt.Sprintf(
				"Guardian Note 🛡️:<br/>"+
					"Joined room after invite: %s",
				util.GetRoomHtmlUrl(evt.RoomID),
			)
			util.SendHtmlNotice(client, ctx, config.mngtRoomId, rawMessage, message)
		} else {
			rawMessage := fmt.Sprintf("Failed to join room after invite: %s", evt.RoomID.String())
			util.Print(rawMessage)
			message := fmt.Sprintf(
				"Guardian Note 🛡️:<br/>"+
					"Failed to join room after invite: %s",
				util.GetRoomHtmlUrl(evt.RoomID),
			)
			util.SendHtmlNotice(client, ctx, config.mngtRoomId, rawMessage, message)
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
	hiddenMode := util.GetEnv("GUARDIAN_HIDDEN_MODE", true, true)
	virusTotalKey := util.GetEnv("GUARDIAN_VIRUS_TOTAL_KEY", true, false)
	useUrlFilter := util.GetEnv("GUARDIAN_URL_FILTER", true, true)
	useUrlCheckVt := util.GetEnv("GUARDIAN_URL_CHECK_VIRUS_TOTAL", true, true)
	useUrlCheckFf := util.GetEnv("GUARDIAN_URL_CHECK_FISHFISH", true, true)
	useMimeFilter := util.GetEnv("GUARDIAN_MIME_FILTER", true, true)
	useVirusCheckVt := util.GetEnv("GUARDIAN_VIRUS_CHECK_VIRUS_TOTAL", true, true)
	mngtRoomReportsBool := true
	testModeBool := false
	hiddenModeBool := false
	useUrlFilterBool := true
	useUrlCheckVtBool := false
	useUrlCheckFfBool := false
	useMimeFilterBool := true
	useVirusCheckVtBool := false

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
	util.Printf("Greetings, %s!", username)
	if password == "" {
		fmt.Println("No password provided!")
		os.Exit(1)
	}
	CheckForDefaultConfig(username, password)
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
	if hiddenMode == "true" {
		hiddenModeBool = true
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
	if useMimeFilter == "false" {
		useMimeFilterBool = false
	}
	if useVirusCheckVt == "true" {
		if virusTotalKey == "" {
			fmt.Println("No VirusTotal API key provided!")
			os.Exit(1)
		}
		useVirusCheckVtBool = true
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
		hiddenMode:      hiddenModeBool,
		virusTotalKey:   virusTotalKey,
		useUrlFilter:    useUrlFilterBool,
		useUrlCheckVt:   useUrlCheckVtBool,
		useUrlCheckFf:   useUrlCheckFfBool,
		useMimeFilter:   useMimeFilterBool,
		useVirusCheckVt: useVirusCheckVtBool,
	}
	return config
}
