package oncall

import (
	"fmt"
	"strings"
)

type MackerelWebhook struct {
	Orgname  string        `json:"orgName"`
	Event    string        `json:"event"`
	ImageURL string        `json:"imageURL"`
	Memo     string        `json:"memo"`
	Alert    MackerelAlert `json:"alert"`
}

type MackerelAlert struct {
	Monitorname       string  `json:"monitorName"`
	Criticalthreshold int     `json:"criticalThreshold"`
	Metricvalue       float64 `json:"metricValue"`
	Monitoroperator   string  `json:"monitorOperator"`
	Trigger           string  `json:"trigger"`
	URL               string  `json:"url"`
	Openedat          *int64  `json:"openedAt"`
	Duration          *int64  `json:"duration"`
	Createdat         *int64  `json:"createdAt"`
	Isopen            bool    `json:"isOpen"`
	Metriclabel       string  `json:"metricLabel"`
	ID                string  `json:"id"`
	Closedat          *int64  `json:"closedAt"`
	Status            string  `json:"status"`
}

func (h MackerelWebhook) IsTestPayload() bool {
	return h.Orgname == "" && h.Alert.ID == "" && h.Alert.Status == ""
}

func (h MackerelWebhook) IsAlertEvent() bool {
	return h.Event == "alert"
}

func (h MackerelWebhook) IsCriticalOrOK() bool {
	s := strings.ToLower(h.Alert.Status)
	return s == "critical" || s == "ok"
}

func (h MackerelWebhook) ID() string {
	return h.Alert.ID
}

func (h MackerelWebhook) IncidentTitle() string {
	return fmt.Sprintf("[%s] %s is %s", h.Orgname, h.Alert.Monitorname, h.Alert.Status)
}

func (h MackerelWebhook) ToGrafanaOnCallFormattedWebhook() GrafanaOnCallFormattedWebhook {
	var alerting string
	if h.Alert.Isopen {
		alerting = "alerting"
	} else {
		alerting = "ok"
	}
	return GrafanaOnCallFormattedWebhook{
		AlertUID:              h.Alert.ID,
		Title:                 h.IncidentTitle(),
		ImageURL:              h.ImageURL,
		State:                 alerting,
		LinkToUpstreamDetails: h.Alert.URL,
		Message:               h.Memo,
	}
}
