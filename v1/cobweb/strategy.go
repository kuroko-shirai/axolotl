package cobweb

import (
	"context"

	"github.com/redis/rueidis"
)

type (
	// Executor абстрагирует способ выполнения команд.
	Executor interface {
		Execute(ctx context.Context, client rueidis.Client) ([]rueidis.RedisResult, error)
	}

	// SingleCmd — для одной команды.
	SingleCmd struct {
		Cmd rueidis.Completed
	}

	// MultiCmd — для DoMulti.
	MultiCmd struct {
		Cmds []rueidis.Completed
	}

	// CacheCmd — для DoCache.
	CacheCmd struct {
		Cmd rueidis.CacheableTTL
	}

	// MultiCacheCmd — для DoMultiCache.
	MultiCacheCmd struct {
		Cmds []rueidis.CacheableTTL
	}
)

func (it SingleCmd) Execute(ctx context.Context, client rueidis.Client) ([]rueidis.RedisResult, error) {
	if !it.Cmd.IsReadOnly() {
		return nil, ErrWriteCommand
	}
	return []rueidis.RedisResult{client.Do(ctx, it.Cmd)}, nil
}

func (it MultiCmd) Execute(ctx context.Context, client rueidis.Client) ([]rueidis.RedisResult, error) {
	for _, cmd := range it.Cmds {
		if !cmd.IsReadOnly() {
			return nil, ErrWriteCommand
		}
	}
	return client.DoMulti(ctx, it.Cmds...), nil
}

func (it CacheCmd) Execute(ctx context.Context, client rueidis.Client) ([]rueidis.RedisResult, error) {
	return []rueidis.RedisResult{client.DoCache(ctx, it.Cmd.Cmd, it.Cmd.TTL)}, nil
}

func (it MultiCacheCmd) Execute(ctx context.Context, client rueidis.Client) ([]rueidis.RedisResult, error) {
	return client.DoMultiCache(ctx, it.Cmds...), nil
}
