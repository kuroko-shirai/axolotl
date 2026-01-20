package redis

import (
	"context"
	"time"

	"github.com/redis/rueidis"
)

type (
	Config struct {
		Username  string
		Password  string
		Addresses []string
	}

	Redis struct {
		Client rueidis.Client
		TTL    time.Duration
	}
)

func New(config *Config) (Redis, error) {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		Username:    config.Username,
		Password:    config.Password,
		InitAddress: config.Addresses,
	})
	if err != nil {
		return Redis{}, err
	}

	return Redis{
		Client: client,
	}, nil
}

func (it *Redis) Hget(key, field string) rueidis.Completed {
	return it.Client.B().Hget().Key(key).Field(field).Build()
}

func (it *Redis) Get(key string) rueidis.Completed {
	return it.Client.B().Get().Key(key).Build()
}

func (it *Redis) SMembers(key string) rueidis.Completed {
	return it.Client.B().Smembers().Key(key).Build()
}

func (it *Redis) Keys(ctx context.Context, pattern string) rueidis.Completed {
	return it.Client.B().Keys().Pattern(pattern).Build()
}

func (it *Redis) Scan(ctx context.Context, cursor uint64, match string, count int64) rueidis.Completed {
	return it.Client.B().Scan().Cursor(cursor).Match(match).Count(count).Build()
}

func (it *Redis) Hgetall(key string) rueidis.Completed {
	return it.Client.B().Hgetall().Key(key).Build()
}

func (it *Redis) Hmget(key string, fields ...string) rueidis.Completed {
	return it.Client.B().Hmget().Key(key).Field(fields...).Build()
}

func (it *Redis) Hscan(ctx context.Context, key string, fieldMatch string, cursor uint64, count int64) rueidis.Completed {
	return it.Client.B().Hscan().Key(key).Cursor(cursor).Match(fieldMatch).Count(count).Build()
}

// Команды с кэшированием:

func (it *Redis) CTGet(key string, ttl time.Duration) rueidis.CacheableTTL {
	return rueidis.CT(it.Client.B().Get().Key(key).Cache(), ttl)
}

// Команды на групповое выполнение:

func (it *Redis) DoMultiCache(ctx context.Context, commands ...rueidis.CacheableTTL) []rueidis.RedisResult {
	return it.Client.DoMultiCache(ctx, commands...)
}

func (it *Redis) DoMulti(ctx context.Context, multi ...rueidis.Completed) []rueidis.RedisResult {
	return it.Client.DoMulti(ctx, multi...)
}

func (it *Redis) DoCache(ctx context.Context, cmd rueidis.CacheableTTL) rueidis.RedisResult {
	return it.Client.DoCache(ctx, cmd.Cmd, cmd.TTL)
}

func (it *Redis) Close() {
	it.Client.Close()
}
