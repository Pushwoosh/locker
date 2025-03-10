package locker

import (
	"context"
	"math"
	"time"

	"github.com/pkg/errors"
	infraredis "github.com/pushwoosh/infra/redis"
	"github.com/redis/go-redis/v9"
)

type redisLocker struct {
	conn    redis.UniversalClient
	metrics *metrics
}

func newRedisLocker(cfg *RedisConfig, redises *infraredis.Container, metrics *metrics) (*redisLocker, error) {
	conn := redises.Get(cfg.RedisConnection)
	if conn == nil {
		return nil, errors.Errorf("connection \"%s\" does not exist", cfg.RedisConnection)
	}

	return &redisLocker{
		conn:    conn,
		metrics: metrics,
	}, nil
}

func (r *redisLocker) Lock(ctx context.Context, key string, expiration time.Duration) error {
	const (
		sleepBase  = time.Millisecond * 50 // initial sleep time
		sleepCap   = time.Second * 5       // max sleep time
		multiplier = 1.6487212707          // sqrt(e). just for no reason
	)

	// fast path: lock and return
	ok, err := r.conn.SetNX(ctx, key, "", expiration).Result()
	if err != nil {
		return errors.Wrap(err, "unexpected redis error")
	}

	if ok {
		return nil
	}

	// slow path: sleep and retry until context expires
	for attempt := 0.0; ; attempt++ {
		// math.Min is used to prevent going to float's infinity
		sleepTime := time.Duration(float64(sleepBase) * math.Pow(multiplier, math.Min(attempt, 1000)))
		if sleepTime > sleepCap {
			sleepTime = sleepCap
		}

		time.Sleep(sleepTime)

		r.metrics.lockWait.Add(float64(sleepTime) / float64(time.Second))

		// value is not used. set it to sleeping time just for debugging purposes
		value := sleepTime.String()
		set, setErr := r.conn.SetNX(ctx, key, value, expiration).Result()
		if setErr != nil {
			return errors.Wrap(setErr, "unexpected redis error")
		}

		if set {
			return nil
		}
	}
}

func (r *redisLocker) Unlock(key string, expiration time.Duration) error {
	if expiration == 0 {
		expiration = time.Second * 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), expiration)
	defer cancel()

	_, err := r.conn.Del(ctx, key).Result()
	return errors.Wrap(err, "unexpected redis error")
}
