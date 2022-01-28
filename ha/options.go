package ha

import "dk/locker"

type Options struct {
	Locker locker.Locker
	NodeId string
}

type Option func(options *Options)
