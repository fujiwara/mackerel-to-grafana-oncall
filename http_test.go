package oncall_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	oncall "github.com/fujiwara/mackerel-to-grafana-oncall"
	"github.com/google/go-cmp/cmp"
)

var testCases = []struct {
	input  oncall.MackerelWebhook
	status int
	want   oncall.GrafanaOnCallFormattedWebhook
}{
	{
		input: oncall.MackerelWebhook{
			Orgname:  "test org",
			Event:    "alert",
			Memo:     "test memo",
			ImageURL: "https://example.com/alerts/1234.png",
			Alert: oncall.MackerelAlert{
				ID:          "1234",
				Status:      "critical",
				URL:         "https://example.com/alerts/1234",
				Monitorname: "test monitor",
				Isopen:      true,
			},
		},
		status: http.StatusOK,
		want: oncall.GrafanaOnCallFormattedWebhook{
			AlertUID:              "1234",
			Title:                 "[test org] test monitor is critical",
			ImageURL:              "https://example.com/alerts/1234.png",
			LinkToUpstreamDetails: "https://example.com/alerts/1234",
			State:                 "alerting",
			Message:               "test memo",
		},
	},
	{
		input: oncall.MackerelWebhook{
			Orgname:  "test org",
			Event:    "alert",
			Memo:     "test memo",
			ImageURL: "https://example.com/alerts/1234.png",
			Alert: oncall.MackerelAlert{
				ID:          "9999",
				Status:      "unknown",
				URL:         "https://example.com/alerts/1234",
				Monitorname: "test monitor",
				Isopen:      true,
			},
		},
		status: http.StatusNoContent,
	},
	{
		input: oncall.MackerelWebhook{
			Orgname:  "test org",
			Event:    "alert",
			Memo:     "test memo",
			ImageURL: "https://example.com/alerts/1234.png",
			Alert: oncall.MackerelAlert{
				ID:          "1234",
				Status:      "ok",
				URL:         "https://example.com/alerts/1234",
				Monitorname: "test monitor",
				Isopen:      false,
			},
		},
		status: http.StatusOK,
		want: oncall.GrafanaOnCallFormattedWebhook{
			AlertUID:              "1234",
			Title:                 "[test org] test monitor is ok",
			ImageURL:              "https://example.com/alerts/1234.png",
			LinkToUpstreamDetails: "https://example.com/alerts/1234",
			State:                 "ok",
			Message:               "test memo",
		},
	},
}

func TestHTTPServer(t *testing.T) {
	oncall.CriticalOnly = true
	ch := make(chan oncall.GrafanaOnCallFormattedWebhook, 1)
	grafanaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hook := oncall.GrafanaOnCallFormattedWebhook{}
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
		oncall.GrafanaOnCallURL = grafanaServer.URL
		oncall.AllowOnCallURLParam = true
		oncall.HandleWebhook(w, r)
	}))
	defer proxyServer.Close()

	// test
	for _, tc := range testCases {
		payload := tc.input
		t.Run(payload.IncidentTitle(), func(t *testing.T) {
			b, _ := json.Marshal(payload)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "POST", proxyServer.URL, bytes.NewReader(b))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("failed to post: %s", err)
				return
			}
			if resp.StatusCode != tc.status {
				t.Errorf("unexpected status code: %d (expected %d", resp.StatusCode, tc.status)
			}
			if tc.status != http.StatusOK {
				return
			}

			// check
			timeout := time.After(3 * time.Second)
			select {
			case <-timeout:
				t.Errorf("timeout")
			case grafanaRecieved := <-ch:
				expected := tc.want
				if d := cmp.Diff(expected, grafanaRecieved); d != "" {
					t.Errorf("unexpected GrafanaOnCallFormattedWebhook: %s", d)
				}
			}
		})
	}
}
