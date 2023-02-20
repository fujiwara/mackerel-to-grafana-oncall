package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	main "github.com/fujiwara/mackerel-to-grafana-oncall"
	"github.com/google/go-cmp/cmp"
)

func TestHTTPServer(t *testing.T) {
	ch := make(chan main.GrafanaOnCallFormattedWebhook, 1)
	grafanaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hook := main.GrafanaOnCallFormattedWebhook{}
		if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
			t.Errorf("failed to decode request body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		ch <- hook
		io.WriteString(w, "Ok")
	}))
	defer grafanaServer.Close()

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		main.GrafanaOnCallURL = grafanaServer.URL
		main.AllowOnCallURLParam = true
		main.HandleWebhook(w, r)
	}))
	defer proxyServer.Close()

	// test
	payload := main.MackerelWebhook{
		Orgname:  "test org",
		Event:    "alert",
		Memo:     "test memo",
		ImageURL: "https://example.com/alerts/1234.png",
		Alert: main.MackerelAlert{
			ID:          "1234",
			Status:      "critical",
			URL:         "https://example.com/alerts/1234",
			Monitorname: "test monitor",
			Isopen:      true,
		},
	}
	b, _ := json.Marshal(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", proxyServer.URL, bytes.NewReader(b))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("failed to post: %s", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: %d (expected %d", resp.StatusCode, http.StatusOK)
	}

	// check
	timeout := time.After(3 * time.Second)
	select {
	case <-timeout:
		t.Errorf("timeout")
	case grafanaRecieved := <-ch:
		expected := main.GrafanaOnCallFormattedWebhook{
			AlertUID:              "1234",
			Title:                 "[test org] test monitor is critical",
			ImageURL:              "https://example.com/alerts/1234.png",
			LinkToUpstreamDetails: "https://example.com/alerts/1234",
			State:                 "alerting",
			Message:               "test memo",
		}
		if d := cmp.Diff(expected, grafanaRecieved); d != "" {
			t.Errorf("unexpected GrafanaOnCallFormattedWebhook: %s", d)
		}
	}
}
