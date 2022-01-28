package restful

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"dk/messenger"
	"dk/share/xid"
	_ "dk/xtesting"
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

			form := PublishForm{}
			form.Topic = `this is topic`
			form.CreatedAtMs = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
			form.DelaySecond = "2"
			form.Body = struct {
				Name string `json:"name"`
			}{Name: `ppx`}

			msg, err := createMessage(form)
			So(err, ShouldBeNil)

			p.EXPECT().Publish(gomock.Any(), form.Topic, msg).Return(id, nil).Times(1)

			body, err := json.Marshal(&form)
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

			form := PublishForm{}
			form.Topic = `this is topic`
			form.CreatedAtMs = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
			form.DelaySecond = "2"
			form.Body = struct {
				Name string
			}{Name: `ppx`}

			message, err := createMessage(form)
			So(err, ShouldBeNil)

			p.EXPECT().Publish(gomock.Any(), form.Topic, message).Return(id, errors.New(`mocked err`)).Times(1)

			body, err := json.Marshal(&form)
			So(err, ShouldBeNil)

			engine.POST("/messages", PublishWithProducer(p))
			req, _ := http.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
