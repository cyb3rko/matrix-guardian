package util

import (
	"context"
	"fmt"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func getRoomUrl(roomId id.RoomID) string {
	return fmt.Sprintf("https://matrix.to/#/%s", roomId)
}

func getUserUrl(userId id.UserID) string {
	return fmt.Sprintf("https://matrix.to/#/%s", userId)
}

func getHtmlUrl(url string, text string) string {
	return fmt.Sprintf("<a href='%s'>%s</a>", url, text)
}

func GetRoomHtmlUrl(roomId id.RoomID) string {
	return getHtmlUrl(getRoomUrl(roomId), roomId.String())
}

func GetUserHtmlUrl(userId id.UserID) string {
	return getHtmlUrl(getUserUrl(userId), userId.String())
}

func SendHtmlNotice(client *mautrix.Client, ctx context.Context, mngtRoomId id.RoomID, rawMessage string, message string) {
	contentJson := &event.MessageEventContent{
		MsgType:       event.MsgNotice,
		Format:        event.FormatHTML,
		Body:          rawMessage,
		FormattedBody: message,
	}
	_, err := client.SendMessageEvent(ctx, mngtRoomId, event.EventMessage, contentJson)
	if err != nil {
		return
	}
}
