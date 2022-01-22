package restful

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"kdqueue/messenger"
)

type PublishForm struct {
	Topic       string      `json:"topic"`
	DelaySecond string      `json:"delay_second"`  // 延迟多少秒
	CreatedAtMs string      `json:"created_at_ms"` // 毫秒时间戳
	Body        interface{} `json:"body"`          // 元数据
}

func PublishWithProducer(p messenger.Producer) gin.HandlerFunc {

	return func(c *gin.Context) {

		var form PublishForm

		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("err: %v", err.Error())})
			return
		}

		msg, err := createMessage(form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("err: %v", err.Error())})
			return
		}

		id, err := p.Publish(context.Background(), form.Topic, msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Sprintf("err: %v", err.Error())})
			return
		}

		resp := struct {
			ID string `json:"id"`
		}{
			ID: strconv.FormatUint(id, 10),
		}

		c.JSON(http.StatusOK, resp)
	}
}

func createMessage(form PublishForm) (messenger.Message, error) {

	var msg messenger.Message

	delaySecond, err := strconv.ParseInt(form.DelaySecond, 10, 64)
	if err != nil {
		return msg, errors.Wrap(err, "parse delay_second")
	}
	delayMs := delaySecond * 1e3

	createAtMs, err := strconv.ParseInt(form.CreatedAtMs, 10, 64)
	if err != nil {
		return msg, errors.Wrap(err, "parse created_at_ms")
	}

	bodyBytes, err := json.Marshal(form.Body)
	if err != nil {
		return msg, errors.Wrap(err, "marshal body")
	}

	msg.DelayMs = delayMs
	msg.CreatedAtMs = createAtMs
	msg.Body = string(bodyBytes)

	return msg, nil

}
