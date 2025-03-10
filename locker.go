package locker

import (
	"context"
	"time"

	infralog "github.com/pushwoosh/infra/log"
	infraredis "github.com/pushwoosh/infra/redis"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Locker struct {
	cfg     *Config
	metrics *metrics
	backend locker
}

type locker interface {
	// Lock acquires a lock.
	// The function blocks either until the lock is successfully acquired or until the context expires.
	// Lock must be released automatically after `expire`.
	Lock(ctx context.Context, key string, expiration time.Duration) error

	// Unlock releases acquired lock
	// Unlock should create its own context to protect from the situation when locked
	// code is timed out, and we have no time left to release lock.
	Unlock(key string, expiration time.Duration) error
}

func NewRedis(cfg *Config, name string, redises *infraredis.Container) (*Locker, error) {
	if cfg.Redis == nil {
		return nil, errors.New("empty redis config")
	}

	mtr := newMetrics(name)

	l, err := newRedisLocker(cfg.Redis, redises, mtr)
	if err != nil {
		return nil, errors.Wrap(err, "redis locker")
	}

	return &Locker{
		cfg:     cfg,
		metrics: mtr,
		backend: l,
	}, nil
}

func (l *Locker) WithLock(ctx context.Context, key string, f func(context.Context) error) error {
	if err := l.backend.Lock(ctx, key, l.cfg.LockTimeout); err != nil {
		return err
	}
	l.metrics.locksAcquired.Inc()
	l.metrics.locksHeld.Inc()

	defer func() {
		defer l.metrics.locksReleased.Inc()
		defer l.metrics.locksHeld.Dec()
		if err := l.backend.Unlock(key, l.cfg.UnlockTimeout); err != nil {
			infralog.Error("failed to release lock", zap.String("key", key), zap.Error(err))
		}
	}()
	return f(ctx)
}
