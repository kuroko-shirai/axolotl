package node

import (
	"context"

	"github.com/redis/rueidis"
)

type (
	Node interface {
		Client() rueidis.Client
		LoadThreshold(ctx context.Context) (float64, error)
	}

	Config struct {
		Password string
		Username string
	}

	NodeConfig struct {
		Config
		Address string
	}
)
