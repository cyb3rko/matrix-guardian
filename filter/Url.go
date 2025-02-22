package filter

import (
	"database/sql"
	"matrix-guardian/db"
	"net/url"
	"strings"
)

const RegexUrl = "[a-zA-Z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?"

func IsUrlFiltered(database *sql.DB, urls []url.URL) bool {
	for _, u := range urls {
		if db.IsDomainBlocked(database, u.Host) {
			return true
		}
	}
	return false
}

func DropTrustedUrls(urls []string) []url.URL {
	var result []url.URL
	for _, u := range urls {
		u = strings.ToLower(u)
		if !strings.HasPrefix(u, "http") {
			u = "http://" + u
		}
		parsedUrl, err := url.Parse(u)
		if err != nil {
			continue
		}
		if isDomainTrusted(parsedUrl.Host) {
			continue
		}
		result = append(result, *parsedUrl)
	}
	return result
}

func isDomainTrusted(domain string) bool {
	return domain == "matrix.org" || domain == "matrix.to"
}
