package restful

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"dk/messenger"
)

func CancelWithProducer(p messenger.Producer) gin.HandlerFunc {

	return func(c *gin.Context) {

		topic := c.Query("topic")
		idStr := c.Query("id")

		if len(topic) == 0 {
			c.JSON(http.StatusBadRequest, "empty topic")
			return
		}

		if len(idStr) == 0 {
			c.JSON(http.StatusBadRequest, "empty id")
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		if err := p.Cancel(context.Background(), topic, id); err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.Status(http.StatusOK)
	}
}
