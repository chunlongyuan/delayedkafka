package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"kdqueue/messenger"
	"kdqueue/restful"
)

func main() {

	var delayDuration time.Duration
	flag.DurationVar(&delayDuration, "delay", time.Second*2, "delay duration")
	flag.Parse()

	url := `http://localhost:8000/kdqueue/v1/messages`
	contentType := `application/json`

	form := restful.PublishForm{
		Topic: "test-topic",
		Message: messenger.Message{
			DelayMs:     delayDuration.Milliseconds(),
			Body:        `{\"a\":\"b\",\"c\":10}`,
			CreatedAtMs: time.Now().UnixNano() / 1e6,
		},
	}

	body, _ := json.Marshal(form)

	resp, err := http.Post(url, contentType, bytes.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
}
