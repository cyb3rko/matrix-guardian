package check

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type FishFishReport struct {
	Category string `json:"category"`
}

const fishFishUrl = "https://api.fishfish.gg/v1/domains/%s"

func HasFishFishWarning(urls []string, clientIdentifier string) bool {
	userAgent := "Matrix Guardian Bot (" + clientIdentifier + ")"
	for _, u := range urls {
		if checkFfSingleUrl(u, userAgent) {
			return true
		}
	}
	return false
}

func checkFfSingleUrl(u string, userAgent string) bool {
	if !strings.HasPrefix(u, "http") {
		u = "http://" + u
	}
	parsedUrl, err := url.Parse(u)
	if err != nil || parsedUrl.Host == "" {
		return false
	}
	report := FishFishReport{}
	responseCode, err := getJson(fmt.Sprintf(fishFishUrl, parsedUrl.Hostname()), userAgent, &report)
	if err != nil || responseCode != 200 {
		return false
	}
	return report.Category != "safe"
}

func getJson(u string, userAgent string, target *FishFishReport) (int, error) {
	req, err := http.NewRequest("GET", u, nil)
	req.Header = http.Header{
		"Content-Type": []string{"application/json"},
		"User-Agent":   []string{userAgent},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}
