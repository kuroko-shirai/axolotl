package monitor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/redis/rueidis"
)

type (
	Config struct {
		Password  string
		Username  string
		Addresses []string
		Delay     time.Duration
	}

	info struct {
		user   float64
		sys    float64
		cpu    float64
		lastTs time.Time
	}

	node struct {
		client  rueidis.Client
		address string
	}

	Monitor struct {
		nodes []node
		mu    sync.RWMutex
		stats map[string]info
		delay time.Duration
	}
)

func New(config Config) (Monitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stats := make(map[string]info, len(config.Addresses))
	nodes := make([]node, 0, len(config.Addresses))
	for _, address := range config.Addresses {
		client, err := rueidis.NewClient(rueidis.ClientOption{
			Username:       config.Username,
			Password:       config.Password,
			InitAddress:    []string{address},
			SendToReplicas: func(cmd rueidis.Completed) bool { return cmd.IsReadOnly() },
			Standalone: rueidis.StandaloneOption{
				ReplicaAddress: []string{address},
			},
		})
		if err != nil {
			for _, n := range nodes {
				n.client.Close()
			}
			return Monitor{}, fmt.Errorf("failed to connect to %s: %w", address, err)
		}

		infoResp := client.Do(ctx, client.B().Info().Build())
		if err := infoResp.Error(); err != nil {
			for _, n := range nodes {
				n.client.Close()
			}
			return Monitor{}, fmt.Errorf("failed to get INFO from %s: %w", address, err)
		}

		infoStr, err := infoResp.ToString()
		if err != nil {
			client.Close()
			for _, n := range nodes {
				n.client.Close()
			}
			return Monitor{}, fmt.Errorf("failed to parse INFO from %s: %w", address, err)
		}

		current, err := parseCPUFromInfo(infoStr)
		if err != nil {
			client.Close()
			for _, n := range nodes {
				n.client.Close()
			}
			return Monitor{}, fmt.Errorf("failed to extract CPU from INFO of %s: %w", address, err)
		}

		nodes = append(nodes, node{
			client:  client,
			address: address,
		})

		stats[address] = info{
			user:   current.User,
			sys:    current.Sys,
			cpu:    -1,
			lastTs: time.Now(),
		}
	}

	return Monitor{
		nodes: nodes,
		stats: stats,
		delay: config.Delay,
	}, nil
}

func (it *Monitor) Close() {
	for _, n := range it.nodes {
		n.client.Close()
	}
}

func (it *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(it.delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			it.updateAndLogCPU()
		case <-ctx.Done():
			log.Println("Monitor stopped")
			return
		}
	}
}

// updateAndLogCPU обновляет статистику CPU и выводит её в лог.
func (it *Monitor) updateAndLogCPU() {
	if err := it.updateCPUStats(); err != nil {
		log.Printf("Partial failure during CPU update: %v", err)
	}

	it.mu.RLock()
	for address, stat := range it.stats {
		if stat.cpu < 0 {
			log.Printf("address %s: initializing...", address)
		}
	}
	it.mu.RUnlock()
}

// updateCPUStats обновляет CPU-статистику для всех узлов.
// Возвращает ошибку только если **все** узлы завершились с ошибкой.
func (it *Monitor) updateCPUStats() error {
	var (
		mu         sync.Mutex
		errs       []error
		anySuccess bool
		wg         sync.WaitGroup
	)

	for _, nd := range it.nodes {
		wg.Add(1)
		go func(n node) {
			defer wg.Done()
			if err := it.updateNodeCPU(n); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("node %s: %w", n.address, err))
				mu.Unlock()
			} else {
				mu.Lock()
				anySuccess = true
				mu.Unlock()
			}
		}(nd)
	}

	wg.Wait()

	if !anySuccess && len(errs) > 0 {
		var msg strings.Builder
		for _, e := range errs {
			msg.WriteString(e.Error() + "; ")
		}
		return fmt.Errorf("all nodes failed: %s", msg.String())
	}
	return nil
}

// updateNodeCPU обновляет статистику для одного узла.
func (it *Monitor) updateNodeCPU(n node) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp := n.client.Do(ctx, n.client.B().Info().Build())
	if err := resp.Error(); err != nil {
		return fmt.Errorf("INFO command failed: %w", err)
	}

	infoStr, err := resp.ToString()
	if err != nil {
		return fmt.Errorf("failed to convert response to string: %w", err)
	}

	current, err := parseCPUFromInfo(infoStr)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	now := time.Now()

	it.mu.Lock()
	defer it.mu.Unlock()

	prev, exists := it.stats[n.address]
	if !exists {
		return fmt.Errorf("unknown address")
	}

	if current.User < prev.user || current.Sys < prev.sys {
		it.stats[n.address] = info{
			user:   current.User,
			sys:    current.Sys,
			cpu:    -1,
			lastTs: now,
		}
		return nil
	}

	deltaTime := now.Sub(prev.lastTs).Seconds()
	if deltaTime <= 0 {
		return nil
	}

	deltaUser := current.User - prev.user
	deltaSys := current.Sys - prev.sys
	totalDelta := deltaUser + deltaSys

	usagePercent := (totalDelta / deltaTime) * 100

	it.stats[n.address] = info{
		user:   current.User,
		sys:    current.Sys,
		cpu:    usagePercent,
		lastTs: now,
	}

	return nil
}

// Snapshot возвращает копию текущей CPU-статистики.
// Значения < 0 (например, -1) исключаются.
func (it *Monitor) Snapshot() map[string]float64 {
	it.mu.RLock()
	defer it.mu.RUnlock()

	result := make(map[string]float64, len(it.stats))
	for addr, stat := range it.stats {
		if stat.cpu >= 0 {
			result[addr] = stat.cpu
		}
	}
	return result
}
