package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	pool          *redis.Client
	redisHost     = "localhost:6379"
	redisPassword = "Lollzp1999!"
)

// NewRedisPool: create a new Redis connection pool
func NewRedisPool() (*redis.Client, error) {
	// Create a new Redis client
	pool := redis.NewClient(&redis.Options{
		Addr:         redisHost,
		Password:     redisPassword,
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		IdleTimeout:  300 * time.Second,
	})

	// Ping the Redis server to check if the connection is successful
	_, err := pool.Ping(pool.Context()).Result()
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func init() {
	// Initialize a new Redis connection pool
	pool, _ = NewRedisPool()
}

func RedisPool() *redis.Client {
	return pool
}
