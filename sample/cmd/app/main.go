package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/kuroko-shirai/axolotl/sample/config"
	"github.com/kuroko-shirai/axolotl/sample/redis"
	"github.com/kuroko-shirai/axolotl/sample/service"
	"github.com/kuroko-shirai/axolotl/v1/cluster"
	"github.com/kuroko-shirai/axolotl/v1/cobweb"
	"github.com/kuroko-shirai/axolotl/v1/monitor"
)

const configPath = "./config.yaml"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Println(cfg)

	mastersAddresses := cfg.Masters.Addresses
	mastersMaxThreshold := cfg.Masters.MaxThreshold
	replicasAddresses := cfg.Replicas.Addresses
	replicasMaxThreshold := cfg.Replicas.MaxThreshold
	username := cfg.Username
	password := cfg.Password

	monitor, err := monitor.New(
		monitor.Config{
			Username:  username,
			Password:  password,
			Addresses: append(mastersAddresses, replicasAddresses...),
			Delay:     1 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("failed to create monitor: ", err)
	}
	defer monitor.Close()
	go monitor.Run(ctx)

	// Ждём инициализации монитора
	retryAttempts := 0
	for {
		if len(monitor.Snapshot()) == len(mastersAddresses)+len(replicasAddresses) {
			break
		}
		log.Println("monitor has not been prepared...", len(monitor.Snapshot()))
		if retryAttempts >= 5 {
			log.Fatal("failed to connect monitor to redis")
		}
		retryAttempts++
		time.Sleep(time.Second)
	}
	log.Println("monitor has been prepared...")

	// Создаём Cobweb
	cobweb, err := cobweb.New(
		&cobweb.Config{
			Masters: &cluster.Config{
				Username:     username,
				Password:     password,
				Addresses:    mastersAddresses,
				MaxThreshold: mastersMaxThreshold,
			},
			Replicas: &cluster.Config{
				Username:     username,
				Password:     password,
				Addresses:    replicasAddresses,
				MaxThreshold: replicasMaxThreshold,
			},
			Monitor: &monitor,
		},
	)
	if err != nil {
		log.Fatalf("failed to create cobweb: %v", err)
	}

	// Создаём Redis-клиент (standalone, все адреса)
	redisClient, err := redis.New(&redis.Config{
		Username:  username,
		Password:  password,
		Addresses: append(mastersAddresses, replicasAddresses...),
	})
	if err != nil {
		log.Fatalf("failed to create redis-client: %v", err)
	}
	defer redisClient.Close()

	// Создаём сервис
	svc, err := service.New(service.Config{
		Cobweb: &cobweb,
		Redis:  &redisClient,
	})
	if err != nil {
		log.Fatalf("failed to create service: %v", err)
	}

	// Тестовый вызов
	value, err := svc.GetChangePointsForShop(ctx, 5505)
	if err != nil {
		log.Fatalf("failed to get value from redis: %v", err)
	}
	log.Println("got a value from redis:", value)

	<-ctx.Done()
	log.Println("shutting down...")
}
