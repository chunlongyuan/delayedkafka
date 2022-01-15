package restful

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"kdqueue/messenger"
)

type PublishForm struct {
	Topic string `json:"topic"`
	messenger.Message
}

func PublishWithProducer(p messenger.Producer) gin.HandlerFunc {

	return func(c *gin.Context) {

		var msg PublishForm

		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		id, err := p.Publish(context.Background(), msg.Topic, msg.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
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
