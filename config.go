package locker

import (
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	LockTimeout   time.Duration `mapstructure:"lock_timeout"`
	UnlockTimeout time.Duration `mapstructure:"unlock_timeout"`
	Redis         *RedisConfig  `mapstructure:"redis"`
}

type RedisConfig struct {
	RedisConnection string `mapstructure:"redis_connection"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("empty config")
	}

	if err := c.Redis.Validate(); err != nil {
		return errors.Wrap(err, "redis")
	}

	return nil
}

func (c *RedisConfig) Validate() error {
	if c == nil {
		return errors.New("empty config")
	}

	if c.RedisConnection == "" {
		return errors.New("redis_connection is mandatory")
	}

	return nil
}
