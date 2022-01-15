package ha

import "kdqueue/locker"

type Options struct {
	Locker locker.Locker
	NodeId string
}

type Option func(options *Options)
