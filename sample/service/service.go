package service

import (
	"context"
	"fmt"

	"github.com/kuroko-shirai/axolotl/v1/cobweb"
	"github.com/redis/rueidis"
)

type (
	Cobweb interface {
		Execute(context.Context, cobweb.Executor) ([]rueidis.RedisResult, error)
	}

	Redis interface {
		Hget(string, string) rueidis.Completed
		Get(string) rueidis.Completed
	}

	Config struct {
		Cobweb Cobweb
		Redis  Redis
	}

	Service struct {
		Cobweb Cobweb
		Redis  Redis
	}
)

func New(config Config) (Service, error) {
	return Service{
		Cobweb: config.Cobweb,
		Redis:  config.Redis,
	}, nil
}

func (it *Service) GetChangePointsForShop(ctx context.Context, shopID int32) (string, error) {
	cmd := cobweb.SingleCmd{Cmd: it.Redis.Get(fmt.Sprintf("change:points:%d", shopID))}
	result, err := it.Cobweb.Execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return result[0].ToString()
}
