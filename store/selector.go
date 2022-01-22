package store

import (
	"errors"
)

var (
	// ErrNoneAvailable is returned by select when no store were provided
	ErrNoneAvailable = errors.New("none available")
)

type SelectorOptions struct {
	ID string
}

type SelectorOption func(*SelectorOptions)

// Selector selects a store
type Selector interface {
	// Select a store using the strategy
	Select([]Store, ...SelectorOption) (Next, error)

	Record(Store, error) error
	// Reset the selector
	Reset() error
	// String returns the name of the selector
	String() string
}

// Next returns the next store
type Next func() Store
