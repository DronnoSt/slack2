package network

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"runtime/trace"
	"time"

	"errors"

	"github.com/slack-go/slack"
	"golang.org/x/time/rate"

	"github.com/rusq/slackdump/v2/logger"
)

// defNumAttempts is the default number of retry attempts.
const defNumAttempts = 3

// Logger is the package logger.
var Logger logger.Interface = logger.Default

// ErrRetryFailed is returned if number of retry attempts exceeded the retry attempts limit and
// function wasn't able to complete without errors.
var ErrRetryFailed = errors.New("callback was not able to complete without errors within the allowed number of retries")

// withRetry will run the callback function fn. If the function returns
// slack.RateLimitedError, it will delay, and then call it again up to
// maxAttempts times. It will return an error if it runs out of attempts.
func WithRetry(ctx context.Context, l *rate.Limiter, maxAttempts int, fn func() error) error {
	var ok bool
	if maxAttempts == 0 {
		maxAttempts = defNumAttempts
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var err error
		trace.WithRegion(ctx, "withRetry.wait", func() {
			err = l.Wait(ctx)
		})
		if err != nil {
			return err
		}

		cbErr := fn()
		if cbErr == nil {
			ok = true
			break
		}

		tracelogf(ctx, "error", "%s: WithRetry: %s after %d attempts", caller(1), cbErr, attempt+1)
		var rateErr *slack.RateLimitedError
		if !errors.As(cbErr, &rateErr) {
			return fmt.Errorf("%s: WithRetry: callback error: %w", caller(1), cbErr)
		}

		tracelogf(ctx, "info", "got rate limited, sleeping %s", rateErr.RetryAfter)
		time.Sleep(rateErr.RetryAfter)
	}
	if !ok {
		return ErrRetryFailed
	}
	return nil
}

func tracelogf(ctx context.Context, category string, fmt string, a ...any) {
	trace.Logf(ctx, category, fmt, a...)
	l().Debugf(fmt, a...)
}

func l() logger.Interface {
	if Logger == nil {
		return logger.Default
	}
	return Logger
}

func caller(steps int) string {
	name := "?"
	if pc, _, _, ok := runtime.Caller(steps + 1); ok {
		name = filepath.Base(runtime.FuncForPC(pc).Name())
	}
	return name
}
