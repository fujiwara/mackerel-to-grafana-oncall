package main

type GrafanaOnCallFormattedWebhook struct {
	AlertUID              string `json:"alert_uid"`
	Title                 string `json:"title"`
	ImageURL              string `json:"image_url"`
	State                 string `json:"state"`
	LinkToUpstreamDetails string `json:"link_to_upstream_details"`
	Message               string `json:"message"`
}

var GrafanaOnCallFormattedWebhookTestPayload = GrafanaOnCallFormattedWebhook{
	Title: "This is a test webhook from Mackerel",
	State: "ok",
}
