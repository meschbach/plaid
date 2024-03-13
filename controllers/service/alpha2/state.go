package alpha2

import (
	"context"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/controllers/tooling/kit"
)

type State struct {
	next     *tokenState
	stable   *tokenState
	stopping []*tokenState
	old      []*tokenState
	bridge   kit.Manager
}

func (s *State) progress(ctx context.Context, env tooling.Env) error {
	//todo: probably best to accumulate errors
	if s.next != nil {
		if err := s.next.progressBuild(ctx, env, s); err != nil {
			return err
		}
	}
	nextStopping := make([]*tokenState, 0, len(s.stopping))
	for _, stopping := range s.stopping {
		if move, err := stopping.progressStopping(ctx, env); err != nil {
			return err
		} else if move {
			s.old = append(s.old, stopping)
		} else {
			nextStopping = append(nextStopping, stopping)
		}
	}
	s.stopping = nextStopping

	if s.stable != nil {
		if err := s.stable.progressRun(ctx, env, s); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) promoteNext() {
	if s.stable != nil {
		s.stopping = append(s.stopping, s.stable)
	}
	s.stable = s.next
	s.next = nil
}
