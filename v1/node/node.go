package node

import (
	"errors"

	"github.com/redis/rueidis"
)

type (
	Node struct {
		client  rueidis.Client
		address string
	}

	Config struct {
		Password string
		Username string
		Address  string
	}
)

func New(config *Config) (Node, error) {
	if len(config.Address) == 0 {
		return Node{}, errors.New("invalid node address")
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
		return Node{}, err
	}

	return Node{
		address: config.Address,
		client:  client,
	}, nil
}

func (it Node) Address() string {
	return it.address
}

func (it Node) Client() rueidis.Client {
	return it.client
}
