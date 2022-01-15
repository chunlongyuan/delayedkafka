package restful

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"kdqueue/messenger"
	"kdqueue/share/xid"
	_ "kdqueue/xtesting"
)

func TestCancelWithProducer(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("test cancel", t, func() {

		p := messenger.NewMockProducer(ctrl)

		recorder := httptest.NewRecorder()
		_, engine := gin.CreateTestContext(recorder)

		Convey("should 200", func() {

			id := xid.Get()
			topic := `this is topic`

			p.EXPECT().Cancel(gomock.Any(), topic, id).Return(nil).Times(1)

			engine.DELETE("/messages", CancelWithProducer(p))
			req, _ := http.NewRequest(http.MethodDelete, "/messages?topic="+topic+"&id="+strconv.FormatUint(id, 10), nil)
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})

		Convey("should 400", func() {

			topic := `this is topic`

			engine.DELETE("/messages", CancelWithProducer(p))
			req, _ := http.NewRequest(http.MethodDelete, "/messages?topic="+topic+"&id=a", nil)
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("should 400 without id", func() {

			topic := `this is topic`

			engine.DELETE("/messages", CancelWithProducer(p))
			req, _ := http.NewRequest(http.MethodDelete, "/messages?topic="+topic+"&id=0", nil)
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("should 400 without topic", func() {

			engine.DELETE("/messages", CancelWithProducer(p))
			req, _ := http.NewRequest(http.MethodDelete, "/messages?topic=&id=10", nil)
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("should 404", func() {

			id := xid.Get()
			topic := `this is topic`

			engine.DELETE("/messages", CancelWithProducer(p))
			req, _ := http.NewRequest(http.MethodPost, "/messages?topic="+topic+"&id="+strconv.FormatUint(id, 10), nil)
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusNotFound)
		})

		SkipConvey("should 500", func() {

			id := xid.Get()
			topic := `this is topic`

			p.EXPECT().Cancel(gomock.Any(), topic, id).Return(errors.New(`mock err`)).Times(1)

			engine.DELETE("/messages", CancelWithProducer(p))
			req, _ := http.NewRequest(http.MethodDelete, "/messages?topic="+topic+"&id="+strconv.FormatUint(id, 10), nil)
			engine.ServeHTTP(recorder, req)
			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
