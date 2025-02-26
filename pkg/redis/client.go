package redis

import (
	"github.com/redis/rueidis"
)

func NewRedisClient(url string) (rueidis.Client, error) {
	redisClient, err := rueidis.NewClient(rueidis.MustParseURL(url))
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}
