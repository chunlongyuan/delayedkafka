package xtesting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-uuid"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"kdqueue/config"
	"kdqueue/initial"
)

func init() {

	rand.Seed(time.Now().UnixNano())

	if len(os.Getenv("ENV")) == 0 {
		// if no ENV specified
		envFile := ".env"
		for i := 0; i < 10; i++ {
			if _, err := os.Stat(envFile); err != nil {
				envFile = "../" + envFile
				continue
			}
			break
		}
		// local debugging
		err := godotenv.Load(envFile)
		if err != nil {
			panic(err)
		}

		logrus.Trace("dot env loaded")
	}

	err := env.Parse(&config.Cfg)
	if err != nil {
		panic(err)
	}

	logrus.SetLevel(logrus.DebugLevel)

	initial.DefDB = initial.InitGoOrm().Debug()
	initial.DefRedisPool = initial.InitRedis()
}

func ReadFile(path string) (string, error) {
	var (
		err   error
		bytes []byte
	)
	for i := 0; i < 10; i++ {
		bytes, err = ioutil.ReadFile(path)
		if err != nil {
			path = "../" + path
			continue
		}
		return string(bytes), nil
	}
	return "", fmt.Errorf("read empty from %s", path)
}

func GinContextWithMethod(userID string, storeID uint64, target, method string, params map[string]interface{}) (c *gin.Context, r *gin.Engine) {
	return GinContextWithOptions(func(opt *GinCtxOptions) {
		opt.UserID = userID
		opt.Target = target
		opt.Method = method
		opt.Params = params
	})
}

func GinContext(userID string, storeID uint64, target string, params map[string]interface{}) (c *gin.Context, r *gin.Engine) {
	return GinContextWithOptions(func(opt *GinCtxOptions) {
		opt.UserID = userID
		opt.Target = target
		opt.Params = params
	})
}

func GinContextWithOptions(opts ...GinCtxOption) (c *gin.Context, r *gin.Engine) {
	opt := GinCtxOptions{
		UserID: "100",
		Method: http.MethodGet,
	}
	for _, o := range opts {
		o(&opt)
	}
	if len(opt.Target) == 0 {
		opt.Target = "/test"
	}
	c, r = gin.CreateTestContext(httptest.NewRecorder())
	if opt.Params != nil && len(opt.Params) > 0 {
		paramsBytes, _ := json.Marshal(opt.Params)
		c.Request = httptest.NewRequest(opt.Method, opt.Target, bytes.NewBuffer(paramsBytes))
	} else {
		c.Request = httptest.NewRequest(opt.Method, opt.Target, nil)
	}
	c.Request.Header.Add("Content-Type", "application/json")
	for k, v := range opt.Header {
		c.Request.Header.Add(k, v)
	}
	return
}

func RandomStr() string {
	uuid, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}
	return strings.ReplaceAll(uuid, "-", "")
}

func RandomInt() int {
	return rand.Int()
}

// options

type GinCtxOptions struct {
	UserID string
	Target string
	Method string
	Params map[string]interface{}
	Header map[string]string
}

type GinCtxOption func(options *GinCtxOptions)
