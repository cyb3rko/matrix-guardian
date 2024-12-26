package main

import (
	"context"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"strings"
)

const help = "🛡️ <b>Guardian Help Page</b> 🛡️:<br/>" +
	"<code>!gd <<option>> <<args>></code><br/><br>" +
	"<b>Options</b>:<br/>" +
	"<code>url</code>: <i>Handle URLs in messages</i>"

const urlHelp = "🛡️ <b>Guardian Help Page [url]</b> 🛡️:<br/>" +
	"<code>!gd url <<args>></code><br/><br/>" +
	"<b>Arguments</b>:<br/>" +
	"<code>block <<domain>></code>: <i>Block domain in messages</i><br/>" +
	"<code>unblock <<domain>></code>: <i>Unblock domain in messages</i><br/>" +
	"<code>list</code>: <i>List blocked domains in messages</i>"

const mimeHelp = "🛡️ <b>Guardian Help Page [mime]</b> 🛡️:<br/>" +
	"<code>!gd mime <<args>></code><br/><br/>" +
	"<b>Arguments</b>:<br/>" +
	"<code>block <<mimetype>></code>: <i>Block MIME type in messages</i><br/>" +
	"<code>unblock <<mimetype>></code>: <i>Unblock MIME type in messages</i><br/>" +
	"<code>list</code>: <i>List blocked MIME type in messages</i>"

func getRawMessage(source string) string {
	source = strings.ReplaceAll(source, "<b>", "")
	source = strings.ReplaceAll(source, "</b>", "")
	source = strings.ReplaceAll(source, "<i>", "")
	source = strings.ReplaceAll(source, "</i><br/>", "; ")
	source = strings.ReplaceAll(source, "</i>", " ")
	source = strings.ReplaceAll(source, "<br/><br/>", ". ")
	source = strings.ReplaceAll(source, "<br/>", " ")
	source = strings.ReplaceAll(source, "<code>", "")
	source = strings.ReplaceAll(source, "</code>", "")
	source = strings.ReplaceAll(source, "<<", "<")
	source = strings.ReplaceAll(source, ">>", ">")
	return source
}
func getFormmatedMessage(source string) string {
	source = strings.ReplaceAll(source, "<<", "&lt;")
	source = strings.ReplaceAll(source, ">>", "&gt;")
	return source
}

func sendHtmlMessage(client *mautrix.Client, ctx context.Context, mngtRoomId id.RoomID, message string) {
	contentJson := &event.MessageEventContent{
		MsgType:       event.MsgNotice,
		Format:        event.FormatHTML,
		Body:          getRawMessage(message),
		FormattedBody: getFormmatedMessage(message),
	}
	_, err := client.SendMessageEvent(ctx, mngtRoomId, event.EventMessage, contentJson)
	if err != nil {
		return
	}
}

func ShowHelp(client *mautrix.Client, ctx context.Context, mngtRoomId id.RoomID) {
	sendHtmlMessage(client, ctx, mngtRoomId, help)
}

func ShowUrlHelp(client *mautrix.Client, ctx context.Context, mngtRoomId id.RoomID) {
	sendHtmlMessage(client, ctx, mngtRoomId, urlHelp)
}

func ShowMimeHelp(client *mautrix.Client, ctx context.Context, mngtRoomId id.RoomID) {
	sendHtmlMessage(client, ctx, mngtRoomId, mimeHelp)
}
