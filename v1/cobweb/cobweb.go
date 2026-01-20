package cobweb

import (
	"axolotl/v1/cluster"
	"context"
	"errors"
	"log"

	"github.com/redis/rueidis"
)

var (
	ErrWriteCommand = errors.New("non-read command routed to cobweb")
)

type (
	Monitor interface {
		Snapshot() map[string]float64
	}

	ReadCommand interface {
		// Build возвращает rueidis.Completed команду.
		Build(client rueidis.Client) rueidis.Completed

		// IsReadOnly возвращает true если команда для чтения
		IsReadOnly() bool
	}

	Config struct {
		Masters  *cluster.Config
		Replicas *cluster.Config
		Monitor  Monitor
	}

	Core struct {
		Addresses []string
		Threshold float64
		Cluster   cluster.Cluster
	}

	Cobweb struct {
		Masters  Core
		Replicas Core
		Monitor  Monitor
	}
)

func New(config *Config) (Cobweb, error) {
	if len(config.Masters.Addresses) == 0 && len(config.Replicas.Addresses) == 0 {
		log.Fatal("incorrect system's configuration with empty nodes")
	}

	if config.Monitor == nil {
		log.Fatal("incorrect system's configuration with empty monitor")
	}

	mastersCluster, err := cluster.NewMasters(config.Masters)
	if err != nil {
		return Cobweb{}, err
	}

	replicasCluster, err := cluster.NewReplicas(config.Replicas)
	if err != nil {
		return Cobweb{}, err
	}

	return Cobweb{
		Masters: Core{
			Addresses: config.Masters.Addresses,
			Threshold: config.Masters.MaxThreshold,
			Cluster:   mastersCluster,
		},
		Replicas: Core{
			Addresses: config.Replicas.Addresses,
			Threshold: config.Replicas.MaxThreshold,
			Cluster:   replicasCluster,
		},
		Monitor: config.Monitor,
	}, nil
}

func (it *Cobweb) Execute(ctx context.Context, exec Executor) ([]rueidis.RedisResult, error) {
	snapshot := it.Monitor.Snapshot()

	var replicasCPUs, mastersCPUs []float64
	for _, addr := range it.Replicas.Addresses {
		if cpu, ok := snapshot[addr]; ok {
			replicasCPUs = append(replicasCPUs, cpu)
		}
	}
	for _, addr := range it.Masters.Addresses {
		if cpu, ok := snapshot[addr]; ok {
			mastersCPUs = append(mastersCPUs, cpu)
		}
	}

	replicasMedian := median(replicasCPUs)
	mastersMedian := median(mastersCPUs)

	var client rueidis.Client
	if replicasMedian > it.Replicas.Threshold {
		if mastersMedian <= it.Masters.Threshold {
			// Мастера в среднем свободны — читаем с них
			client = it.Masters.Cluster.Client()
		} else {
			// Обе группы "в среднем" перегружены — читаем с реплик (меньше влияние на запись)
			client = it.Replicas.Cluster.Client()
		}
	} else {
		// Реплики в среднем свободны — читаем с них
		client = it.Replicas.Cluster.Client()
	}

	return exec.Execute(ctx, client)
}
