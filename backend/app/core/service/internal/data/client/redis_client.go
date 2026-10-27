package client

import (
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/redis/go-redis/v9"
	"go-wind-shop/pkg/logx"
	"go-wind-shop/pkg/redisx"
)

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(ctx *bootstrap.Context) (*redis.Client, func(), error) {
	l := logx.New(ctx.NewLoggerHelper("redis/data/core-service"))
	cfg := ctx.GetConfig()
	if cfg == nil || cfg.Data == nil {
		return nil, func() {}, nil
	}
	cli := redisx.NewClient(cfg.Data, l)
	return cli, func() { _ = cli.Close() }, nil
}
