package node

import (
	"errors"

	"github.com/redis/rueidis"
)

type Master struct {
	client   rueidis.Client
	replicas map[Replica]struct{}
	address  string
}

func NewMaster(config *NodeConfig) (Master, error) {
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
		address:  config.Address,
		replicas: make(map[Replica]struct{}),
		client:   client,
	}, nil
}

func (it Master) Address() string {
	return it.address
}

func (it Master) AddReplica(replica *Replica) {
	it.replicas[*replica] = struct{}{}
}

func (it Master) ReplicasAddress() []string {
	replicasAddress := make([]string, 0, len(it.replicas))
	for replica := range it.replicas {
		replicasAddress = append(replicasAddress, replica.Address())
	}

	return replicasAddress
}

func (it Master) Client() rueidis.Client {
	return it.client
}
