package check

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/VirusTotal/vt-go"
	"io"
	"mime/multipart"
	"net/http"
)

const endpointFile = "https://www.virustotal.com/api/v3/files/%s"

func newRequest(key string, method string, url string, body io.ReadCloser) (*http.Request, error) {
	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		var fw io.Writer
		if fw, err = writer.CreateFormFile("file", "file"); err != nil {
			return nil, err
		}
		if _, err = io.Copy(fw, body); err != nil {
			return nil, err
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		if req, err = http.NewRequest(method, url, &buf); err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Apikey", key)
	return req, nil
}

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

func HasVirusTotalFinding(key string, file io.ReadCloser) bool {
	hasher := sha256.New()
	_, _ = io.Copy(hasher, file)
	hash := hex.EncodeToString(hasher.Sum(nil))
	req, err := newRequest(key, http.MethodGet, fmt.Sprintf(endpointFile, hash), nil)
	report, err := http.DefaultClient.Do(req)
	if report == nil || err != nil {
		return false
	}
	var body []byte
	if body, err = io.ReadAll(report.Body); err != nil {
		return false
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	data := result["data"]
	if _, valid := data.(map[string]interface{}); !valid {
		return false
	}
	attributes := data.(map[string]interface{})["attributes"]
	if _, valid := attributes.(map[string]interface{}); !valid {
		return false
	}
	stats := attributes.(map[string]interface{})["last_analysis_stats"]
	if _, valid := stats.(map[string]interface{}); !valid {
		return false
	}
	validStats := stats.(map[string]interface{})
	return validStats["malicious"].(float64) > 1 || validStats["suspicious"].(float64) > 3
}

//func HasVirusTotalFinding(key string, file io.ReadCloser) bool {
//	req, err := newRequest(key, http.MethodPost, endpointFiles, file)
//	if err != nil {
//		fmt.Println(err)
//		return false
//	}
//	resp, _ := http.DefaultClient.Do(req)
//	defer func(Body io.ReadCloser) {
//		_ = Body.Close()
//	}(resp.Body)
//	body, err := io.ReadAll(resp.Body)
//	if err != nil || body == nil {
//		return false
//	}
//	var result map[string]interface{}
//	err = json.Unmarshal(body, &result)
//	data := result["data"]
//	if _, valid := data.(map[string]interface{}); !valid {
//		return false
//	}
//	links := data.(map[string]interface{})["links"]
//	if _, valid := links.(map[string]interface{}); !valid {
//		return false
//	}
//	selfLink := links.(map[string]interface{})["self"].(string)
//	counter := 0
//	for counter < 8 {
//		req, err = newRequest(key, http.MethodGet, selfLink, nil)
//		report, _ := http.DefaultClient.Do(req)
//		defer func(Body io.ReadCloser) {
//			_ = Body.Close()
//		}(report.Body)
//		if body, err = io.ReadAll(report.Body); err != nil {
//			return false
//		}
//		err = json.Unmarshal(body, &result)
//		data = result["data"]
//		if _, valid := data.(map[string]interface{}); !valid {
//			return false
//		}
//		attributes := data.(map[string]interface{})["attributes"]
//		if _, valid := attributes.(map[string]interface{}); !valid {
//			return false
//		}
//		status := attributes.(map[string]interface{})["status"]
//		if status != "completed" {
//			// scan not completed (yet)
//			counter++
//			util.Printf("File scan not completed yet (%d)", counter)
//			time.Sleep(2 * time.Second)
//			continue
//		}
//		stats := attributes.(map[string]interface{})["stats"]
//		if _, valid := stats.(map[string]interface{}); !valid {
//			return false
//		}
//		validStats := stats.(map[string]interface{})
//		return validStats["malicious"].(float64) > 1 || validStats["suspicious"].(float64) > 3
//	}
//	return false
//}
