package filter

import (
	"database/sql"
	"github.com/joeguo/tldextract"
	"matrix-guardian/db"
	"maunium.net/go/mautrix/event"
	"strings"
)

const RegexUrl = "[a-zA-Z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?"

func IsUrlFiltered(database *sql.DB, urls []string) bool {
	for _, u := range urls {
		if db.IsDomainBlocked(database, u) {
			return true
		}
	}
	return false
}

func ParseValidUrls(urls []string) []string {
	var result []string
	extract, _ := tldextract.New("data/tld.cache", true)
	for _, u := range urls {
		parsedUrl := extract.Extract(u)
		if parsedUrl.Root == "" || parsedUrl.Tld == "" {
			continue
		}
		urlString := parsedUrl.Root + "." + parsedUrl.Tld
		if isDomainTrusted(urlString) {
			continue
		}
		result = append(result, urlString)
	}
	return result
}

func DropMentionedUsers(body string, users *event.Mentions) string {
	for _, user := range users.UserIDs {
		body = strings.ReplaceAll(body, user.String(), "")
	}
	return body
}

func isDomainTrusted(domain string) bool {
	return domain == "matrix.org" || domain == "matrix.to"
}
