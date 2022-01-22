package main

import (
	"context"
	"fmt"
	"time"

	. "github.com/Shopify/sarama"
)

type exampleConsumerGroupHandler struct{}

func (exampleConsumerGroupHandler) Setup(_ ConsumerGroupSession) error   { return nil }
func (exampleConsumerGroupHandler) Cleanup(_ ConsumerGroupSession) error { return nil }
func (h exampleConsumerGroupHandler) ConsumeClaim(sess ConsumerGroupSession, claim ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d value:%v on:%v\n", msg.Topic, msg.Partition, msg.Offset, string(msg.Value), time.Now().Format("20060102 15:04:05"))
		sess.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	config := NewConfig()
	config.Version = V2_0_0_0 // specify appropriate version
	config.Consumer.Return.Errors = true

	group, err := NewConsumerGroup([]string{"127.0.0.1:9092"}, "my-group", config)
	if err != nil {
		panic(err)
	}
	defer func() { _ = group.Close() }()

	// Track errors
	go func() {
		for err := range group.Errors() {
			fmt.Println("ERROR", err)
		}
	}()

	// Iterate over consumer sessions.
	ctx := context.Background()
	for {
		topics := []string{"test-topic"}
		handler := exampleConsumerGroupHandler{}

		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		err := group.Consume(ctx, topics, handler)
		if err != nil {
			panic(err)
		}
	}
}
