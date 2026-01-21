package node

import (
	"errors"

	"github.com/redis/rueidis"
)

type (
	Replica struct {
		client  rueidis.Client
		address string
	}
)

func NewReplica(config *Config) (Replica, error) {
	if len(config.Address) == 0 {
		return Replica{}, errors.New("invalid replica setting: need at least one replica address")
	}

	client, err := rueidis.NewClient(rueidis.ClientOption{
		Username:       config.Username,
		Password:       config.Password,
		InitAddress:    []string{config.Address},
		SendToReplicas: func(cmd rueidis.Completed) bool { return cmd.IsReadOnly() },
		Standalone: rueidis.StandaloneOption{
			ReplicaAddress: []string{config.Address},
		},
	})
	if err != nil {
		return Replica{}, err
	}

	return Replica{
		address: config.Address,
		client:  client,
	}, nil
}

func (it Replica) Address() string {
	return it.address
}

func (it Replica) Client() rueidis.Client {
	return it.client
}
