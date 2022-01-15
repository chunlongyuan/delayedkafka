package xenv

import (
	"os"
	"strconv"
	"time"
)

func Duration(key string, def time.Duration) time.Duration {
	env := os.Getenv(key)
	if len(key) == 0 || len(env) == 0 {
		return def
	}
	duration, err := time.ParseDuration(env)
	if err != nil {
		return def
	}
	return duration
}

func String(key string, def string) string {
	env := os.Getenv(key)
	if len(key) == 0 || len(env) == 0 {
		return def
	}
	return env
}

func Int(key string, def int) int {
	env := os.Getenv(key)
	if len(key) == 0 || len(env) == 0 {
		return def
	}
	value, err := strconv.Atoi(env)
	if err != nil {
		return def
	}
	return value
}
