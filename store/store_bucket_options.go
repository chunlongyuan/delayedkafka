package store

import (
	"time"
)

type BucketStoreOptions struct {
	BucketCount  int
	Selector     Selector
	WaitDuration time.Duration
}

type BucketStoreOption func(*BucketStoreOptions)
