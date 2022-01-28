package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"dk/restful"
)

func main() {

	var delayDuration time.Duration
	flag.DurationVar(&delayDuration, "delay", time.Second*2, "delay duration")
	flag.Parse()

	url := `http://localhost:8000/dk/v1/messages`
	contentType := `application/json`

	sentFunc := func() {

		delaySeconds := strconv.Itoa(int(delayDuration.Seconds()))
		createdAtMs := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

		fmt.Println(delaySeconds, createdAtMs)

		form := restful.PublishForm{
			Topic:       "test-topic",
			DelaySecond: delaySeconds,
			CreatedAtMs: createdAtMs,
			Body: struct {
				Age  int    `json:"age"`
				Name string `json:"name"`
			}{Age: 18, Name: `ppx`},
		}
		body, _ := json.Marshal(form)

		resp, err := http.Post(url, contentType, bytes.NewReader(body))
		if err != nil {
			fmt.Println(err)
			return
		}

		defer resp.Body.Close()

		bytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(bytes))
	}

	for i := 0; i < 10; i++ {
		sentFunc()
	}
}
