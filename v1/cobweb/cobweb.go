package cobweb

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/kuroko-shirai/axolotl/v1/cluster"
	"github.com/redis/rueidis"
)

var (
	ErrWriteCommand = errors.New("non-read command routed to cobweb")
)

type (
	// Monitor интерфейс мониторинга cpu master- и replica-нод системы.
	Monitor interface {
		Snapshot() map[string]float64 // Метод снятия текущей нагрузки системы.
	}

	Config struct {
		Masters  *cluster.Config
		Replicas *cluster.Config
		Monitor  Monitor
	}

	core struct {
		addresses []string
		threshold float64
		cluster   cluster.Cluster
	}

	Cobweb struct {
		Masters  core
		Replicas core
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

	masters, err := cluster.NewMasters(config.Masters)
	if err != nil {
		return Cobweb{}, fmt.Errorf("failed to create masters-cluster: %v", err)
	}

	replicas, err := cluster.NewReplicas(config.Replicas)
	if err != nil {
		return Cobweb{}, fmt.Errorf("failed to create replicas-cluster: %v", err)
	}

	return Cobweb{
		Masters: core{
			addresses: config.Masters.Addresses,
			threshold: config.Masters.MaxThreshold,
			cluster:   masters,
		},
		Replicas: core{
			addresses: config.Replicas.Addresses,
			threshold: config.Replicas.MaxThreshold,
			cluster:   replicas,
		},
		Monitor: config.Monitor,
	}, nil
}

func (it *Cobweb) Execute(ctx context.Context, exec Executor) ([]rueidis.RedisResult, error) {
	snapshot := it.Monitor.Snapshot()
	replicasMedian := it.median(snapshot, it.Replicas.addresses)
	mastersMedian := it.median(snapshot, it.Masters.addresses)

	var client rueidis.Client
	if replicasMedian > it.Replicas.threshold {
		if mastersMedian <= it.Masters.threshold {
			// Мастера в среднем свободны — читаем с них
			client = it.Masters.cluster.Client()
		} else {
			// Обе группы "в среднем" перегружены — читаем с реплик (меньше влияние на запись)
			client = it.Replicas.cluster.Client()
		}
	} else {
		// Реплики в среднем свободны — читаем с них
		client = it.Replicas.cluster.Client()
	}

	return exec.Execute(ctx, client)
}

func (it *Cobweb) median(snapshot map[string]float64, addresses []string) float64 {
	cpus := make([]float64, 0, len(addresses))
	for _, addr := range addresses {
		if cpu, ok := snapshot[addr]; ok {
			cpus = append(cpus, cpu)
		}
	}

	return median(cpus)
}
