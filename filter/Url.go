package filter

import (
	"database/sql"
	"matrix-guardian/db"
	"maunium.net/go/mautrix/event"
	"net/url"
	"regexp"
)

const regexUrl = "(http:\\/\\/www\\.|https:\\/\\/www\\.|http:\\/\\/|https:\\/\\/)?[a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?"

func IsUrlFiltered(database *sql.DB, content *event.Content) bool {
	reg := regexp.MustCompile(regexUrl)
	urls := reg.FindAllString(content.AsMessage().Body, -1)
	for _, u := range urls {
		parsedUrl, err := url.Parse(u)
		if err != nil {
			return false
		}
		if db.IsDomainBlocked(database, parsedUrl.Host) {
			return true
		}
	}
	return false
}
