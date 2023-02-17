# mackerel-to-grafana-oncall

A proxy of Mackerel alert webhook to [Grafana OnCall](https://grafana.com/products/oncall/) Webhook.

## Usage

```console
$ ./mackerel-to-grafana-oncall [-port xxxx]
```

Environment Variables:
- `GRAFANA_ONCALL_URL`: Your Grafana OnCall integrations WebHook URL. (Formatted Webhook is recommended)

This server endpoint works as Mackerel alerting Webhook URL.

## Description

This server accepts [Mackerel Webhook](https://mackerel.io/ja/docs/entry/howto/alerts/webhook) events, and deleagates to [Grafana OnCall](https://grafana.com/products/oncall/).

## LISENCE

MIT
