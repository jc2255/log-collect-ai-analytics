package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	once   sync.Once
)

// Init 初始化Redis连接
func Init(addr, password string, db int) error {
	var err error
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})
		ctx := context.Background()
		if e := client.Ping(ctx).Err(); e != nil {
			err = fmt.Errorf("redis connect failed: %w", e)
			return
		}
	})
	return err
}

// GetClient 获取Redis客户端
func GetClient() *redis.Client {
	return client
}

// Close 关闭连接
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}
