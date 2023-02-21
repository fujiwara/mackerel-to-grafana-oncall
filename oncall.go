package oncall

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fujiwara/logutils"
	"github.com/fujiwara/ridge"
)

var (
	Version                 = "current"
	AllowOnCallURLParam     = false
	Debug                   = false
	GrafanaOnCallURL        = ""
	GrafanaOnCallURLAliases = ""
	GrafanaTimeout          = 30 * time.Second

	Aliases = OnCallURLAliases{}
)

func validate() error {
	if !AllowOnCallURLParam && GrafanaOnCallURL == "" {
		return errors.New("-grafana-oncall-url or -allow-oncall-url-param is required")
	}
	return nil
}

func envToFlag(f *flag.Flag) {
	name := strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
	if s, ok := os.LookupEnv(name); ok {
		f.Value.Set(s)
	}
}

func Run() error {
	var port int
	var showVersion bool
	flag.BoolVar(&Debug, "debug", false, "debug mode")
	flag.BoolVar(&AllowOnCallURLParam, "allow-oncall-url-param", false, "allow Grafana oncall by url param")
	flag.StringVar(&GrafanaOnCallURL, "grafana-oncall-url", "", "Grafana oncall webhook url")
	flag.StringVar(&GrafanaOnCallURLAliases, "grafana-oncall-url-aliases", "", "Grafana oncall webhook url aliases(string of JSON object)")
	flag.IntVar(&port, "port", 8000, "port number")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.VisitAll(envToFlag)
	flag.Parse()

	if showVersion {
		fmt.Println(Version)
		return nil
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "info", "warn", "error"},
		MinLevel: logutils.LogLevel("info"),
		Writer:   os.Stderr,
	}
	if Debug {
		filter.MinLevel = logutils.LogLevel("debug")
	}
	log.SetOutput(filter)

	if err := validate(); err != nil {
		return err
	}
	log.Println("[info] starting mackerel-to-grafana-oncall version:", Version)
	log.Println("[info] grafana-oncall-url:", maskURL(GrafanaOnCallURL))
	log.Println("[info] allow-oncall-url-param:", AllowOnCallURLParam)
	if GrafanaOnCallURLAliases != "" {
		log.Println("[info] grafana-oncall-url-aliases:", GrafanaOnCallURLAliases)
		if err := parseAliases(GrafanaOnCallURLAliases, &Aliases); err != nil {
			return err
		}
		for a, u := range Aliases {
			log.Println("[info] alias:", a, "->", maskURL(u))
		}
	}

	var mux = http.NewServeMux()
	mux.HandleFunc("/", handleWebhook)
	mux.HandleFunc("/health", handleHealth)
	ridge.Run(fmt.Sprintf(":%d", port), "/", mux)
	return nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func resolveOnCallURL(p string) (string, error) {
	if p == "" {
		return GrafanaOnCallURL, nil
	}
	if strings.HasPrefix(p, "https://") {
		return p, nil
	}
	url, ok := Aliases.FindByAlias(p)
	if !ok {
		return "", fmt.Errorf("oncall_url %s not found in aliases", p)
	}
	return url, nil
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
	log.Printf("[info] recieved webhook: %s", hook.IncidentTitle())
	log.Printf("[debug] %#v", hook)

	var grafanaHook GrafanaOnCallFormattedWebhook
	if hook.IsTestPayload() {
		log.Printf("[info] test payload.")
		grafanaHook = GrafanaOnCallFormattedWebhookTestPayload
	} else if hook.IsAlertEvent() {
		grafanaHook = hook.ToGrafanaOnCallFormattedWebhook()
	} else {
		log.Printf("[info] not alert event. ignored.")
		errorResponse(w, http.StatusNotModified, fmt.Errorf("not alert event. ignored. Event: %s", hook.Event))
		return
	}

	onCallURL, err := resolveOnCallURL(r.URL.Query().Get("oncall_url"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := postToGrafanaOnCall(onCallURL, grafanaHook); err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	log.Printf("[info] posted to %s", maskURL(onCallURL))
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
	ctx, cancel := context.WithTimeout(context.Background(), GrafanaTimeout)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[warn] failed to read response body. %s", err)
	}
	log.Println("[debug] response body:", string(body))
	return nil
}

func maskURL(s string) string {
	return regexp.MustCompile(".{12}$").ReplaceAllString(s, "************")
}
