package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type AlicloudAlarm struct {
	UserID        string            `json:"userId"`
	AlertName     string            `json:"alertName"`
	Timestamp     int64             `json:"timestamp"`
	AlertState    string            `json:"alertState"`
	Dimensions    map[string]string `json:"dimensions"`
	Expression    string            `json:"expression"`
	Curvalue      string            `json:"Curvalue"`
	MetricName    string            `json:"metricName"`
	MetricProject string            `json:"metricProject"`
}

var slackURL = os.Getenv("SLACK_WEBHOOK")

var alarmMessage = `[%s] %s`

func main() {
	http.HandleFunc("/", receiveAlarm)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func decodeAliArray(in string) (out map[string]string) {
	out = map[string]string{}
	in, _ = url.QueryUnescape(in)
	if !strings.HasPrefix(in, "{") || !strings.HasSuffix(in, "}") {
		return
	}
	in = strings.TrimPrefix(in, "{")
	in = strings.TrimSuffix(in, "}")

	for _, dim := range strings.Split(in, ",") {
		dim = strings.TrimSpace(dim)
		parsed := strings.Split(dim, "=")
		val, _ := url.QueryUnescape(parsed[1])
		out[parsed[0]] = val
	}

	return
}

func receiveAlarm(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("alertName") == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name, _ := url.QueryUnescape(r.FormValue("alertName"))
	t, _ := strconv.ParseInt(r.FormValue("timestamp"), 10, 64)
	t = t / 1000

	alarm := AlicloudAlarm{
		AlertName:  name,
		Timestamp:  t,
		AlertState: r.FormValue("alertState"),
		MetricName: r.FormValue("metricName"),
		Curvalue:   r.FormValue("curValue"),
		Dimensions: decodeAliArray(r.FormValue("dimensions")),
	}

	sendWebhook(alarm)
}

type SlackWebhook struct {
	Text        string             `json:"text"`
	Attachments []*SlackAttachment `json:"attachments"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SlackAttachment struct {
	Title  string        `json:"title"`
	Text   string        `json:"text"`
	Color  string        `json:"color"`
	Fields []*SlackField `json:"fields"`
}

func sendWebhook(alarm AlicloudAlarm) error {
	msg := &SlackWebhook{
		Attachments: []*SlackAttachment{
			{
				Title: fmt.Sprintf(alarmMessage, alarm.AlertState, alarm.AlertName),
				Text:  fmt.Sprintf(""),
				Color: "danger",
			},
		},
	}

	fields := []*SlackField{
		{
			Title: "Timestamp",
			Value: fmt.Sprintf("<!date^%[1]d^At {time_secs} {date_long_pretty}| At %[1]d (Unix timestamp)>", alarm.Timestamp),
		},
	}

	if val, ok := alarm.Dimensions["instanceId"]; ok {
		fields = append(fields, &SlackField{
			Title: "Instance ID",
			Value: val,
			Short: false,
		})
	}

	if val, ok := alarm.Dimensions["region"]; ok {
		fields = append(fields, &SlackField{
			Title: "Region",
			Value: val,
			Short: true,
		})
	}

	if val, ok := alarm.Dimensions["queue"]; ok {
		fields = append(fields, &SlackField{
			Title: "Queue",
			Value: val,
			Short: true,
		})
	}

	msg.Attachments[0].Fields = fields

	if alarm.AlertState == "OK" {
		msg.Attachments[0].Color = "good"
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = http.Post(slackURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	return nil
}
