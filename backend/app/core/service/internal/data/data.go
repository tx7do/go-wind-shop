package data

import (	"go-wind-shop/pkg/logx"
	"errors"

	"github.com/go-kratos/kratos/v2/registry"

	"github.com/tx7do/go-utils/password"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	redisx "go-wind-shop/pkg/redisx"
	"github.com/tx7do/kratos-bootstrap/database/opensearch"
	bRegistry "github.com/tx7do/kratos-bootstrap/registry"

	elasticsearchCrud "github.com/tx7do/go-crud/opensearch"

	"github.com/redis/go-redis/v9"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	"go-wind-shop/pkg/oss"
)

// NewRedisClient 创建Redis客户端
func NewRedisClient(ctx *bootstrap.Context) (*redis.Client, func(), error) {
	cfg := ctx.GetConfig()
	if cfg == nil {
		return nil, func() {}, nil
	}

	l := logx.New(ctx.NewLoggerHelper("redis/data/core-service"))

	cli := redisx.NewClient(cfg.Data, l)

	return cli, func() {
		if err := cli.Close(); err != nil {
			l.Error(err)
		}
	}, nil
}

// NewElasticSearchClient 创建 OpenSearch 客户端（与 go-wind-cms 一致，搜索后端为 OpenSearch）
func NewElasticSearchClient(ctx *bootstrap.Context) (*elasticsearchCrud.Client, func(), error) {
	cfg := ctx.GetConfig()
	if cfg == nil {
		return nil, func() {}, nil
	}

	if cfg.Data != nil && cfg.Data.Opensearch == nil {
		return nil, func() {}, errors.New("opensearch config is nil")
	}

	cli, err := opensearch.NewClient(logx.AsKratos(ctx.GetLogger()), cfg)
	if err != nil {
		return nil, func() {}, err
	}

	return cli, func() {
	}, nil
}

// NewDiscovery 创建服务发现客户端
func NewDiscovery(ctx *bootstrap.Context) registry.Discovery {
	cfg := ctx.GetConfig()
	if cfg == nil {
		return nil
	}

	ret, err := bRegistry.NewDiscovery(cfg.Registry)
	if err != nil {
		return nil
	}
	return ret
}

func NewMinIoClient(ctx *bootstrap.Context) *oss.MinIOClient {
	return oss.NewMinIoClient(ctx.GetConfig(), ctx.GetLogger())
}

func NewPasswordCrypto() password.Crypto {
	crypto, err := password.CreateCrypto("bcrypt")
	if err != nil {
		panic(err)
	}
	return crypto
}
