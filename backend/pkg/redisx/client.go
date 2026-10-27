// Package redisx 是 kratos-bootstrap/cache/redis 的升级兼容层：
// 原 NewClient 依赖旧版 kratos *log.Helper 签名，logger v0.1.3 后在此内联同逻辑。
package redisx

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"
	"go-wind-shop/pkg/logx"
)

// NewClient create go-redis client
func NewClient(conf *conf.Data, logger *logx.Helper) (rdb *redis.Client) {
	if rdb = redis.NewClient(&redis.Options{
		Addr:         conf.GetRedis().GetAddr(),
		Password:     conf.GetRedis().GetPassword(),
		DB:           int(conf.GetRedis().GetDb()),
		DialTimeout:  conf.GetRedis().GetDialTimeout().AsDuration(),
		WriteTimeout: conf.GetRedis().GetWriteTimeout().AsDuration(),
		ReadTimeout:  conf.GetRedis().GetReadTimeout().AsDuration(),
	}); rdb == nil {
		logger.Fatalf("failed opening connection to redis")
		return nil
	}

	// open tracing instrumentation.
	if conf.GetRedis().GetEnableTracing() {
		if err := redisotel.InstrumentTracing(rdb); err != nil {
			logger.Fatalf("failed open tracing: %s", err.Error())
			return nil
		}
	}

	// open metrics instrumentation.
	if conf.GetRedis().GetEnableMetrics() {
		if err := redisotel.InstrumentMetrics(rdb); err != nil {
			logger.Fatalf("failed open metrics: %s", err.Error())
			return nil
		}
	}

	return rdb
}
