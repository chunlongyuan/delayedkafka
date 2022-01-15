package store

import (
	"errors"
)

var (
	// ErrNoneAvailable is returned by select when no routes were provided to select from
	ErrNoneAvailable = errors.New("none available")
)

type SelectorOptions struct {
	ID string
}

type SelectorOption func(*SelectorOptions)

// Selector selects a route from a pool
type Selector interface {
	// Select a route from the pool using the strategy
	Select([]Store, ...SelectorOption) (Next, error)
	// Record the error returned from a route to inform future selection
	Record(Store, error) error
	// Reset the selector
	Reset() error
	// String returns the name of the selector
	String() string
}

// Next returns the next node
type Next func() Store
