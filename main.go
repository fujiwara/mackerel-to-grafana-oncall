package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/fujiwara/ridge"
)

var (
	AllowOnCallURLParam = false
)

func init() {
	AllowOnCallURLParam = os.Getenv("ALLOW_ONCALL_URL_PARAM") == "true"
	if !AllowOnCallURLParam && os.Getenv("GRAFANA_ONCALL_URL") == "" {
		log.Fatal("GRAFANA_ONCALL_URL is required")
	}
}

func main() {
	var mux = http.NewServeMux()
	mux.HandleFunc("/webhook", handleWebhook)
	ridge.Run(":8000", "/", mux)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		errorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}
	var hook MackerelWebhook
	err := json.NewDecoder(r.Body).Decode(&hook)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("[info] recieved %s is %s", hook.IncidentTitle(), hook.Alert.Status)

	if s := hook.Alert.Status; s == "warning" {
		errorResponse(w, http.StatusNotModified, fmt.Errorf("alert status is %s. ignored.", s))
		return
	}
	var grafanaHook = hook.ToGrafanaOnCallFormattedWebhook()

	var onCallURL string
	if onCallURL = os.Getenv("GRAFANA_ONCALL_URL"); onCallURL == "" {
		if AllowOnCallURLParam {
			onCallURL = r.FormValue("oncall_url")
		}
	}
	if onCallURL == "" {
		errorResponse(w, http.StatusBadRequest, fmt.Errorf("oncall_url is required"))
		return
	}

	if err := postToGrafanaOnCall(onCallURL, grafanaHook); err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	log.Printf("[info] posted to %s", onCallURL)
}

func errorResponse(w http.ResponseWriter, code int, err error) {
	log.Printf("[error] %d %s", code, err)
	w.WriteHeader(code)
}

func postToGrafanaOnCall(onCallURL string, hook GrafanaOnCallFormattedWebhook) error {
	u, err := url.Parse(onCallURL)
	if err != nil {
		return err
	}

	b, err := json.Marshal(hook)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
