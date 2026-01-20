package cluster

import (
	"axolotl/v1/node"
	"errors"

	"github.com/redis/rueidis"
)

type Masters struct {
	client  rueidis.Client
	masters []node.Master
}

func NewMasters(config *Config) (Masters, error) {
	if len(config.Addresses) == 0 {
		return Masters{}, errors.New("invalid masters-cluster: need at least one master")
	}

	client, err := rueidis.NewClient(rueidis.ClientOption{
		Username:    config.Username,
		Password:    config.Password,
		InitAddress: config.Addresses,
		ReplicaOnly: false,
	})
	if err != nil {
		return Masters{}, err
	}

	return Masters{
		client: client,
	}, nil
}

func (it Masters) Client() rueidis.Client {
	return it.client
}
