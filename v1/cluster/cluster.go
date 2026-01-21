package cluster

import "github.com/redis/rueidis"

type (
	Cluster interface {
		Client() rueidis.Client
	}

	Config struct {
		Addresses    []string
		Username     string
		Password     string
		MaxThreshold float64
	}
)
