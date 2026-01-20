package cluster

import (
	"errors"

	"github.com/redis/rueidis"
)

type Replicas struct {
	client rueidis.Client
}

func NewReplicas(config *Config) (Replicas, error) {
	if len(config.Addresses) == 0 {
		return Replicas{}, errors.New("invalid replicas-cluster: need at least one replica")
	}

	client, err := rueidis.NewClient(rueidis.ClientOption{
		Username:    config.Username,
		Password:    config.Password,
		InitAddress: config.Addresses,
		ReplicaOnly: true,
	})
	if err != nil {
		return Replicas{}, err
	}

	return Replicas{
		client: client,
	}, nil
}

func (it Replicas) Client() rueidis.Client {
	return it.client
}
