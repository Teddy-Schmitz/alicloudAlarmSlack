# Alicloud Alarms to Slack

Send Alicloud alarm webhooks to slack.

## Usage

`go install github.com/Teddy-Schmitz/alicloudAlarmSlack`

Just provide the slack webhook URL as `SLACK_WEBHOOK` and start the server.  Then point your Alicloud alarms to the URL of this service.



You can deploy on Google App Engine with this app.yaml.

```
runtime: go111
env_variables:
  SLACK_WEBHOOK: "https://your url here"
```

## TODO

Lots to do as it's very basic and I made it quickly just for my usage.  PRs are very welcome.