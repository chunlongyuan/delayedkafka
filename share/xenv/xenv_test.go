package xenv

import (
	"os"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	type args struct {
		key           string
		def           time.Duration
		before, after func()
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{name: "got minute should success", args: args{key: "AB_C", def: time.Second, before: func() { os.Setenv("AB_C", "1m") }, after: func() { os.Unsetenv("AB_C") }}, want: time.Minute},
		{name: "got hour should success", args: args{key: "AB_C", def: time.Second, before: func() { os.Setenv("AB_C", "1h") }, after: func() { os.Unsetenv("AB_C") }}, want: time.Hour},
		{name: "empty key should got default value", args: args{key: "    ", def: time.Second, before: func() { os.Setenv("AB_C", "1m") }, after: func() { os.Unsetenv("AB_C") }}, want: time.Second},
		{name: "empty env should got default value", args: args{key: "AB_C", def: time.Second, before: func() {}, after: func() {}}, want: time.Second},
		{name: "not match should fail", args: args{key: "aB_C", def: time.Second, before: func() { os.Setenv("AB_C", "1m") }, after: func() { os.Unsetenv("AB_C") }}, want: time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.before()
			defer tt.args.after()
			if got := Duration(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	type args struct {
		key           string
		def           string
		before, after func()
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "should success", args: args{key: "AB_C", def: "default", before: func() { os.Setenv("AB_C", "1m") }, after: func() { os.Unsetenv("AB_C") }}, want: "1m"},
		{name: "empty key should got default", args: args{key: "  ", def: "default", before: func() { os.Setenv("AB_C", "1m") }, after: func() { os.Unsetenv("AB_C") }}, want: "default"},
		{name: "not match should got default", args: args{key: "aB_C", def: "default", before: func() { os.Setenv("AB_C", "1m") }, after: func() { os.Unsetenv("AB_C") }}, want: "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.before()
			defer tt.args.after()
			if got := String(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt(t *testing.T) {
	type args struct {
		key           string
		def           int
		before, after func()
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "should success", args: args{key: "AB_C", def: 10, before: func() { os.Setenv("abc", "10") }, after: func() { os.Unsetenv("AB_C") }}, want: 10},
		{name: "invalid value should got default", args: args{key: "AB_C", def: 3, before: func() { os.Setenv("abc", "c") }, after: func() { os.Unsetenv("AB_C") }}, want: 3},
		{name: "empty key should got default", args: args{key: "  ", def: 20, before: func() { os.Setenv("AB_C", "20") }, after: func() { os.Unsetenv("AB_C") }}, want: 20},
		{name: "not match should got default", args: args{key: "aB_C", def: 30, before: func() { os.Setenv("AB_C", "30") }, after: func() { os.Unsetenv("AB_C") }}, want: 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.before()
			defer tt.args.after()
			if got := Int(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}
