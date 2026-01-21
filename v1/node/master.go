package node

import (
	"errors"

	"github.com/redis/rueidis"
)

type (
	Master struct {
		client  rueidis.Client
		address string
	}
)

func NewMaster(config *Config) (Master, error) {
	if len(config.Address) == 0 {
		return Master{}, errors.New("invalid master address")
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
		return Master{}, err
	}

	return Master{
		address: config.Address,
		client:  client,
	}, nil
}

func (it Master) Address() string {
	return it.address
}

func (it Master) Client() rueidis.Client {
	return it.client
}
