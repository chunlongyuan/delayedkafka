package locker

// mock
//go:generate mockgen -source=locker.go -destination=locker_mock.go -package=locker

import (
	"context"
	"errors"
)

var (
	ErrLocked = errors.New("already locked")
)

type Locker interface {
	Lock(ctx context.Context, key, value string, expire int) error
	UnLock(ctx context.Context, key, value string) error
}
