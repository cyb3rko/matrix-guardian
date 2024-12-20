package check

import (
	"encoding/base64"
	"encoding/json"
	"github.com/VirusTotal/vt-go"
)

func HasVirusTotalWarning(key string, urls []string) bool {
	vtClient := vt.NewClient(key)
	for _, url := range urls {
		if checkVtSingleUrl(vtClient, url) {
			return true
		}
	}
	return false
}

func checkVtSingleUrl(client *vt.Client, url string) bool {
	urlId := base64.RawURLEncoding.EncodeToString([]byte(url))
	report, err := client.Get(vt.URL("urls/%s", urlId))
	if report == nil || err != nil {
		return false
	}
	var result map[string]interface{}
	err = json.Unmarshal(report.Data, &result)
	attributes := result["attributes"]
	if _, valid := attributes.(map[string]interface{}); !valid {
		return false
	}
	stats := attributes.(map[string]interface{})["last_analysis_stats"]
	if _, valid := stats.(map[string]interface{}); !valid {
		return false
	}
	validStats := stats.(map[string]interface{})
	return validStats["malicious"].(float64)+validStats["suspicious"].(float64) >= 3
}
