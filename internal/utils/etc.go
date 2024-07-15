package utils

import (
	"context"
	"time"
)

func GetContextByTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	} else {
		return context.WithCancel(ctx)
	}
}
