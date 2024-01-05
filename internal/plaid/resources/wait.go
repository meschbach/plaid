package resources

import (
	"context"
	"time"
)

type WaitPredicate func(ctx context.Context) (bool, error)

func WaitOn(parent context.Context, predicate WaitPredicate) (bool, error) {
	tracedContext, span := tracing.Start(parent, "WaitOn")
	defer span.End()

	waitUpTo := 1 * time.Second
	timeOut, timeOutDone := context.WithTimeout(tracedContext, waitUpTo)
	defer timeOutDone()

	interval := time.NewTicker(5 * time.Millisecond)
	defer interval.Stop()
	for {
		has, problem := predicate(timeOut)
		if problem != nil {
			return has, problem
		}
		if has {
			return true, problem
		}
		select {
		case <-timeOut.Done():
			return false, timeOut.Err()
		case <-interval.C:
		}
	}
}

func ForStatusState[T any](c *Client, ref Meta, status *T, check func(status *T) (bool, error)) WaitPredicate {
	return func(ctx context.Context) (bool, error) {
		if has, err := c.GetStatus(ctx, ref, status); err != nil {
			return false, err
		} else {
			if !has {
				return false, nil
			}
			return check(status)
		}
	}
}

func ForClaimedBy(c *Client, claimer Meta, types []Type, check func(ctx context.Context, match []Meta) (bool, error)) WaitPredicate {
	return func(ctx context.Context) (bool, error) {
		match, err := c.FindClaimedBy(ctx, claimer, types)
		if err != nil {
			return false, err
		}
		if len(match) == 0 {
			return false, nil
		}
		return check(ctx, match)
	}
}
