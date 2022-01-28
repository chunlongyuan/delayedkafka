package initial

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"dk/config"
	"dk/share/xassert"
)

var (
	DefDB *gorm.DB
)

func InitGoOrm(opts ...MySqlOption) *gorm.DB {

	cfg := config.Cfg

	o := MySqlOptions{
		Username: cfg.DBUsername,
		Password: cfg.DBPassword,
		Hostname: cfg.DBHostname,
		Port:     cfg.DBPort,
		Database: cfg.DBDatabase,
	}
	for _, opt := range opts {
		opt(&o)
	}

	xassert.AssertNotEmpty(o.Username, "empty username")
	xassert.AssertNotEmpty(o.Password, "empty password")
	xassert.AssertNotEmpty(o.Hostname, "empty hostname")
	xassert.AssertNotEmpty(o.Port, "empty post")

	db, err := gorm.Open(
		mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
			o.Username, o.Password, o.Hostname, o.Port, o.Database)),
		&gorm.Config{AllowGlobalUpdate: true},
	)
	if err != nil {
		panic(err)
	}

	switch cfg.Env {
	case "debug":
		return db.Debug()
	}
	return db
}

type MySqlOptions struct {
	Username string
	Password string
	Hostname string
	Port     string
	Database string
}

type MySqlOption func(*MySqlOptions)
