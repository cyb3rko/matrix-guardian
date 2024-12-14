package filter

import (
	"database/sql"
	"matrix-guardian/db"
	"net/url"
	"strings"
)

const RegexUrl = "[a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?"

func IsUrlFiltered(database *sql.DB, urls []string) bool {
	for _, u := range urls {
		if !strings.HasPrefix(u, "http") {
			u = "http://" + u
		}
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
