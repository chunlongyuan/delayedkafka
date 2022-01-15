package restful

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"kdqueue/messenger"
	"kdqueue/share/xid"
	_ "kdqueue/xtesting"
)

func TestPublishWithProducer(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("test publish", t, func() {

		p := messenger.NewMockProducer(ctrl)

		recorder := httptest.NewRecorder()
		_, engine := gin.CreateTestContext(recorder)

		Convey("should 200", func() {

			id := xid.Get()

			msg := PublishForm{}
			msg.Topic = `this is topic`
			msg.CreatedAtMs = time.Now().UnixNano() / 1e6
			msg.DelayMs = 1000
			msg.Body = `{"a":"b","c":10}`

			p.EXPECT().Publish(gomock.Any(), msg.Topic, msg.Message).Return(id, nil).Times(1)

			body, err := json.Marshal(&msg)
			So(err, ShouldBeNil)

			engine.POST("/messages", PublishWithProducer(p))
			req, _ := http.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})

		Convey("should 400", func() {

			engine.POST("/messages", PublishWithProducer(p))
			req, _ := http.NewRequest(http.MethodPost, "/messages", bytes.NewReader(nil))
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("should 404", func() {

			engine.POST("/messages", PublishWithProducer(p))
			req, _ := http.NewRequest(http.MethodDelete, "/messages", bytes.NewReader(nil))
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("should 500", func() {

			id := xid.Get()

			msg := PublishForm{}
			msg.Topic = `this is topic`
			msg.CreatedAtMs = time.Now().UnixNano() / 1e6
			msg.DelayMs = 1000
			msg.Body = `{"a":"b","c":10}`

			p.EXPECT().Publish(gomock.Any(), msg.Topic, msg.Message).Return(id, errors.New(`mocked err`)).Times(1)

			body, err := json.Marshal(&msg)
			So(err, ShouldBeNil)

			engine.POST("/messages", PublishWithProducer(p))
			req, _ := http.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
