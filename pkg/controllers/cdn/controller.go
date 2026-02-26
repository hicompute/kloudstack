package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	Client      client.Client
	RedisClient *redis.Client
}

func New() *Controller {
	client, err := k8s.NewClient()
	if err != nil {
		klog.Fatalf("Error on creating kubernetes client : %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", viper.GetString("REDIS_HOST"), viper.GetInt32("REDIS_PORT")),
		Password:     viper.GetString("REDIS_PASSWORD"),
		DB:           viper.GetInt("REDIS_DB"),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		klog.Fatalf("Error on creating redis client : %v", err)
	}

	return &Controller{
		Client:      client,
		RedisClient: redisClient,
	}
}
