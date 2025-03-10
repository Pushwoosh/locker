package locker

import (
	"context"
	"testing"
)

func Test_example_Locker(t *testing.T) {
	t.Skip()

	t.Run("lock error", func(t *testing.T) {
		// name will be used only in metric names
		l, err := NewRedis(nil, "device-locks", nil)
		if err != nil {
			panic(err)
		}
		cbFn := func(ctx context.Context) error { return nil }

		err = l.WithLock(context.Background(), "lock-key", cbFn)
		if err != nil {
			panic(err)
		}
	})
}
