# mackerel-to-grafana-oncall

A proxy of Mackerel alert webhook to [Grafana OnCall](https://grafana.com/products/oncall/) Webhook.

## Usage

```console
Usage of mackerel-to-grafana-oncall:
  -allow-oncall-url-param
        allow Grafana oncall by url param
  -debug
        debug mode
  -grafana-oncall-url string
        Grafana oncall webhook url
  -port int
        port number (default 8000)
```

Environment variables also can set these flags. `GRAFANA_ONCALL_URL`, `ALLOW_ON_CALL_URL_PARAM`, `PORT` and`DEBUG`.

This server endpoint works as Mackerel alerting Webhook URL.

## Description

This server accepts [Mackerel Webhook](https://mackerel.io/ja/docs/entry/howto/alerts/webhook) events, and deleagates to [Grafana OnCall](https://grafana.com/products/oncall/).

## LISENCE

MIT
