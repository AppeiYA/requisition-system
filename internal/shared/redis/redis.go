package redis

import "github.com/redis/go-redis/v9"

type Redis struct {
	client *redis.Client
}

func New(client *redis.Client) *Redis {
	return &Redis{
		client: client,
	}
}

func (r *Redis) Client() *redis.Client {
	return r.client
}