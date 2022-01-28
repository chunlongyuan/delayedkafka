package restful

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"dk/config"
	"dk/messenger"
)

func Run(ctx context.Context) error {

	producer := messenger.DefProducer

	cfg := config.Cfg

	engine := gin.Default()

	v1 := engine.Group("dk/v1")
	{
		v1.POST("/messages", PublishWithProducer(producer))  // 投递消息
		v1.DELETE("/messages", CancelWithProducer(producer)) // 删除消息
	}

	go func() {
		// 服务连接
		if err := manners.ListenAndServe(":"+cfg.Port, engine); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case <-ctx.Done():
		logrus.Infoln("Context done, Shutdown Server ...")
	case <-quit:
		logrus.Infoln("Shutdown Server ...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if ok := manners.Close(); !ok {
		logrus.Fatal("Server Shutdown fail")
	}
	logrus.Println("Server exiting")
	return nil
}
