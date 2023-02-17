package main

import "fmt"

type MackerelWebhook struct {
	Orgname  string        `json:"orgName"`
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

func (h MackerelWebhook) IncidentTitle() string {
	return fmt.Sprintf("[%s] %s", h.Orgname, h.Alert.Monitorname)
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
		ImageURL:              h.Memo,
		State:                 alerting,
		LinkToUpstreamDetails: h.Alert.URL,
		Message:               h.Memo,
	}
}
