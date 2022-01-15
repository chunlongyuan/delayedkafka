package restful

import "context"

// Server api 层抽象的 server, 可实现为 http server 或 rpc server
type Server interface {
	Run(ctx context.Context) error
}
