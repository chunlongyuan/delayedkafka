package messenger

import (
	"dk/store"
)

type ProducerOptions struct {
	Store    store.Store
	Delivery Delivery
}

type ProducerOption func(options *ProducerOptions)
